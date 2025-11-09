package project

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"go-runner/internal/middleware"
	"go-runner/internal/service"
	"go-runner/internal/websocket"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	manager *service.Manager
	hub     *websocket.Hub
	// Track last time buffered logs were sent for each project to avoid duplicates on refresh
	lastBufferedLogsSent map[uint]time.Time
	bufferedLogsMu       sync.RWMutex
}

func NewHandler(db *gorm.DB, manager *service.Manager, hub *websocket.Hub) *Handler {
	return &Handler{
		db:                   db,
		manager:              manager,
		hub:                  hub,
		lastBufferedLogsSent: make(map[uint]time.Time),
	}
}

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, manager *service.Manager, hub *websocket.Hub) {
	h := NewHandler(db, manager, hub)
	
	// Project routes
	projects := r.Group("/projects")
	{
		projects.GET("", h.GetProjects)
		projects.POST("", h.CreateProject)
		projects.GET("/:id", h.GetProject)
		projects.PUT("/:id", h.UpdateProject)
		projects.DELETE("/:id", h.DeleteProject)
		projects.POST("/:id/start", h.StartProject)
		projects.POST("/:id/stop", h.StopProject)
		projects.POST("/:id/restart", h.RestartProject)
		projects.POST("/:id/force-kill", h.ForceKillProject)
		projects.GET("/:id/status", h.GetProjectStatus)
		projects.GET("/:id/logs", h.GetLogs)
		projects.GET("/:id/logs/ws", h.StreamLogs)
		projects.POST("/:id/install", h.InstallPackages)
		projects.GET("/:id/terminal", h.GetTerminalUrl)
		projects.POST("/:id/terminal/open", h.OpenTerminal)
		projects.POST("/import", h.ImportProjects)
		projects.GET("/:id/config", h.GetProjectConfig)
		projects.PUT("/:id/config", h.UpdateProjectFromConfig)
		projects.POST("/detect-services", h.DetectServices)
	}

	// Project group routes
	groups := r.Group("/groups")
	{
		groups.GET("", h.GetProjectGroups)
		groups.POST("", h.CreateProjectGroup)
		groups.GET("/:id", h.GetProjectGroup)
		groups.PUT("/:id", h.UpdateProjectGroup)
		groups.DELETE("/:id", h.DeleteProjectGroup)
		groups.GET("/:id/projects", h.GetGroupProjects)
	}

	// Service management routes
	services := r.Group("/services")
	{
		services.GET("/running", h.GetRunningServices)
		services.POST("/:id/start", h.StartProject)
		services.POST("/:id/stop", h.StopProject)
		services.POST("/:id/restart", h.RestartProject)
	}

	// Port management routes
	ports := r.Group("/ports")
	{
		ports.GET("", h.GetPorts)
		ports.DELETE("/:port", h.KillPort)
	}
}

// GetProjects godoc
// @Summary      Get all projects
// @Description  Get a list of all projects
// @Tags         projects
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "List of projects"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /projects [get]
func (h *Handler) GetProjects(c *gin.Context) {
	var projects []Project
	if err := h.db.Find(&projects).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch projects", err.Error()))
		return
	}
	
	// Verify status for each project and update if needed
	// This ensures status is accurate when listing projects
	for i := range projects {
		project := &projects[i]
		// Verify status using GetServiceStatus (this auto-updates DB if status is wrong)
		updatedProject, err := h.manager.GetServiceStatus(project.ID)
		if err == nil {
			// Update status from verified result
			if status, ok := updatedProject["status"].(string); ok {
				project.Status = ServiceStatus(status)
			}
			if pid, ok := updatedProject["p_id"].(int); ok {
				project.PID = pid
			} else if pidFloat, ok := updatedProject["p_id"].(float64); ok {
				project.PID = int(pidFloat)
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// GetProject godoc
// @Summary      Get a project by ID
// @Description  Get a specific project by its ID
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Project ID"
// @Success      200  {object}  map[string]interface{}  "Project data"
// @Failure      400  {object}  map[string]interface{}  "Bad request"
// @Failure      404  {object}  map[string]interface{}  "Project not found"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /projects/{id} [get]
func (h *Handler) GetProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	// Use GetServiceStatus to get project with accurate status
	project, err := h.manager.GetServiceStatus(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.ErrNotFound)
			return
		}
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch project", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": project})
}

// CreateProject godoc
// @Summary      Create a new project
// @Description  Create a new project with the provided data
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project  body      Project  true  "Project data"
// @Success      201     {object}  map[string]interface{}  "Created project"
// @Failure      400     {object}  map[string]interface{}  "Bad request"
// @Failure      500     {object}  map[string]interface{}  "Internal server error"
// @Router       /projects [post]
func (h *Handler) CreateProject(c *gin.Context) {
	var project Project
	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Create(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": project})
}

func (h *Handler) UpdateProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var project Project
	if err := h.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&project); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.db.Save(&project).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": project})
}

func (h *Handler) DeleteProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.db.Delete(&Project{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

func (h *Handler) StartProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Broadcast status update via WebSocket before starting
	h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
		"project_id": id,
		"status":     "starting",
		"message":    "Project is starting...",
	})

	if err := h.manager.StartService(uint(id)); err != nil {
		// Broadcast error status
		h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
			"project_id": id,
			"status":     "error",
			"message":    fmt.Sprintf("Failed to start: %v", err),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Wait a bit to verify process started successfully
	time.Sleep(500 * time.Millisecond)

	// Get updated status from database
	var project Project
	if err := h.db.First(&project, id).Error; err == nil {
		// Broadcast actual status (should be "running" if successful)
		h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
			"project_id": id,
			"status":     project.Status,
			"message":    fmt.Sprintf("Project status: %s", project.Status),
		})
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project started successfully", "project_id": id})
}

func (h *Handler) StopProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.manager.StopService(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast status update via WebSocket
	h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
		"project_id": id,
		"status":     "stopped",
		"message":    "Project stopped",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Project stopped successfully", "project_id": id})
}

