package service

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-runner/internal/types"

	"gorm.io/gorm"
)

// Manager handles service lifecycle management
type Manager struct {
	db       *gorm.DB
	processes map[uint]*ProcessInfo
	mu       sync.RWMutex
}

// ProcessInfo holds information about a running process
type ProcessInfo struct {
	ProjectID uint
	Process   *exec.Cmd
	Context   context.Context
	Cancel    context.CancelFunc
	StartTime time.Time
	Logs      chan string
	LogBuffer []string // Buffer to store recent logs (last 1000 lines)
	logMu     sync.Mutex
	closed    bool     // Track if channel is closed
	closeMu   sync.Mutex
}

// NewManager creates a new service manager
func NewManager(db *gorm.DB) *Manager {
	return &Manager{
		db:        db,
		processes: make(map[uint]*ProcessInfo),
	}
}

// StartService starts a microservice
func (m *Manager) StartService(projectID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if service is already running
	if _, exists := m.processes[projectID]; exists {
		return fmt.Errorf("service %d is already running", projectID)
	}

	// Get project from database
	var p struct {
		ID          uint           `gorm:"primarykey"`
		Name        string
		Description string
		Type        string
		Path        string
		Command     string
		Args        string
		WorkingDir  string
		Port        int
		Environment string
		EnvFile     string
		EnvVars     string
		Status      string
		PID         int
		StartTime   *time.Time
		StopTime    *time.Time
		LastError   string
	}

	if err := m.db.Table("projects").Where("id = ?", projectID).First(&p).Error; err != nil {
		return fmt.Errorf("project not found: %v", err)
	}

	// Update status to starting
	m.db.Table("projects").Where("id = ?", projectID).Update("status", string(types.StatusStarting))

	// Create context for the process
	ctx, cancel := context.WithCancel(context.Background())

	// Prepare command
	cmd := m.prepareCommand(ctx, &struct {
		Command string
		Args    string
		Type    string
	}{
		Command: p.Command,
		Args:    p.Args,
		Type:    p.Type,
	})

	// Set working directory
	if p.WorkingDir != "" {
		cmd.Dir = p.WorkingDir
	} else {
		cmd.Dir = p.Path
	}

	// Set environment variables
	cmd.Env = m.prepareEnvironment(&struct {
		Port        int
		Environment string
		EnvFile     string
		EnvVars     string
		Path        string
	}{
		Port:        p.Port,
		Environment: p.Environment,
		EnvFile:     p.EnvFile,
		EnvVars:     p.EnvVars,
		Path:        p.Path,
	})

	// Create logs channel with larger buffer to avoid dropping logs
	logs := make(chan string, 1000)

	// Capture stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %v", err)
	}

	// Store process info
	processInfo := &ProcessInfo{
		ProjectID: projectID,
		Process:   cmd,
		Context:   ctx,
		Cancel:    cancel,
		StartTime: time.Now(),
		Logs:      logs,
		LogBuffer: make([]string, 0, 1000), // Buffer for last 1000 lines
	}
	m.processes[projectID] = processInfo

	// Start the process
	if err := cmd.Start(); err != nil {
		delete(m.processes, projectID)
		cancel()
		close(logs)
		m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
			"status":      string(types.StatusError),
			"last_error":  err.Error(),
		})
		return fmt.Errorf("failed to start service: %v", err)
	}

	// Get PID immediately after start
	pid := cmd.Process.Pid
	if pid <= 0 {
		delete(m.processes, projectID)
		cancel()
		close(logs)
		errorMsg := "Process started but PID is invalid."
		m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
			"status":      string(types.StatusError),
			"last_error":  errorMsg,
			"p_id":        0,
		})
		return fmt.Errorf("process started but PID is invalid")
	}

	// Start goroutines to capture stdout and stderr immediately
	// This allows us to see any errors that occur during startup
	// Use captureOutputWithBuffer to also store logs in buffer
	go m.captureOutputWithBuffer(stdout, processInfo, false)
	go m.captureOutputWithBuffer(stderr, processInfo, true)

	// Start monitoring goroutine immediately
	// This will detect when process exits and update status accordingly
	// If process exits quickly, monitorProcess will catch it
	go m.monitorProcess(processInfo)

	// Update project with PID and start time
	// We assume process started successfully if we got a valid PID
	// monitorProcess will update status to "error" or "stopped" if process exits
	now := time.Now()
	m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
		"status":      string(types.StatusRunning),
		"p_id":        pid,
		"start_time":  &now,
		"last_error":  "",
	})

	return nil
}

