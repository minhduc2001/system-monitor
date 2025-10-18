package project

import (
	"net/http"
	"strconv"

	"go-runner/internal/middleware"
	"go-runner/internal/service"
	"go-runner/internal/websocket"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db      *gorm.DB
	manager *service.Manager
	hub     *websocket.Hub
}

func NewHandler(db *gorm.DB, manager *service.Manager, hub *websocket.Hub) *Handler {
	return &Handler{
		db:      db,
		manager: manager,
		hub:     hub,
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
		projects.GET("/:id/status", h.GetProjectStatus)
		projects.GET("/:id/logs", h.GetLogs)
		projects.GET("/:id/logs/ws", h.StreamLogs)
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

	var project Project
	if err := h.db.First(&project, id).Error; err != nil {
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

	if err := h.manager.StartService(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast status update via WebSocket
	h.hub.BroadcastToProject(uint(id), "status_update", gin.H{
		"project_id": id,
		"status":     "starting",
		"message":    "Project is starting...",
	})

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

	// For now, return empty logs. In a real implementation, you would
	// read from log files or database
	c.JSON(http.StatusOK, gin.H{"message": "Logs retrieved successfully", "project_id": id, "logs": []string{}})
}

func (h *Handler) StreamLogs(c *gin.Context) {
	h.hub.HandleWebSocket(c)
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