func (h *Handler) ForceKillProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Broadcast status update via WebSocket
	h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
		"project_id": id,
		"status":     "stopping",
		"message":    "Force killing project...",
	})

	if err := h.manager.ForceKillService(uint(id)); err != nil {
		// Broadcast error status
		h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
			"project_id": id,
			"status":     "error",
			"message":    fmt.Sprintf("Failed to force kill: %v", err),
		})
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast status update via WebSocket
	h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
		"project_id": id,
		"status":     "stopped",
		"message":    "Project force killed",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Project force killed successfully", "project_id": id})
}

func (h *Handler) RestartProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if err := h.manager.RestartService(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast status update via WebSocket
	h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
		"project_id": id,
		"status":     "restarting",
		"message":    "Project is restarting...",
	})

	c.JSON(http.StatusOK, gin.H{"message": "Project restarted successfully", "project_id": id})
}

func (h *Handler) GetProjectStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, err := h.manager.GetServiceStatus(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": project})
}

func (h *Handler) GetLogs(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// First try to get logs from running service (memory buffer)
	logs := h.manager.GetServiceLogBuffer(uint(id))
	
	// If no logs in memory, try to get from database
	if len(logs) == 0 {
		var project Project
		if err := h.db.First(&project, id).Error; err == nil && project.Logs != "" {
			// Parse logs from database
			var dbLogs []string
			if err := json.Unmarshal([]byte(project.Logs), &dbLogs); err == nil {
				logs = dbLogs
			}
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"logs": logs,
			"count": len(logs),
		},
	})
}

func (h *Handler) StreamLogs(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Upgrade to WebSocket first
	h.hub.HandleProjectWebSocket(c)

	// Start streaming logs in background
	go func() {
		// Wait a bit for WebSocket to establish
		time.Sleep(200 * time.Millisecond)
		
		// Check if we've sent buffered logs recently for this project
		// If yes, this is likely a refresh/reconnect - don't send buffered logs again
		h.bufferedLogsMu.RLock()
		lastSent, hasSent := h.lastBufferedLogsSent[uint(id)]
		timeSinceLastSent := time.Since(lastSent)
		h.bufferedLogsMu.RUnlock()
		
		// Only send buffered logs if:
		// 1. We've never sent them before, OR
		// 2. It's been more than 10 seconds since we last sent them (new page load, not refresh)
		shouldSendBuffered := !hasSent || timeSinceLastSent > 10*time.Second
		
		if shouldSendBuffered {
			// First, try to get logs from running service (memory buffer)
			bufferedLogs := h.manager.GetServiceLogBuffer(uint(id))
			
			// If no logs in memory, try to get from database
			if len(bufferedLogs) == 0 {
				var project Project
				if err := h.db.First(&project, id).Error; err == nil && project.Logs != "" {
					// Parse logs from database
					var dbLogs []string
					if err := json.Unmarshal([]byte(project.Logs), &dbLogs); err == nil {
						bufferedLogs = dbLogs
					}
				}
			}
			
			// Send buffered logs if available
			if len(bufferedLogs) > 0 {
				// Send buffered logs
				for _, logLine := range bufferedLogs {
					h.hub.BroadcastToProject(uint(id), "log", logLine)
					time.Sleep(10 * time.Millisecond) // Small delay to avoid overwhelming
				}
				// Send separator if service is running
				logs := h.manager.GetServiceLogs(uint(id))
				if logs != nil {
					h.hub.BroadcastToProject(uint(id), "log", "--- Live logs ---")
				}
				
				// Mark that we've sent buffered logs for this project
				h.bufferedLogsMu.Lock()
				h.lastBufferedLogsSent[uint(id)] = time.Now()
				h.bufferedLogsMu.Unlock()
			}
		}
		
		// Check if service is actually running
		isRunning := h.manager.IsServiceRunning(uint(id))
		
		if !isRunning {
			// Service not running - check and update status in DB if needed
			var project Project
			if err := h.db.First(&project, id).Error; err == nil {
				statusChanged := false
				if project.Status == "running" || project.Status == "starting" {
					// Update status to stopped
					h.db.Model(&project).Updates(map[string]interface{}{
						"status": "stopped",
						"p_id":   0,
					})
					statusChanged = true
				}
				
				// Only send message if we haven't sent buffered logs recently (to avoid duplicate messages)
				if shouldSendBuffered {
					// Get buffered logs for message
					bufferedLogs := h.manager.GetServiceLogBuffer(uint(id))
					if len(bufferedLogs) == 0 {
						var project Project
						if err := h.db.First(&project, id).Error; err == nil && project.Logs != "" {
							var dbLogs []string
							if err := json.Unmarshal([]byte(project.Logs), &dbLogs); err == nil {
								bufferedLogs = dbLogs
							}
						}
					}
					
					// Send appropriate message with more context
					if len(bufferedLogs) == 0 {
						h.hub.BroadcastToProject(uint(id), "log", fmt.Sprintf("[INFO] Service '%s' is not running. No logs available.", project.Name))
						if statusChanged {
							h.hub.BroadcastToProject(uint(id), "log", "[WARN] Service was marked as running but process is not active. Status updated to 'stopped'.")
							h.hub.BroadcastToProject(uint(id), "log", "[INFO] Please start the service again to see live logs.")
						} else {
							h.hub.BroadcastToProject(uint(id), "log", "[INFO] Please start the service to see live logs.")
						}
					} else {
						h.hub.BroadcastToProject(uint(id), "log", fmt.Sprintf("[INFO] Service '%s' is not running. Showing last logs from database.", project.Name))
						if statusChanged {
							h.hub.BroadcastToProject(uint(id), "log", "[WARN] Service was marked as running but process is not active. Status updated to 'stopped'.")
						}
						h.hub.BroadcastToProject(uint(id), "log", "[INFO] To see live logs, please start the service.")
					}
				}
			} else {
				// Project not found
				if shouldSendBuffered {
					h.hub.BroadcastToProject(uint(id), "log", "[ERROR] Project not found.")
				}
			}
			return
		}

		// Service is running, stream new logs from channel
		logs := h.manager.GetServiceLogs(uint(id))
		if logs == nil {
			// Fallback: service might have just stopped
			if shouldSendBuffered {
				h.hub.BroadcastToProject(uint(id), "log", "[INFO] Service stopped. Showing last logs from database.")
			}
			return
		}

		// Stream new logs (only new logs from channel, not buffered)
		for logLine := range logs {
			h.hub.BroadcastToProject(uint(id), "log", logLine)
		}
	}()
}