// StopService stops a microservice
func (m *Manager) StopService(projectID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// First check actual process status from database
	var p struct {
		ID     uint
		PID    int `gorm:"column:p_id"`
		Status string
	}
	if err := m.db.Table("projects").Where("id = ?", projectID).First(&p).Error; err != nil {
		return fmt.Errorf("project not found: %v", err)
	}

	// Check if process is actually running by checking PID
	processRunning := false
	if p.PID > 0 {
		proc, err := os.FindProcess(p.PID)
		if err == nil {
			// Try to signal the process (signal 0 doesn't kill, just checks existence)
			err = proc.Signal(os.Signal(nil))
			if err == nil {
				processRunning = true
			}
		}
	}

	// Check if we have process info in memory
	processInfo, exists := m.processes[projectID]
	
	// If process is not running (neither in memory nor by PID), update DB and return
	if !processRunning && !exists {
		// Process is already stopped, just update DB
		now := time.Now()
		m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
			"status":     string(types.StatusStopped),
			"stop_time":  &now,
			"p_id":       0,
		})
		return nil // Not an error, process is already stopped
	}

	// Update status to stopping
	m.db.Table("projects").Where("id = ?", projectID).Update("status", string(types.StatusStopping))

	// If we have process info in memory, stop it properly
	if exists {
		// Cancel context to stop the process
		processInfo.Cancel()

		// Wait for process to finish or timeout (reduced to 10 seconds)
		done := make(chan error, 1)
		go func() {
			done <- processInfo.Process.Wait()
		}()

		select {
		case <-done:
			// Process finished normally
		case <-time.After(10 * time.Second):
			// Force kill if timeout
			if processInfo.Process.Process != nil {
				processInfo.Process.Process.Kill()
				// Wait a bit more
				select {
				case <-done:
					// Process killed successfully
				case <-time.After(2 * time.Second):
					// Still not dead, force kill again
					if processInfo.Process.Process != nil {
						processInfo.Process.Process.Kill()
					}
				}
			}
		}

		// Clean up
		processInfo.safeCloseChannel()
		delete(m.processes, projectID)
	} else if processRunning && p.PID > 0 {
		// Process is running but not in our map (maybe server restarted)
		// Try to kill it by PID
		proc, err := os.FindProcess(p.PID)
		if err == nil {
			proc.Kill()
			// Wait a bit for process to die
			time.Sleep(1 * time.Second)
			// Try again if still alive
			if err := proc.Signal(os.Signal(nil)); err == nil {
				proc.Kill()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	// Update project status
	now := time.Now()
	m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
		"status":     string(types.StatusStopped),
		"stop_time":  &now,
		"p_id":       0,
	})

	return nil
}

// ForceKillService forcefully kills a service process
func (m *Manager) ForceKillService(projectID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Get PID from database
	var p struct {
		ID  uint
		PID int `gorm:"column:p_id"`
	}
	if err := m.db.Table("projects").Where("id = ?", projectID).First(&p).Error; err != nil {
		return fmt.Errorf("project not found: %v", err)
	}

	// Kill process by PID if exists
	if p.PID > 0 {
		proc, err := os.FindProcess(p.PID)
		if err == nil {
			// Try graceful kill first
			proc.Kill()
			time.Sleep(500 * time.Millisecond)
			// Check if still alive and force kill
			if err := proc.Signal(os.Signal(nil)); err == nil {
				proc.Kill()
				time.Sleep(500 * time.Millisecond)
			}
		}
	}

	// Clean up from memory if exists
	if processInfo, exists := m.processes[projectID]; exists {
		processInfo.Cancel()
		processInfo.safeCloseChannel()
		delete(m.processes, projectID)
	}

	// Update database
	now := time.Now()
	m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
		"status":     string(types.StatusStopped),
		"stop_time":  &now,
		"p_id":       0,
		"last_error": "Force killed",
	})

	return nil
}

// RestartService restarts a microservice
func (m *Manager) RestartService(projectID uint) error {
	// Stop if running
	if _, exists := m.processes[projectID]; exists {
		if err := m.StopService(projectID); err != nil {
			return fmt.Errorf("failed to stop service: %v", err)
		}
		// Wait a bit before restarting
		time.Sleep(2 * time.Second)
	}

	// Start the service
	return m.StartService(projectID)
}

