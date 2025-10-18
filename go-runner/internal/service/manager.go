package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
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

	if err := m.db.Table("projects").First(&p, projectID).Error; err != nil {
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
	}{
		Port:        p.Port,
		Environment: p.Environment,
		EnvFile:     p.EnvFile,
		EnvVars:     p.EnvVars,
	})

	// Create logs channel
	logs := make(chan string, 100)

	// Store process info
	processInfo := &ProcessInfo{
		ProjectID: projectID,
		Process:   cmd,
		Context:   ctx,
		Cancel:    cancel,
		StartTime: time.Now(),
		Logs:      logs,
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

	// Update project with PID and start time
	now := time.Now()
	m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
		"status":      string(types.StatusRunning),
		"pid":         cmd.Process.Pid,
		"start_time":  &now,
		"last_error":  "",
	})

	// Start log monitoring in goroutine
	go m.monitorProcess(processInfo)

	return nil
}

// StopService stops a microservice
func (m *Manager) StopService(projectID uint) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	processInfo, exists := m.processes[projectID]
	if !exists {
		return fmt.Errorf("service %d is not running", projectID)
	}

	// Update status to stopping
	m.db.Table("projects").Where("id = ?", projectID).Update("status", string(types.StatusStopping))

	// Cancel context to stop the process
	processInfo.Cancel()

	// Wait for process to finish or timeout
	done := make(chan error, 1)
	go func() {
		done <- processInfo.Process.Wait()
	}()

	select {
	case <-done:
		// Process finished normally
	case <-time.After(30 * time.Second):
		// Force kill if timeout
		processInfo.Process.Process.Kill()
	}

	// Update project status
	now := time.Now()
	m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
		"status":     string(types.StatusStopped),
		"stop_time":  &now,
		"pid":        0,
	})

	// Clean up
	close(processInfo.Logs)
	delete(m.processes, projectID)

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
	m.mu.RLock()
	defer m.mu.RUnlock()

	var p map[string]interface{}
	if err := m.db.Table("projects").First(&p, projectID).Error; err != nil {
		return nil, err
	}

	// Check if process is still running
	if processInfo, exists := m.processes[projectID]; exists {
		if processInfo.Process.ProcessState != nil && processInfo.Process.ProcessState.Exited() {
			// Process has exited, update status
			now := time.Now()
			m.db.Table("projects").Where("id = ?", projectID).Updates(map[string]interface{}{
				"status":     string(types.StatusStopped),
				"stop_time":  &now,
				"pid":        0,
			})
			delete(m.processes, projectID)
		}
	}

	return p, nil
}

// GetRunningServices returns all currently running services
func (m *Manager) GetRunningServices() []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var services []map[string]interface{}
	for _, processInfo := range m.processes {
		var p map[string]interface{}
		if err := m.db.Table("projects").First(&p, processInfo.ProjectID).Error; err == nil {
			services = append(services, p)
		}
	}
	return services
}

// GetServiceLogs returns logs for a service
func (m *Manager) GetServiceLogs(projectID uint) <-chan string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if processInfo, exists := m.processes[projectID]; exists {
		return processInfo.Logs
	}
	return nil
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
}) []string {
	env := os.Environ()

	// Add project-specific environment variables
	if p.EnvFile != "" {
		// TODO: Load .env file
	}

	if p.EnvVars != "" {
		// TODO: Parse JSON env vars and add to environment
	}

	// Add common environment variables
	env = append(env, fmt.Sprintf("PORT=%d", p.Port))
	env = append(env, fmt.Sprintf("ENVIRONMENT=%s", p.Environment))

	return env
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
		"pid":         0,
		"last_error":  lastError,
	})

	// Clean up
	close(processInfo.Logs)
	delete(m.processes, processInfo.ProjectID)
}