func (h *Handler) GetRunningServices(c *gin.Context) {
	services := h.manager.GetRunningServices()
	c.JSON(http.StatusOK, gin.H{"data": services})
}

// Project Group handlers
func (h *Handler) GetProjectGroups(c *gin.Context) {
	var groups []ProjectGroup
	if err := h.db.Preload("Projects").Find(&groups).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": groups})
}

func (h *Handler) CreateProjectGroup(c *gin.Context) {
	var req CreateProjectGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	group := ProjectGroup{
		Name:        req.Name,
		Description: req.Description,
		Color:       req.Color,
	}

	if err := h.db.Create(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"data": group})
}

func (h *Handler) GetProjectGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group ProjectGroup
	if err := h.db.Preload("Projects").First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": group})
}

func (h *Handler) UpdateProjectGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var group ProjectGroup
	if err := h.db.First(&group, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Group not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req UpdateProjectGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields if provided
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = *req.Description
	}
	if req.Color != nil {
		group.Color = *req.Color
	}

	if err := h.db.Save(&group).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": group})
}

func (h *Handler) DeleteProjectGroup(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	if err := h.db.Delete(&ProjectGroup{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Group deleted successfully"})
}

func (h *Handler) GetGroupProjects(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID"})
		return
	}

	var projects []Project
	if err := h.db.Where("group_id = ?", id).Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": projects})
}

// InstallPackagesRequest represents the request to install packages
type InstallPackagesRequest struct {
	PackageManager string   `json:"package_manager" binding:"required,oneof=npm yarn pnpm go pip"`
	Packages       []string `json:"packages"`
}

// InstallPackages installs packages for a project
func (h *Handler) InstallPackages(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	var req InstallPackagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	// Get project from database
	var project Project
	if err := h.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.ErrNotFound)
			return
		}
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch project", err.Error()))
		return
	}

	// Determine working directory
	workingDir := project.Path
	if project.WorkingDir != "" {
		workingDir = project.WorkingDir
	}

	// Build command based on package manager
	var cmd *exec.Cmd
	switch req.PackageManager {
	case "npm":
		if len(req.Packages) > 0 {
			cmd = exec.Command("npm", append([]string{"install"}, req.Packages...)...)
		} else {
			cmd = exec.Command("npm", "install")
		}
	case "yarn":
		if len(req.Packages) > 0 {
			cmd = exec.Command("yarn", append([]string{"add"}, req.Packages...)...)
		} else {
			cmd = exec.Command("yarn", "install")
		}
	case "pnpm":
		if len(req.Packages) > 0 {
			cmd = exec.Command("pnpm", append([]string{"add"}, req.Packages...)...)
		} else {
			cmd = exec.Command("pnpm", "install")
		}
	case "go":
		if len(req.Packages) > 0 {
			cmd = exec.Command("go", append([]string{"get"}, req.Packages...)...)
		} else {
			cmd = exec.Command("go", "mod", "download")
		}
	case "pip":
		if len(req.Packages) > 0 {
			cmd = exec.Command("pip", append([]string{"install"}, req.Packages...)...)
		} else {
			cmd = exec.Command("pip", "install", "-r", "requirements.txt")
		}
	default:
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Unsupported package manager", fmt.Sprintf("Package manager %s is not supported", req.PackageManager)))
		return
	}

	cmd.Dir = workingDir

	// Execute command
	output, err := cmd.CombinedOutput()
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to install packages", string(output)))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Packages installed successfully",
		"output":  string(output),
	})
}

// GetTerminalUrl returns terminal access information for a project
func (h *Handler) GetTerminalUrl(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	// Get project from database
	var project Project
	if err := h.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.ErrNotFound)
			return
		}
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch project", err.Error()))
		return
	}

	// Determine working directory
	workingDir := project.Path
	if project.WorkingDir != "" {
		workingDir = project.WorkingDir
	}

	// For now, return the path and instructions
	// In a real implementation, you might want to integrate with a web terminal like xterm.js
	// or provide instructions for opening a terminal in the project directory
	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		absPath = workingDir
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"path":        absPath,
			"working_dir": absPath,
			"instructions": fmt.Sprintf(
				"To open a terminal for this project, navigate to: %s\n"+
					"On macOS/Linux: cd %s\n"+
					"On Windows: cd /d %s",
				absPath, absPath, absPath,
			),
		},
	})
}

// OpenTerminalRequest represents the request to open a terminal
type OpenTerminalRequest struct {
	OS string `json:"os"` // "macos", "linux", "windows", or "auto" (auto-detect from User-Agent)
}