// GetServiceStatus returns the current status of a service
func (m *Manager) GetServiceStatus(projectID uint) (map[string]interface{}, error) {
	if projectID == 0 {
		return nil, fmt.Errorf("invalid project ID: %d", projectID)
	}
	
	m.mu.RLock()
	
	// Query into struct first, then convert to map
	var p struct {
		ID          uint           `gorm:"column:id"`
		Name        string         `gorm:"column:name"`
		Description string         `gorm:"column:description"`
		Type        string         `gorm:"column:type"`
		Path        string         `gorm:"column:path"`
		Command     string         `gorm:"column:command"`
		Args        string         `gorm:"column:args"`
		WorkingDir  string         `gorm:"column:working_dir"`
		Port        int            `gorm:"column:port"`
		Ports       string         `gorm:"column:ports"`
		Environment string         `gorm:"column:environment"`
		EnvFile     string         `gorm:"column:env_file"`
		EnvVars     string         `gorm:"column:env_vars"`
		Editor      string         `gorm:"column:editor"`
		EditorArgs  string         `gorm:"column:editor_args"`
		Status      string         `gorm:"column:status"`
		PID         int            `gorm:"column:p_id"`
		StartTime   *time.Time     `gorm:"column:start_time"`
		StopTime    *time.Time     `gorm:"column:stop_time"`
		LastError   string         `gorm:"column:last_error"`
		HealthCheckURL string       `gorm:"column:health_check_url"`
		HealthStatus   string       `gorm:"column:health_status"`
		AutoRestart   bool         `gorm:"column:auto_restart"`
		MaxRestarts   int          `gorm:"column:max_restarts"`
		CPULimit      string       `gorm:"column:cpu_limit"`
		MemoryLimit   string       `gorm:"column:memory_limit"`
		GroupID       *uint        `gorm:"column:group_id"`
		CreatedAt     time.Time    `gorm:"column:created_at"`
		UpdatedAt     time.Time    `gorm:"column:updated_at"`
		Logs          string       `gorm:"column:logs"`
	}
	
	if err := m.db.Table("projects").Where("id = ?", projectID).First(&p).Error; err != nil {
		m.mu.RUnlock()
		return nil, err
	}
	
	// Convert struct to map
	result := map[string]interface{}{
		"id":            p.ID,
		"name":          p.Name,
		"description":   p.Description,
		"type":          p.Type,
		"path":          p.Path,
		"command":       p.Command,
		"args":          p.Args,
		"working_dir":   p.WorkingDir,
		"port":          p.Port,
		"ports":         p.Ports,
		"environment":   p.Environment,
		"env_file":      p.EnvFile,
		"env_vars":      p.EnvVars,
		"editor":        p.Editor,
		"editor_args":   p.EditorArgs,
		"status":        p.Status,
		"p_id":          p.PID,
		"start_time":    p.StartTime,
		"stop_time":     p.StopTime,
		"last_error":    p.LastError,
		"health_check_url": p.HealthCheckURL,
		"health_status":    p.HealthStatus,
		"auto_restart":     p.AutoRestart,
		"max_restarts":     p.MaxRestarts,
		"cpu_limit":        p.CPULimit,
		"memory_limit":     p.MemoryLimit,
		"group_id":         p.GroupID,
		"created_at":       p.CreatedAt,
		"updated_at":       p.UpdatedAt,
		"logs":             p.Logs,
	}

	// Use IsServiceRunning to check actual status (checks port, PID, child processes)
	// This is more reliable than just checking PID
	isRunning := m.IsServiceRunning(projectID)
	currentStatus := p.Status
	pid := p.PID
	
	// Check if we have process info in memory
	processInfo, exists := m.processes[projectID]
	
	m.mu.RUnlock() // Release read lock before checking/updating
	
	// If service is actually running but status says it's stopped/starting, update it
	if isRunning && (currentStatus == string(types.StatusStopped) || currentStatus == string(types.StatusStarting)) {
		m.mu.Lock()
		now := time.Now()
		// Try to get actual PID from port if PID is 0
		actualPID := pid
		if actualPID == 0 {
			// Get PID from port
			var project struct {
				Port int
			}
			if err := m.db.Table("projects").Where("id = ?", projectID).Select("port").First(&project).Error; err == nil && project.Port > 0 {
				if pidFromPort, err := m.getPIDByPort(project.Port); err == nil {
					actualPID = pidFromPort
				}
			}
		}
		
		m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
			"status":     string(types.StatusRunning),
			"p_id":       actualPID,
			"start_time": &now,
		})
		result["status"] = string(types.StatusRunning)
		result["p_id"] = actualPID
		result["start_time"] = &now
		m.mu.Unlock()
	} else if !isRunning {
		// Service is not running
		// Check if process in memory has exited
		if exists {
			m.mu.Lock()
			if processInfo.Process.ProcessState != nil && processInfo.Process.ProcessState.Exited() {
				// Process has exited, update status
				now := time.Now()
				m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
					"status":     string(types.StatusStopped),
					"stop_time":  &now,
					"p_id":       0,
				})
				processInfo.safeCloseChannel()
				delete(m.processes, projectID)
				result["status"] = string(types.StatusStopped)
				result["p_id"] = 0
				result["stop_time"] = &now
			} else {
				// Check if process is really dead
				if processInfo.Process != nil && processInfo.Process.Process != nil {
					err := processInfo.Process.Process.Signal(os.Signal(nil))
					if err != nil {
						// Process is dead
						now := time.Now()
						m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
							"status":     string(types.StatusStopped),
							"stop_time":  &now,
							"p_id":       0,
						})
						processInfo.safeCloseChannel()
						delete(m.processes, projectID)
						result["status"] = string(types.StatusStopped)
						result["p_id"] = 0
						result["stop_time"] = &now
					}
				}
			}
			m.mu.Unlock()
		} else if currentStatus == string(types.StatusRunning) || currentStatus == string(types.StatusStarting) {
			// Process not in memory and not running, but status says it is
			m.mu.Lock()
			now := time.Now()
			m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
				"status":     string(types.StatusStopped),
				"stop_time":  &now,
				"p_id":       0,
			})
			result["status"] = string(types.StatusStopped)
			result["p_id"] = 0
			result["stop_time"] = &now
			m.mu.Unlock()
		}
	}

	return result, nil
}

// GetRunningServices returns all currently running services
func (m *Manager) GetRunningServices() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var services []map[string]interface{}
	for _, processInfo := range m.processes {
		var p map[string]interface{}
		if err := m.db.Table("projects").Where("id = ?", processInfo.ProjectID).First(&p).Error; err == nil {
			services = append(services, p)
		}
	}
	return services
}

// GetServiceLogs returns logs for a service
// It returns the logs channel if service is running, or nil if not
func (m *Manager) GetServiceLogs(projectID uint) <-chan string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if processInfo, exists := m.processes[projectID]; exists {
		// Double-check that process is actually running
		if processInfo.Process != nil && processInfo.Process.Process != nil {
			// Check if process is still alive
			err := processInfo.Process.Process.Signal(os.Signal(nil))
			if err == nil {
				return processInfo.Logs
			}
			// Process is dead, but we still have it in memory
			// This will be cleaned up by GetServiceStatus
		}
	}
	return nil
}

// IsServiceRunning checks if a service is actually running (in memory or by PID)
// It also checks for child processes and port availability
func (m *Manager) IsServiceRunning(projectID uint) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check if in memory
	if processInfo, exists := m.processes[projectID]; exists {
		if processInfo.Process != nil && processInfo.Process.Process != nil {
			// Check if process is still alive
			err := processInfo.Process.Process.Signal(os.Signal(nil))
			if err == nil {
				return true
			}
		}
	}

	// Get project info including PID and Port
	var project struct {
		PID  int
		Port int
		Path string
	}
	if err := m.db.Table("projects").Where("id = ?", projectID).Select("p_id, port, path").First(&project).Error; err != nil {
		return false
	}

	// Priority 1: Check by port if port is specified
	// This is the most reliable indicator for services like vite, node, etc.
	// Parent process (npm/yarn) may exit, but child process (vite) is still running on port
	if project.Port > 0 {
		if m.isPortInUse(project.Port) {
			return true
		}
	}

	// Priority 2: Check by PID from database (parent process)
	if project.PID > 0 {
		proc, err := os.FindProcess(project.PID)
		if err == nil {
			err = proc.Signal(os.Signal(nil))
			if err == nil {
				return true
			}
		}
		
		// Parent process not found, check for child processes
		// This handles cases where parent (npm/yarn) exits but child (vite/node) is still running
		if m.hasChildProcesses(project.PID) {
			return true
		}
	}

	return false
}

// hasChildProcesses checks if a process has any child processes still running
func (m *Manager) hasChildProcesses(parentPID int) bool {
	// Use pgrep to find child processes (works on macOS and Linux)
	// This finds all processes whose parent PID matches the given PID
	cmd := exec.Command("pgrep", "-P", strconv.Itoa(parentPID))
	output, err := cmd.Output()
	if err == nil && len(output) > 0 {
		// Found child processes - output contains PIDs of child processes
		return true
	}
	
	return false
}

// isPortInUse checks if a port is currently in use
func (m *Manager) isPortInUse(port int) bool {
	// Use lsof to check if port is in use (works on macOS and Linux)
	// -i :port filters by port
	// -sTCP:LISTEN only shows listening ports (not established connections)
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-sTCP:LISTEN")
	err := cmd.Run()
	if err == nil {
		// Port is in use (lsof found processes listening on this port)
		return true
	}
	
	// Fallback: try lsof without TCP filter (for broader compatibility)
	cmd = exec.Command("lsof", "-i", fmt.Sprintf(":%d", port))
	if err := cmd.Run(); err == nil {
		return true
	}
	
	// Fallback: try netstat (Linux)
	cmd = exec.Command("netstat", "-an")
	if output, err := cmd.Output(); err == nil {
		portStr := fmt.Sprintf(":%d", port)
		// Look for LISTEN state on this port
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, portStr) && strings.Contains(line, "LISTEN") {
				return true
			}
		}
	}
	
	// Fallback: try ss command (modern alternative to netstat on Linux)
	cmd = exec.Command("ss", "-an")
	if output, err := cmd.Output(); err == nil {
		portStr := fmt.Sprintf(":%d", port)
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, portStr) && strings.Contains(line, "LISTEN") {
				return true
			}
		}
	}
	
	return false
}

// PortInfo represents information about a port in use
type PortInfo struct {
	Port        int    `json:"port"`
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	User        string `json:"user"`
	Command     string `json:"command"`
	Status      string `json:"status"`
}