// OpenTerminal creates a script to open terminal in the project directory
func (h *Handler) OpenTerminal(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	// Get project from database
	var project Project
	if err := h.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.ErrNotFound)
			return
		}
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch project", err.Error()))
		return
	}

	// Determine working directory
	workingDir := project.Path
	if project.WorkingDir != "" {
		workingDir = project.WorkingDir
	}

	absPath, err := filepath.Abs(workingDir)
	if err != nil {
		absPath = workingDir
	}

	// Detect OS from User-Agent or request
	var req OpenTerminalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Try to detect from User-Agent
		userAgent := c.GetHeader("User-Agent")
		if strings.Contains(strings.ToLower(userAgent), "mac") || strings.Contains(strings.ToLower(userAgent), "darwin") {
			req.OS = "macos"
		} else if strings.Contains(strings.ToLower(userAgent), "linux") {
			req.OS = "linux"
		} else if strings.Contains(strings.ToLower(userAgent), "win") {
			req.OS = "windows"
		} else {
			req.OS = "auto"
		}
	}

	// Escape path for shell commands
	// For paths with spaces or special characters, we need to quote them properly
	// For AppleScript: use single quotes inside the double-quoted string to avoid escaping issues
	escapedPathForAppleScript := strings.ReplaceAll(absPath, `'`, `'\''`) // Escape single quotes for AppleScript
	// For shell commands in general, escape double quotes
	escapedPath := strings.ReplaceAll(absPath, `"`, `\"`)

	// Generate command based on OS
	var command string
	var instructions string

	switch req.OS {
	case "macos":
		// macOS: Use osascript to open Terminal.app
		// Use AppleScript's "quoted form of" to safely handle any path with special characters
		// This is the most reliable method that handles spaces, quotes, and other special chars
		escapedPathForAppleScript = strings.ReplaceAll(absPath, `"`, `\"`)
		escapedPathForAppleScript = strings.ReplaceAll(escapedPathForAppleScript, `\`, `\\`)
		command = fmt.Sprintf(`osascript -e 'tell application "Terminal" to do script "cd " & quoted form of (POSIX path of "%s")'`, escapedPathForAppleScript)
		instructions = fmt.Sprintf(
			"Command copied to clipboard!\n\n"+
				"To open terminal:\n"+
				"1. Open Terminal app\n"+
				"2. Paste the command (Cmd+V)\n"+
				"3. Press Enter\n\n"+
				"Or manually run:\n"+
				"cd \"%s\"",
			absPath,
		)
	case "linux":
		// Linux: Try different terminal emulators
		command = fmt.Sprintf(`gnome-terminal --working-directory="%s" 2>/dev/null || xterm -e "cd \"%s\" && exec $SHELL" 2>/dev/null || konsole --workdir "%s" 2>/dev/null || echo "Please install a terminal emulator (gnome-terminal, xterm, or konsole)"`, escapedPath, escapedPath, escapedPath)
		instructions = fmt.Sprintf(
			"Command copied to clipboard!\n\n"+
				"To open terminal:\n"+
				"1. Open a terminal\n"+
				"2. Paste the command (Ctrl+Shift+V or right-click)\n"+
				"3. Press Enter\n\n"+
				"Or manually run:\n"+
				"cd \"%s\"",
			absPath,
		)
	case "windows":
		// Windows: Use PowerShell or cmd
		command = fmt.Sprintf(`powershell -NoExit -Command "cd '%s'" 2>$null || cmd /k "cd /d \"%s\""`, absPath, escapedPath)
		instructions = fmt.Sprintf(
			"Command copied to clipboard!\n\n"+
				"To open terminal:\n"+
				"1. Open Command Prompt or PowerShell\n"+
				"2. Paste the command (right-click or Ctrl+V)\n"+
				"3. Press Enter\n\n"+
				"Or manually run:\n"+
				"cd /d \"%s\"",
			absPath,
		)
	default:
		// Provide commands for all platforms
		escapedForAppleScript := strings.ReplaceAll(absPath, `"`, `\"`)
		escapedForAppleScript = strings.ReplaceAll(escapedForAppleScript, `\`, `\\`)
		macosCmd := fmt.Sprintf(`osascript -e 'tell application "Terminal" to do script "cd " & quoted form of (POSIX path of "%s")'`, escapedForAppleScript)
		
		c.JSON(http.StatusOK, gin.H{
			"data": gin.H{
				"path":        absPath,
				"working_dir": absPath,
				"commands": map[string]string{
					"macos":   macosCmd,
					"linux":   fmt.Sprintf(`gnome-terminal --working-directory="%s"`, escapedPath),
					"windows": fmt.Sprintf(`cmd /k "cd /d \"%s\""`, escapedPath),
				},
				"simple_command": fmt.Sprintf(`cd "%s"`, absPath),
				"instructions": fmt.Sprintf(
					"Copy and run one of these commands in your terminal:\n\n"+
						"macOS:\n"+
						"  %s\n\n"+
						"Linux:\n"+
						"  gnome-terminal --working-directory=\"%s\"\n\n"+
						"Windows:\n"+
						"  cmd /k \"cd /d \\\"%s\\\"\"\n\n"+
						"Or simply navigate to:\n"+
						"  cd \"%s\"",
					macosCmd, absPath, absPath, absPath,
				),
			},
		})
		return
	}

	// Return command and instructions as JSON
	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"path":           absPath,
			"working_dir":    absPath,
			"command":        command,
			"simple_command": fmt.Sprintf(`cd "%s"`, absPath),
			"os":             req.OS,
			"instructions":  instructions,
		},
	})
}

// GetPorts returns list of ports in use
func (h *Handler) GetPorts(c *gin.Context) {
	ports, err := h.manager.GetPortsInUse()
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get ports", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": ports})
}

// KillPort kills the process using the specified port
func (h *Handler) KillPort(c *gin.Context) {
	portStr := c.Param("port")
	port, err := strconv.Atoi(portStr)
	if err != nil || port <= 0 || port > 65535 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Invalid port", "Port must be between 1 and 65535"))
		return
	}

	err = h.manager.KillPort(port)
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to kill port", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Process on port %d has been killed", port),
		"port":    port,
	})
}

// ImportProjectsRequest represents the import request
type ImportProjectsRequest struct {
	Projects []CreateProjectRequest `json:"projects"`
	Groups   []CreateProjectGroupRequest `json:"groups"`
}

// ImportProjects imports multiple projects from config file
func (h *Handler) ImportProjects(c *gin.Context) {
	// Check if it's a file upload
	file, err := c.FormFile("file")
	if err == nil {
		// Handle file upload
		src, err := file.Open()
		if err != nil {
			middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Failed to open file", err.Error()))
			return
		}
		defer src.Close()

		content, err := io.ReadAll(src)
		if err != nil {
			middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Failed to read file", err.Error()))
			return
		}

		var importData ImportProjectsRequest

		// Try to parse as YAML first
		if strings.HasSuffix(strings.ToLower(file.Filename), ".yaml") || strings.HasSuffix(strings.ToLower(file.Filename), ".yml") {
			if err := yaml.Unmarshal(content, &importData); err != nil {
				middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Failed to parse YAML", err.Error()))
				return
			}
		} else {
			// Try JSON
			if err := json.Unmarshal(content, &importData); err != nil {
				middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Failed to parse JSON", err.Error()))
				return
			}
		}

		result := h.processImport(importData)
		c.JSON(http.StatusOK, gin.H{
			"message": "Import completed",
			"data":    result,
		})
		return
	}

	// Handle JSON body
	var importData ImportProjectsRequest
	if err := c.ShouldBindJSON(&importData); err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	result := h.processImport(importData)
	c.JSON(http.StatusOK, gin.H{
		"message": "Import completed",
		"data":    result,
	})
}

// processImport processes the import data
func (h *Handler) processImport(importData ImportProjectsRequest) map[string]interface{} {
	result := map[string]interface{}{
		"groups_created":   0,
		"groups_updated":   0,
		"projects_created": 0,
		"projects_updated": 0,
		"errors":           []string{},
	}

	// Create/Update groups first
	groupMap := make(map[string]uint)
	for _, groupReq := range importData.Groups {
		var group ProjectGroup
		if err := h.db.Where("name = ?", groupReq.Name).First(&group).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new group
				group = ProjectGroup{
					Name:        groupReq.Name,
					Description: groupReq.Description,
					Color:       groupReq.Color,
				}
				if err := h.db.Create(&group).Error; err != nil {
					result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Failed to create group %s: %v", groupReq.Name, err))
					continue
				}
				result["groups_created"] = result["groups_created"].(int) + 1
			} else {
				result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Failed to query group %s: %v", groupReq.Name, err))
				continue
			}
		} else {
			// Update existing group
			if groupReq.Description != "" {
				group.Description = groupReq.Description
			}
			if groupReq.Color != "" {
				group.Color = groupReq.Color
			}
			if err := h.db.Save(&group).Error; err != nil {
				result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Failed to update group %s: %v", groupReq.Name, err))
				continue
			}
			result["groups_updated"] = result["groups_updated"].(int) + 1
		}
		groupMap[groupReq.Name] = group.ID
	}

	// Create/Update projects
	for _, projectReq := range importData.Projects {
		// Validate path exists
		if _, err := os.Stat(projectReq.Path); os.IsNotExist(err) {
			result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Path does not exist for project %s: %s", projectReq.Name, projectReq.Path))
			continue
		}

		var project Project
		if err := h.db.Where("name = ?", projectReq.Name).First(&project).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new project
				project = Project{
					Name:        projectReq.Name,
					Type:        projectReq.Type,
					Path:        projectReq.Path,
					Description: projectReq.Description,
				}

				// Set optional fields
				if projectReq.GroupID != nil {
					project.GroupID = projectReq.GroupID
				}
				if projectReq.Command != "" {
					project.Command = projectReq.Command
				}
				if projectReq.Args != "" {
					project.Args = projectReq.Args
				}
				if projectReq.WorkingDir != "" {
					project.WorkingDir = projectReq.WorkingDir
				}
				if projectReq.Port > 0 {
					project.Port = projectReq.Port
				}
				if projectReq.Ports != "" {
					project.Ports = projectReq.Ports
				}
				if projectReq.Environment != "" {
					project.Environment = projectReq.Environment
				}
				if projectReq.EnvFile != "" {
					project.EnvFile = projectReq.EnvFile
				}
				if projectReq.EnvVars != "" {
					project.EnvVars = projectReq.EnvVars
				}
				if projectReq.Editor != "" {
					project.Editor = projectReq.Editor
				}
				if projectReq.EditorArgs != "" {
					project.EditorArgs = projectReq.EditorArgs
				}
				if projectReq.HealthCheckURL != "" {
					project.HealthCheckURL = projectReq.HealthCheckURL
				}
				project.AutoRestart = projectReq.AutoRestart
				if projectReq.MaxRestarts > 0 {
					project.MaxRestarts = projectReq.MaxRestarts
				}
				if projectReq.CPULimit != "" {
					project.CPULimit = projectReq.CPULimit
				}
				if projectReq.MemoryLimit != "" {
					project.MemoryLimit = projectReq.MemoryLimit
				}

				if err := h.db.Create(&project).Error; err != nil {
					result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Failed to create project %s: %v", projectReq.Name, err))
					continue
				}
				result["projects_created"] = result["projects_created"].(int) + 1
			} else {
				result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Failed to query project %s: %v", projectReq.Name, err))
				continue
			}
		} else {
			// Update existing project - only update if fields are provided and not empty
			if projectReq.Description != "" {
				project.Description = projectReq.Description
			}
			if projectReq.Path != "" {
				project.Path = projectReq.Path
			}
			if projectReq.Command != "" {
				project.Command = projectReq.Command
			}
			if projectReq.Port > 0 {
				project.Port = projectReq.Port
			}
			if projectReq.Environment != "" {
				project.Environment = projectReq.Environment
			}
			// AutoRestart is a bool, so we always update it
			project.AutoRestart = projectReq.AutoRestart

			if err := h.db.Save(&project).Error; err != nil {
				result["errors"] = append(result["errors"].([]string), fmt.Sprintf("Failed to update project %s: %v", projectReq.Name, err))
				continue
			}
			result["projects_updated"] = result["projects_updated"].(int) + 1
		}
	}

	return result
}

// GetProjectConfig returns the project configuration in YAML or JSON format
func (h *Handler) GetProjectConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	var project Project
	if err := h.db.Preload("Group").First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.ErrNotFound)
			return
		}
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch project", err.Error()))
		return
	}

	format := c.Query("format")
	if format == "" {
		format = "yaml"
	}

	// Create config structure
	config := map[string]interface{}{
		"name":           project.Name,
		"description":    project.Description,
		"type":           project.Type,
		"path":           project.Path,
		"command":        project.Command,
		"args":           project.Args,
		"working_dir":    project.WorkingDir,
		"port":           project.Port,
		"ports":          project.Ports,
		"environment":    project.Environment,
		"env_file":       project.EnvFile,
		"env_vars":       project.EnvVars,
		"editor":         project.Editor,
		"editor_args":    project.EditorArgs,
		"health_check_url": project.HealthCheckURL,
		"auto_restart":   project.AutoRestart,
		"max_restarts":   project.MaxRestarts,
		"cpu_limit":      project.CPULimit,
		"memory_limit":   project.MemoryLimit,
	}

	if project.GroupID != nil {
		config["group_id"] = *project.GroupID
	}

	if format == "json" {
		c.Header("Content-Type", "application/json")
		jsonData, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to marshal JSON", err.Error()))
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": string(jsonData)})
		return
	}

	// Convert to YAML
	yamlData, err := yaml.Marshal(config)
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to marshal YAML", err.Error()))
		return
	}

	c.Header("Content-Type", "application/x-yaml")
	c.JSON(http.StatusOK, gin.H{"data": string(yamlData)})
}

// UpdateProjectFromConfigRequest represents the request to update project from config
type UpdateProjectFromConfigRequest struct {
	Config string `json:"config" binding:"required"`
	Format string `json:"format"` // yaml or json
}

// UpdateProjectFromConfig updates a project from YAML or JSON config
func (h *Handler) UpdateProjectFromConfig(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		middleware.HandleError(c, middleware.ErrBadRequest)
		return
	}

	var req UpdateProjectFromConfigRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	// Get existing project
	var project Project
	if err := h.db.First(&project, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			middleware.HandleError(c, middleware.ErrNotFound)
			return
		}
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to fetch project", err.Error()))
		return
	}

	// Parse config
	format := req.Format
	if format == "" {
		// Try to detect format
		if strings.TrimSpace(req.Config)[0] == '{' {
			format = "json"
		} else {
			format = "yaml"
		}
	}

	var configMap map[string]interface{}
	if format == "json" {
		if err := json.Unmarshal([]byte(req.Config), &configMap); err != nil {
			middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Failed to parse JSON", err.Error()))
			return
		}
	} else {
		if err := yaml.Unmarshal([]byte(req.Config), &configMap); err != nil {
			middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Failed to parse YAML", err.Error()))
			return
		}
	}

	// Update project fields
	if name, ok := configMap["name"].(string); ok && name != "" {
		project.Name = name
	}
	if desc, ok := configMap["description"].(string); ok {
		project.Description = desc
	}
	if typ, ok := configMap["type"].(string); ok {
		project.Type = ServiceType(typ)
	}
	if path, ok := configMap["path"].(string); ok && path != "" {
		// Validate path exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Path does not exist", path))
			return
		}
		project.Path = path
	}
	if cmd, ok := configMap["command"].(string); ok {
		project.Command = cmd
	}
	if args, ok := configMap["args"].(string); ok {
		project.Args = args
	}
	if wd, ok := configMap["working_dir"].(string); ok {
		project.WorkingDir = wd
	}
	if port, ok := configMap["port"].(int); ok {
		project.Port = port
	} else if port, ok := configMap["port"].(float64); ok {
		project.Port = int(port)
	}
	if ports, ok := configMap["ports"].(string); ok {
		project.Ports = ports
	}
	if env, ok := configMap["environment"].(string); ok {
		project.Environment = env
	}
	if envFile, ok := configMap["env_file"].(string); ok {
		project.EnvFile = envFile
	}
	if envVars, ok := configMap["env_vars"].(string); ok {
		project.EnvVars = envVars
	}
	if editor, ok := configMap["editor"].(string); ok {
		project.Editor = editor
	}
	if editorArgs, ok := configMap["editor_args"].(string); ok {
		project.EditorArgs = editorArgs
	}
	if healthURL, ok := configMap["health_check_url"].(string); ok {
		project.HealthCheckURL = healthURL
	}
	if autoRestart, ok := configMap["auto_restart"].(bool); ok {
		project.AutoRestart = autoRestart
	}
	if maxRestarts, ok := configMap["max_restarts"].(int); ok {
		project.MaxRestarts = maxRestarts
	} else if maxRestarts, ok := configMap["max_restarts"].(float64); ok {
		project.MaxRestarts = int(maxRestarts)
	}
	if cpuLimit, ok := configMap["cpu_limit"].(string); ok {
		project.CPULimit = cpuLimit
	}
	if memLimit, ok := configMap["memory_limit"].(string); ok {
		project.MemoryLimit = memLimit
	}
	if groupID, ok := configMap["group_id"].(float64); ok {
		gid := uint(groupID)
		project.GroupID = &gid
	} else if groupID, ok := configMap["group_id"].(int); ok {
		gid := uint(groupID)
		project.GroupID = &gid
	}

	// Save project
	if err := h.db.Save(&project).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to update project", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Project updated successfully",
		"data":    project,
	})
}

// DetectServicesRequest represents the request to detect services in a path
type DetectServicesRequest struct {
	Path string `json:"path" binding:"required"`
}

// ServiceDetection represents a detected service
type ServiceDetection struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Path        string `json:"path"`
	Command     string `json:"command"`
	Port        int    `json:"port"`
	PackageFile string `json:"package_file"`
}

// DetectServices detects services in a project path
func (h *Handler) DetectServices(c *gin.Context) {
	var req DetectServicesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Invalid request", err.Error()))
		return
	}

	// Validate path exists
	if _, err := os.Stat(req.Path); os.IsNotExist(err) {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Path does not exist", req.Path))
		return
	}

	services := []ServiceDetection{}

	// Track visited directories to avoid duplicates
	visitedDirs := make(map[string]bool)

	// List of directories to skip (common dependency/build directories)
	skipDirs := map[string]bool{
		"node_modules": true,
		".git":         true,
		"vendor":       true,
		"dist":         true,
		"build":        true,
		".next":        true,
		".nuxt":        true,
		"coverage":     true,
		".env":         true,
		".vscode":      true,
		".idea":        true,
		"__pycache__":  true,
		".pytest_cache": true,
		".venv":        true,
		"venv":         true,
		"env":          true,
		"target":       true,
		"bin":          true,
		"obj":          true,
		".gradle":      true,
		".mvn":         true,
	}

	// Maximum depth to scan (3 levels from base path)
	maxDepth := 3

	// Scan for common service types
	err := filepath.Walk(req.Path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Calculate relative depth from base path
		relPath, relErr := filepath.Rel(req.Path, path)
		if relErr != nil {
			return nil // Skip if we can't calculate relative path
		}

		// Skip root path itself
		if relPath == "." {
			return nil
		}

		// Calculate depth (number of path separators)
		depth := 0
		if relPath != "." {
			// Count separators in relative path
			depth = strings.Count(relPath, string(filepath.Separator))
			if info.IsDir() {
				// For directories, depth is the number of separators
			} else {
				// For files, depth is number of separators in parent
				depth = strings.Count(filepath.Dir(relPath), string(filepath.Separator))
			}
		}

		// Skip if it's a directory in our skip list
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir // Skip this directory and all its children
			}
			
			// Limit scan depth - don't go too deep (skip directories deeper than maxDepth)
			if depth >= maxDepth {
				return filepath.SkipDir // Skip deep directories
			}
			return nil
		}

		// Check for package.json (Node.js/React/Vue projects)
		if info.Name() == "package.json" {
			dir := filepath.Dir(path)
			
			// Skip if we've already processed this directory
			if visitedDirs[dir] {
				return nil
			}
			
			// Skip if too deep
			if depth >= maxDepth {
				return nil
			}
			
			// Skip if this package.json is inside a node_modules or other skipped directory
			dirRelPath, err := filepath.Rel(req.Path, dir)
			if err == nil {
				pathParts := strings.Split(dirRelPath, string(filepath.Separator))
				for _, part := range pathParts {
					if skipDirs[part] {
						return nil // Skip this package.json
					}
				}
			}
			
			visitedDirs[dir] = true
			
			// Read package.json to detect type
			content, err := os.ReadFile(path)
			if err == nil {
				var pkg map[string]interface{}
				if json.Unmarshal(content, &pkg) == nil {
					name := filepath.Base(dir)
					if n, ok := pkg["name"].(string); ok && n != "" {
						name = n
					}
					
					// Determine service type
					serviceType := "frontend"
					if deps, ok := pkg["dependencies"].(map[string]interface{}); ok {
						if _, hasReact := deps["react"]; hasReact {
							serviceType = "frontend"
						} else if _, hasVue := deps["vue"]; hasVue {
							serviceType = "frontend"
						} else if _, hasExpress := deps["express"]; hasExpress {
							serviceType = "backend"
						} else if _, hasNestjs := deps["@nestjs/core"]; hasNestjs {
							serviceType = "backend"
						}
					}
					
					// Check devDependencies too
					if serviceType == "frontend" {
						if devDeps, ok := pkg["devDependencies"].(map[string]interface{}); ok {
							if _, hasExpress := devDeps["express"]; hasExpress {
								serviceType = "backend"
							}
						}
					}

					command := "npm start"
					if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
						if _, hasStart := scripts["start"]; hasStart {
							command = "npm run start"
						} else if _, hasDev := scripts["dev"]; hasDev {
							command = "npm run dev"
						} else if _, hasServe := scripts["serve"]; hasServe {
							command = "npm run serve"
						}
					}

					// Try to detect port from various sources
					detectedPort := 0
					
					// 1. Check package.json scripts for PORT environment variable
					if scripts, ok := pkg["scripts"].(map[string]interface{}); ok {
						for _, scriptValue := range scripts {
							if scriptStr, ok := scriptValue.(string); ok {
								// Look for PORT=3000 or --port 3000 patterns
								if port := extractPortFromString(scriptStr); port > 0 {
									detectedPort = port
									break
								}
							}
						}
					}
					
					// 2. Check .env file in the same directory
					if detectedPort == 0 {
						envPath := filepath.Join(dir, ".env")
						if envPort := readPortFromEnvFile(envPath); envPort > 0 {
							detectedPort = envPort
						}
					}
					
					// 3. Check .env.local, .env.development, etc.
					if detectedPort == 0 {
						envFiles := []string{".env.local", ".env.development", ".env.production"}
						for _, envFile := range envFiles {
							envPath := filepath.Join(dir, envFile)
							if envPort := readPortFromEnvFile(envPath); envPort > 0 {
								detectedPort = envPort
								break
							}
						}
					}
					
					// 4. Check vite.config.js, next.config.js, etc.
					if detectedPort == 0 {
						configFiles := []string{"vite.config.js", "vite.config.ts", "next.config.js", "nuxt.config.js"}
						for _, configFile := range configFiles {
							configPath := filepath.Join(dir, configFile)
							if configPort := readPortFromConfigFile(configPath); configPort > 0 {
								detectedPort = configPort
								break
							}
						}
					}
					
					// 5. Default ports based on type (always set a default)
					switch serviceType {
					case "frontend":
						if detectedPort == 0 {
							detectedPort = 3000 // Default for React/Vue
						}
					case "backend":
						if detectedPort == 0 {
							detectedPort = 8000 // Default for Node.js backend
						}
					default:
						if detectedPort == 0 {
							detectedPort = 3000 // Default for other Node.js projects
						}
					}

					// Ensure port is always set (fallback to 3000 if still 0)
					if detectedPort == 0 {
						detectedPort = 3000
					}

					services = append(services, ServiceDetection{
						Name:        name,
						Type:        serviceType,
						Path:        dir,
						Command:     command,
						Port:        detectedPort,
						PackageFile: path,
					})
				}
			}
		}

		// Check for go.mod (Go projects)
		if info.Name() == "go.mod" {
			dir := filepath.Dir(path)
			
			// Skip if we've already processed this directory
			if visitedDirs[dir] {
				return nil
			}
			
			// Skip if this go.mod is inside a skipped directory
			relPath, err := filepath.Rel(req.Path, dir)
			if err == nil {
				pathParts := strings.Split(relPath, string(filepath.Separator))
				for _, part := range pathParts {
					if skipDirs[part] {
						return nil // Skip this go.mod
					}
				}
			}
			
			// Skip if too deep
			if depth >= maxDepth {
				return nil
			}
			
			visitedDirs[dir] = true
			
			content, err := os.ReadFile(path)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				name := "go-service"
				if len(lines) > 0 && strings.HasPrefix(lines[0], "module ") {
					moduleName := strings.TrimPrefix(lines[0], "module ")
					moduleName = strings.TrimSpace(moduleName)
					parts := strings.Split(moduleName, "/")
					if len(parts) > 0 {
						name = parts[len(parts)-1]
					}
				}

				// Check for main.go
				mainPath := filepath.Join(dir, "main.go")
				cmdPath := filepath.Join(dir, "cmd")
				var command string
				if _, err := os.Stat(mainPath); err == nil {
					command = "go run main.go"
				} else if _, err := os.Stat(cmdPath); err == nil {
					// Try to find main.go in cmd subdirectories
					cmdDirs, _ := os.ReadDir(cmdPath)
					for _, cmdDir := range cmdDirs {
						if cmdDir.IsDir() {
							mainFile := filepath.Join(cmdPath, cmdDir.Name(), "main.go")
							if _, err := os.Stat(mainFile); err == nil {
								command = fmt.Sprintf("go run ./cmd/%s/main.go", cmdDir.Name())
								break
							}
						}
					}
					if command == "" {
						command = "go run ./..."
					}
				} else {
					command = "go run ."
				}

				// Try to detect port for Go services
				detectedPort := 0
				
				// Check .env file
				envPath := filepath.Join(dir, ".env")
				if envPort := readPortFromEnvFile(envPath); envPort > 0 {
					detectedPort = envPort
				}
				
				// Check main.go or config files for port
				if detectedPort == 0 {
					mainPath := filepath.Join(dir, "main.go")
					if mainPort := readPortFromGoFile(mainPath); mainPort > 0 {
						detectedPort = mainPort
					}
				}
				
				// Default port for Go backend
				if detectedPort == 0 {
					detectedPort = 8080
				}

				services = append(services, ServiceDetection{
					Name:        name,
					Type:        "backend",
					Path:        dir,
					Command:     command,
					Port:        detectedPort,
					PackageFile: path,
				})
			}
		}

		// Check for requirements.txt (Python projects)
		if info.Name() == "requirements.txt" {
			dir := filepath.Dir(path)
			
			// Skip if we've already processed this directory
			if visitedDirs[dir] {
				return nil
			}
			
			// Skip if this requirements.txt is inside a skipped directory
			relPath, err := filepath.Rel(req.Path, dir)
			if err == nil {
				pathParts := strings.Split(relPath, string(filepath.Separator))
				for _, part := range pathParts {
					if skipDirs[part] {
						return nil // Skip this requirements.txt
					}
				}
			}
			
			// Skip if too deep
			if depth >= maxDepth {
				return nil
			}
			
			visitedDirs[dir] = true
			
			name := filepath.Base(dir)
			
			// Check for common Python frameworks
			command := "python app.py"
			if _, err := os.Stat(filepath.Join(dir, "manage.py")); err == nil {
				command = "python manage.py runserver"
				name = "django-service"
			} else if _, err := os.Stat(filepath.Join(dir, "main.py")); err == nil {
				command = "python main.py"
			} else if _, err := os.Stat(filepath.Join(dir, "app.py")); err == nil {
				command = "python app.py"
			}

			// Try to detect port for Python services
			detectedPort := 0
			
			// Check .env file
			envPath := filepath.Join(dir, ".env")
			if envPort := readPortFromEnvFile(envPath); envPort > 0 {
				detectedPort = envPort
			}
			
			// Default port for Python backend
			if detectedPort == 0 {
				detectedPort = 8000
			}

			services = append(services, ServiceDetection{
				Name:        name,
				Type:        "backend",
				Path:        dir,
				Command:     command,
				Port:        detectedPort,
				PackageFile: path,
			})
		}

		return nil
	})

	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to scan path", err.Error()))
		return
	}

	// If no services found, create a default one for the root path
	if len(services) == 0 {
		services = append(services, ServiceDetection{
			Name:    filepath.Base(req.Path),
			Type:    "other",
			Path:    req.Path,
			Command: "",
			Port:    3000, // Default port for unknown services
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": services,
	})
}

// Helper functions for port detection

// extractPortFromString extracts port number from a string (e.g., "PORT=3000", "--port 3000", ":3000")
func extractPortFromString(s string) int {
	// Look for PORT=3000 pattern
	if matches := strings.Split(s, "PORT="); len(matches) > 1 {
		portStr := strings.Fields(matches[1])[0]
		if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
			return port
		}
	}
	
	// Look for --port 3000 or -p 3000 pattern
	parts := strings.Fields(s)
	for i, part := range parts {
		if (part == "--port" || part == "-p") && i+1 < len(parts) {
			if port, err := strconv.Atoi(parts[i+1]); err == nil && port > 0 && port < 65536 {
				return port
			}
		}
	}
	
	// Look for :3000 pattern (e.g., "listen(3000)" or ":3000")
	re := strings.NewReplacer(":", " ", "(", " ", ")", " ", ",", " ")
	normalized := re.Replace(s)
	parts = strings.Fields(normalized)
	for _, part := range parts {
		if port, err := strconv.Atoi(part); err == nil && port > 1000 && port < 65536 {
			return port
		}
	}
	
	return 0
}

// readPortFromEnvFile reads PORT from .env file
func readPortFromEnvFile(envPath string) int {
	content, err := os.ReadFile(envPath)
	if err != nil {
		return 0
	}
	
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PORT=") {
			portStr := strings.TrimPrefix(line, "PORT=")
			portStr = strings.TrimSpace(portStr)
			if port, err := strconv.Atoi(portStr); err == nil && port > 0 && port < 65536 {
				return port
			}
		}
	}
	
	return 0
}

// readPortFromConfigFile reads port from JS/TS config files (vite.config.js, next.config.js, etc.)
func readPortFromConfigFile(configPath string) int {
	content, err := os.ReadFile(configPath)
	if err != nil {
		return 0
	}
	
	contentStr := string(content)
	
	// Look for port: 3000 or port:3000
	if port := extractPortFromString(contentStr); port > 0 {
		return port
	}
	
	// Look for server: { port: 3000 }
	lines := strings.Split(contentStr, "\n")
	for i, line := range lines {
		if strings.Contains(line, "port") && i+1 < len(lines) {
			// Check current and next line
			combined := line + " " + lines[i+1]
			if port := extractPortFromString(combined); port > 0 {
				return port
			}
		}
	}
	
	return 0
}

// readPortFromGoFile reads port from Go main.go file
func readPortFromGoFile(mainPath string) int {
	content, err := os.ReadFile(mainPath)
	if err != nil {
		return 0
	}
	
	contentStr := string(content)
	
	// Look for Listen(":8080") or ListenAndServe(":8080", nil)
	// Or port = 8080
	return extractPortFromString(contentStr)
}