// GetPortsInUse returns list of all ports in use with process information
func (m *Manager) GetPortsInUse() ([]PortInfo, error) {
	var ports []PortInfo

	// Use lsof to get all listening ports
	// -iTCP -sTCP:LISTEN -P -n: TCP LISTEN state, show port numbers, no hostname resolution
	// macOS uses -iTCP instead of -i 4-6
	cmd := exec.Command("lsof", "-iTCP", "-sTCP:LISTEN", "-P", "-n")
	output, err := cmd.Output()
	if err != nil {
		// If lsof fails, try netstat as fallback
		return m.getPortsFromNetstat()
	}

	// Parse lsof output
	// Format: COMMAND   PID USER   FD   TYPE DEVICE SIZE/OFF NODE NAME
	// Example: node    52071 ducnm   24u  IPv6 0x...      0t0  TCP [::1]:5173 (LISTEN)
	// Example: node    52071 ducnm   24u  IPv6 0x...      0t0  TCP *:5173 (LISTEN)
	lines := strings.Split(string(output), "\n")
	for i, line := range lines {
		if i == 0 || strings.TrimSpace(line) == "" {
			continue // Skip header and empty lines
		}

		fields := strings.Fields(line)
		if len(fields) < 8 {
			continue
		}

		// Parse fields - minimum required: COMMAND PID USER FD TYPE ... NAME
		command := fields[0]
		pidStr := fields[1]
		user := fields[2]
		
		// Get PID
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			continue
		}

		// Get port from NAME field (last field contains address:port)
		// The NAME field is typically the last field, but we need to handle cases where
		// it might contain spaces (e.g., "TCP *:5173 (LISTEN)")
		// Look for the field containing ":" which indicates an address:port
		var nameField string
		for j := len(fields) - 1; j >= 0; j-- {
			if strings.Contains(fields[j], ":") {
				// Found the address:port field, but it might be split or combined
				// Reconstruct from this field to the end
				nameField = strings.Join(fields[j:], " ")
				break
			}
		}
		
		if nameField == "" {
			continue
		}
		
		port := extractPortFromLsofName(nameField)
		if port == 0 {
			continue
		}

		// Get full command line
		fullCommand := m.getProcessCommand(pid)

		ports = append(ports, PortInfo{
			Port:        port,
			PID:         pid,
			ProcessName: command,
			User:        user,
			Command:     fullCommand,
			Status:      "LISTEN",
		})
	}

	return ports, nil
}

// getPortsFromNetstat is a fallback method to get ports using netstat
func (m *Manager) getPortsFromNetstat() ([]PortInfo, error) {
	var ports []PortInfo

	cmd := exec.Command("netstat", "-tlnp")
	output, err := cmd.Output()
	if err != nil {
		// Try ss command as another fallback
		return m.getPortsFromSS()
	}

	// Parse netstat output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LISTEN") {
			fields := strings.Fields(line)
			if len(fields) < 4 {
				continue
			}

			// Extract port from local address (format: 0.0.0.0:5173 or :::5173)
			localAddr := fields[3]
			port := extractPortFromAddr(localAddr)
			if port == 0 {
				continue
			}

			// Try to extract PID if available (Linux format: 1234/process)
			pid := 0
			if len(fields) >= 7 {
				pidStr := strings.Split(fields[6], "/")[0]
				if p, err := strconv.Atoi(pidStr); err == nil {
					pid = p
				}
			}

			if pid > 0 {
				fullCommand := m.getProcessCommand(pid)
				processName := "unknown"
				if parts := strings.Split(fields[6], "/"); len(parts) > 1 {
					processName = parts[1]
				}

				ports = append(ports, PortInfo{
					Port:        port,
					PID:         pid,
					ProcessName: processName,
					User:        "unknown",
					Command:     fullCommand,
					Status:      "LISTEN",
				})
			}
		}
	}

	return ports, nil
}

// getPortsFromSS is a fallback method to get ports using ss command
func (m *Manager) getPortsFromSS() ([]PortInfo, error) {
	var ports []PortInfo

	cmd := exec.Command("ss", "-tlnp")
	output, err := cmd.Output()
	if err != nil {
		return ports, fmt.Errorf("failed to get ports: %v", err)
	}

	// Parse ss output (similar to netstat)
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "LISTEN") {
			fields := strings.Fields(line)
			if len(fields) < 5 {
				continue
			}

			// Extract port from local address
			localAddr := fields[3]
			port := extractPortFromAddr(localAddr)
			if port == 0 {
				continue
			}

			// Extract PID from process field (format: pid=1234,cmd=node)
			pid := 0
			processName := "unknown"
			if len(fields) >= 6 {
				processField := fields[5]
				if pidMatch := extractPIDFromSS(processField); pidMatch > 0 {
					pid = pidMatch
					processName = extractProcessNameFromSS(processField)
				}
			}

			if pid > 0 {
				fullCommand := m.getProcessCommand(pid)
				ports = append(ports, PortInfo{
					Port:        port,
					PID:         pid,
					ProcessName: processName,
					User:        "unknown",
					Command:     fullCommand,
					Status:      "LISTEN",
				})
			}
		}
	}

	return ports, nil
}

// KillPort kills the process using the specified port
func (m *Manager) KillPort(port int) error {
	// Get PID of process using the port
	pid, err := m.getPIDByPort(port)
	if err != nil {
		return fmt.Errorf("failed to find process on port %d: %v", port, err)
	}

	if pid == 0 {
		return fmt.Errorf("no process found on port %d", port)
	}

	// Kill the process
	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process %d: %v", pid, err)
	}

	// Try graceful kill first
	err = proc.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill process %d: %v", pid, err)
	}

	// Wait a bit and check if process is still alive
	time.Sleep(500 * time.Millisecond)
	if err := proc.Signal(os.Signal(nil)); err == nil {
		// Process still alive, force kill
		proc.Kill()
	}

	return nil
}

// Helper functions

// extractPortFromLsofName extracts port from lsof NAME field
// Format: *:5173, localhost:5173, 127.0.0.1:5173, [::1]:5173
func extractPortFromLsofName(name string) int {
	// Remove IPv6 brackets
	name = strings.Trim(name, "[]")
	
	// Find last colon
	lastColon := strings.LastIndex(name, ":")
	if lastColon == -1 {
		return 0
	}
	
	portStr := name[lastColon+1:]
	// Remove (LISTEN) or other suffixes
	if parenIdx := strings.Index(portStr, "("); parenIdx != -1 {
		portStr = portStr[:parenIdx]
	}
	
	port, err := strconv.Atoi(strings.TrimSpace(portStr))
	if err != nil || port <= 0 || port > 65535 {
		return 0
	}
	
	return port
}

// extractPortFromAddr extracts port from address string
// Format: 0.0.0.0:5173, :::5173, *:5173
func extractPortFromAddr(addr string) int {
	lastColon := strings.LastIndex(addr, ":")
	if lastColon == -1 {
		return 0
	}
	
	portStr := addr[lastColon+1:]
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		return 0
	}
	
	return port
}

// extractPIDFromSS extracts PID from ss process field
// Format: pid=1234,cmd=node
func extractPIDFromSS(processField string) int {
	// Look for pid=1234 pattern
	if strings.HasPrefix(processField, "pid=") {
		parts := strings.Split(processField, ",")
		if len(parts) > 0 {
			pidStr := strings.TrimPrefix(parts[0], "pid=")
			if pid, err := strconv.Atoi(pidStr); err == nil {
				return pid
			}
		}
	}
	return 0
}

// extractProcessNameFromSS extracts process name from ss process field
func extractProcessNameFromSS(processField string) string {
	// Look for cmd=node pattern
	if idx := strings.Index(processField, "cmd="); idx != -1 {
		cmdPart := processField[idx+4:]
		if commaIdx := strings.Index(cmdPart, ","); commaIdx != -1 {
			return cmdPart[:commaIdx]
		}
		return cmdPart
	}
	return "unknown"
}

// getPIDByPort gets the PID of process using the specified port
func (m *Manager) getPIDByPort(port int) (int, error) {
	// Use lsof to find PID
	cmd := exec.Command("lsof", "-ti", fmt.Sprintf(":%d", port))
	output, err := cmd.Output()
	if err == nil {
		pidStr := strings.TrimSpace(string(output))
		if pid, err := strconv.Atoi(pidStr); err == nil {
			return pid, nil
		}
	}

	// Fallback: use netstat
	cmd = exec.Command("netstat", "-tlnp")
	output, err = cmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, fmt.Sprintf(":%d", port)) && strings.Contains(line, "LISTEN") {
				fields := strings.Fields(line)
				if len(fields) >= 7 {
					pidStr := strings.Split(fields[6], "/")[0]
					if pid, err := strconv.Atoi(pidStr); err == nil {
						return pid, nil
					}
				}
			}
		}
	}

	return 0, fmt.Errorf("process not found on port %d", port)
}

// getProcessCommand gets the full command line for a process
func (m *Manager) getProcessCommand(pid int) string {
	// Try ps command
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "command=")
	output, err := cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	// Fallback: try ps with different format
	cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=")
	output, err = cmd.Output()
	if err == nil {
		return strings.TrimSpace(string(output))
	}

	return "unknown"
}

// GetServiceLogBuffer returns the buffered logs for a service
// This works even if the service is not currently running
func (m *Manager) GetServiceLogBuffer(projectID uint) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if processInfo, exists := m.processes[projectID]; exists {
		return processInfo.getLogBuffer()
	}
	return []string{}
}

// prepareCommand creates the command to execute
func (m *Manager) prepareCommand(ctx context.Context, p *struct {
	Command string
	Args    string
	Type    string
}) *exec.Cmd {
	var cmd *exec.Cmd

	if p.Command != "" {
		// Split command and arguments
		parts := strings.Fields(p.Command)
		if len(parts) > 1 {
			cmd = exec.CommandContext(ctx, parts[0], parts[1:]...)
		} else {
			cmd = exec.CommandContext(ctx, parts[0])
		}

		// Add additional arguments if provided
		if p.Args != "" {
			args := strings.Fields(p.Args)
			cmd.Args = append(cmd.Args, args...)
		}
	} else {
		// Default command based on project type
		switch p.Type {
		case string(types.TypeBackend):
			cmd = exec.CommandContext(ctx, "go", "run", "main.go")
		case string(types.TypeFrontend):
			cmd = exec.CommandContext(ctx, "npm", "start")
		default:
			cmd = exec.CommandContext(ctx, "sh", "-c", "echo 'No command specified'")
		}
	}

	return cmd
}

// prepareEnvironment sets up environment variables
func (m *Manager) prepareEnvironment(p *struct {
	Port        int
	Environment string
	EnvFile     string
	EnvVars     string
	Path        string
}) []string {
	env := os.Environ()

	// Add project-specific environment variables from .env file
	// First, try EnvFile path if specified
	envFilePath := p.EnvFile
	if envFilePath == "" && p.Path != "" {
		// Default to .env in project path
		envFilePath = filepath.Join(p.Path, ".env")
	}
	
	if envFilePath != "" {
		envVarsFromFile := m.loadEnvFile(envFilePath)
		// Merge env vars from file (file takes precedence over system env)
		for k, v := range envVarsFromFile {
			// Remove existing env var with same key
			env = m.removeEnvVar(env, k)
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Parse and add JSON env vars (takes highest precedence)
	if p.EnvVars != "" {
		envVarsFromJSON := m.parseEnvVarsJSON(p.EnvVars)
		for k, v := range envVarsFromJSON {
			// Remove existing env var with same key
			env = m.removeEnvVar(env, k)
			env = append(env, fmt.Sprintf("%s=%s", k, v))
		}
	}

	// Add common environment variables (only if not already set)
	// Check if PORT is already set
	portSet := false
	for _, e := range env {
		if strings.HasPrefix(e, "PORT=") {
			portSet = true
			break
		}
	}
	
	// Only set PORT if not already set and p.Port > 0
	if !portSet && p.Port > 0 {
		env = append(env, fmt.Sprintf("PORT=%d", p.Port))
	}
	
	// Set ENVIRONMENT if not already set
	envSet := false
	for _, e := range env {
		if strings.HasPrefix(e, "ENVIRONMENT=") {
			envSet = true
			break
		}
	}
	if !envSet && p.Environment != "" {
		env = append(env, fmt.Sprintf("ENVIRONMENT=%s", p.Environment))
	}

	return env
}

// loadEnvFile loads environment variables from a .env file
func (m *Manager) loadEnvFile(envPath string) map[string]string {
	envVars := make(map[string]string)
	
	// Check if file exists
	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		return envVars
	}
	
	content, err := os.ReadFile(envPath)
	if err != nil {
		return envVars
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		// Parse KEY=VALUE format
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			// Remove quotes if present
			if len(value) >= 2 && ((value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'')) {
				value = value[1 : len(value)-1]
			}
			envVars[key] = value
		}
	}
	
	return envVars
}

// parseEnvVarsJSON parses JSON string of environment variables
func (m *Manager) parseEnvVarsJSON(envVarsJSON string) map[string]string {
	envVars := make(map[string]string)
	
	var jsonMap map[string]interface{}
	if err := json.Unmarshal([]byte(envVarsJSON), &jsonMap); err != nil {
		return envVars
	}
	
	for k, v := range jsonMap {
		// Convert value to string
		var strValue string
		switch val := v.(type) {
		case string:
			strValue = val
		case float64:
			strValue = strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			strValue = strconv.FormatBool(val)
		case nil:
			strValue = ""
		default:
			strValue = fmt.Sprintf("%v", val)
		}
		envVars[k] = strValue
	}
	
	return envVars
}

// removeEnvVar removes an environment variable from the env slice
func (m *Manager) removeEnvVar(env []string, key string) []string {
	result := make([]string, 0, len(env))
	prefix := key + "="
	for _, e := range env {
		if !strings.HasPrefix(e, prefix) {
			result = append(result, e)
		}
	}
	return result
}

// monitorProcess monitors a running process and handles cleanup
func (m *Manager) monitorProcess(processInfo *ProcessInfo) {
	// Wait for process to finish
	err := processInfo.Process.Wait()

	// Update project status
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	status := string(types.StatusStopped)
	lastError := ""

	if err != nil {
		status = string(types.StatusError)
		lastError = err.Error()
	}

	m.db.Table("projects").Where("id = ?", processInfo.ProjectID).Updates(map[string]interface{}{
		"status":      status,
		"stop_time":   &now,
		"p_id":        0, // Use p_id (snake_case) as GORM converts PID to p_id
		"last_error":  lastError,
	})

	// Clean up
	processInfo.safeCloseChannel()
	delete(m.processes, processInfo.ProjectID)
}

// captureOutput reads from a pipe and sends lines to the logs channel
func (m *Manager) captureOutput(pipe io.ReadCloser, logs chan<- string, isStderr bool) {
	defer pipe.Close()
	scanner := bufio.NewScanner(pipe)
	for scanner.Scan() {
		line := scanner.Text()
		// Add prefix to distinguish stderr
		var logLine string
		if isStderr {
			logLine = fmt.Sprintf("[ERROR] %s", line)
		} else {
			logLine = line
		}
		
		// Send to channel (non-blocking)
		select {
		case logs <- logLine:
		default:
			// Channel full, skip this log line
		}
	}
	if err := scanner.Err(); err != nil {
		errorMsg := fmt.Sprintf("[ERROR] Error reading output: %v", err)
		select {
		case logs <- errorMsg:
		default:
			// Channel might be closed, ignore
		}
	}
}

// stripANSI removes ANSI escape sequences from a string
// ANSI escape codes: \u001b[...m, \033[...m, \x1b[...m
// Also handles JSON-encoded escape sequences like \\u001b
var ansiEscapeRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
var ansiEscapeRegexJSON = regexp.MustCompile(`\\u001b\[[0-9;]*[a-zA-Z]`)

func stripANSI(s string) string {
	// First, handle JSON-encoded escape sequences (\\u001b)
	result := ansiEscapeRegexJSON.ReplaceAllString(s, "")
	
	// Remove actual ANSI escape sequences (\x1b, \033, \u001b)
	result = ansiEscapeRegex.ReplaceAllString(result, "")
	
	// Remove common escape sequence markers
	result = strings.ReplaceAll(result, "\u001b", "")
	result = strings.ReplaceAll(result, "\033", "")
	result = strings.ReplaceAll(result, "\\u001b", "")
	result = strings.ReplaceAll(result, "\\033", "")
	
	// Remove any remaining escape-like patterns
	// Pattern: [number;number;...m or [numberm
	result = regexp.MustCompile(`\[[0-9;]+m`).ReplaceAllString(result, "")
	
	// Clean up multiple spaces that might result from removal
	result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
	
	return strings.TrimSpace(result)
}

// safeSendLog safely sends a log line to the channel, handling closed channel gracefully
func safeSendLog(ch chan<- string, logLine string) bool {
	defer func() {
		// Recover from panic if channel is closed
		if r := recover(); r != nil {
			// Channel is closed, ignore
		}
	}()
	
	select {
	case ch <- logLine:
		return true
	default:
		// Channel full, skip sending
		return false
	}
}

// safeCloseChannel safely closes a channel, handling already-closed channel gracefully
func (p *ProcessInfo) safeCloseChannel() {
	p.closeMu.Lock()
	defer p.closeMu.Unlock()
	
	if p.closed {
		return // Already closed
	}
	
	defer func() {
		// Recover from panic if channel is already closed
		if r := recover(); r != nil {
			// Channel already closed, ignore
		}
		p.closed = true
	}()
	
	close(p.Logs)
}

// captureOutputWithBuffer reads from a pipe, sends to channel, and buffers logs
func (m *Manager) captureOutputWithBuffer(pipe io.ReadCloser, processInfo *ProcessInfo, isStderr bool) {
	defer func() {
		// Recover from any panic (e.g., sending to closed channel)
		if r := recover(); r != nil {
			// Log the error but don't crash
			// Channel might be closed, which is expected when service stops
		}
		pipe.Close()
	}()
	
	scanner := bufio.NewScanner(pipe)
	lastSaveTime := time.Now()
	
	for scanner.Scan() {
		line := scanner.Text()
		
		// Strip ANSI escape codes
		cleanLine := stripANSI(line)
		
		// Skip empty lines after stripping ANSI codes
		if cleanLine == "" {
			continue
		}
		
		// Add prefix to distinguish stderr
		var logLine string
		if isStderr {
			logLine = fmt.Sprintf("[ERROR] %s", cleanLine)
		} else {
			logLine = cleanLine
		}
		
		// Add to buffer
		processInfo.addToLogBuffer(logLine)
		
		// Send to channel safely (handles closed channel)
		safeSendLog(processInfo.Logs, logLine)
		
		// Save logs to database every 2 seconds
		if time.Since(lastSaveTime) > 2*time.Second {
			m.saveLogsToDatabase(processInfo.ProjectID, processInfo.getLogBuffer())
			lastSaveTime = time.Now()
		}
	}
	if err := scanner.Err(); err != nil {
		errorMsg := fmt.Sprintf("[ERROR] Error reading output: %v", err)
		processInfo.addToLogBuffer(errorMsg)
		// Send safely (handles closed channel)
		safeSendLog(processInfo.Logs, errorMsg)
		// Save final logs
		m.saveLogsToDatabase(processInfo.ProjectID, processInfo.getLogBuffer())
	}
}

// addToLogBuffer adds a log line to the buffer (thread-safe)
func (p *ProcessInfo) addToLogBuffer(line string) {
	p.logMu.Lock()
	defer p.logMu.Unlock()
	
	// Add to buffer
	p.LogBuffer = append(p.LogBuffer, line)
	
	// Keep only last 1000 lines
	maxBufferSize := 1000
	if len(p.LogBuffer) > maxBufferSize {
		p.LogBuffer = p.LogBuffer[len(p.LogBuffer)-maxBufferSize:]
	}
}

// getLogBuffer returns a copy of the log buffer (thread-safe)
func (p *ProcessInfo) getLogBuffer() []string {
	p.logMu.Lock()
	defer p.logMu.Unlock()
	
	// Return a copy
	result := make([]string, len(p.LogBuffer))
	copy(result, p.LogBuffer)
	return result
}

// saveLogsToDatabase saves logs to database
func (m *Manager) saveLogsToDatabase(projectID uint, logs []string) {
	if len(logs) == 0 {
		return
	}
	
	// Convert to JSON
	logsJSON, err := json.Marshal(logs)
	if err != nil {
		return // Silently fail, don't block log capture
	}
	
	// Save to database (non-blocking, no lock needed as we're in a goroutine)
	go func() {
		// Use Update without lock since we're in a separate goroutine
		// and database operations are thread-safe
		m.db.Table("projects").Where("id = ?", projectID).Update("logs", string(logsJSON))
	}()
}