package system

import (
	"net/http"
	"strconv"
	"time"

	"go-runner/internal/middleware"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler handles system monitoring requests
type Handler struct {
	db       *gorm.DB
	detector *Detector
}

// NewHandler creates a new system handler
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{
		db:       db,
		detector: NewDetector(),
	}
}

// GetSystemInfo godoc
// @Summary      Get system information
// @Description  Get comprehensive system information including CPU, memory, disk, and network
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "System information"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /system/info [get]
func (h *Handler) GetSystemInfo(c *gin.Context) {
	info, err := h.detector.GetSystemInfo()
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get system info", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": info,
	})
}

// GetSystemStatus godoc
// @Summary      Get system status
// @Description  Get current system status and health indicators
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "System status"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /system/status [get]
func (h *Handler) GetSystemStatus(c *gin.Context) {
	status, err := h.detector.GetSystemStatus()
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get system status", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": status,
	})
}

// GetSystemMetrics godoc
// @Summary      Get system metrics
// @Description  Get historical system metrics with pagination
// @Tags         system
// @Accept       json
// @Produce      json
// @Param        page     query     int  false  "Page number"
// @Param        limit    query     int  false  "Items per page"
// @Param        hours    query     int  false  "Hours of data to retrieve"
// @Success      200      {object}  map[string]interface{}  "System metrics"
// @Failure      400      {object}  map[string]interface{}  "Bad request"
// @Failure      500      {object}  map[string]interface{}  "Internal server error"
// @Router       /system/metrics [get]
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	hours, _ := strconv.Atoi(c.DefaultQuery("hours", "24"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 100
	}
	if hours < 1 {
		hours = 24
	}

	// Calculate time range
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	// Query metrics
	var metrics []SystemMetrics
	var total int64

	query := h.db.Model(&SystemMetrics{}).Where("timestamp >= ?", startTime)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to count metrics", err.Error()))
		return
	}

	// Get paginated results
	offset := (page - 1) * limit
	if err := query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&metrics).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get metrics", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": metrics,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSystemAlerts godoc
// @Summary      Get system alerts
// @Description  Get system alerts with filtering options
// @Tags         system
// @Accept       json
// @Produce      json
// @Param        type     query     string  false  "Alert type filter"
// @Param        level    query     string  false  "Alert level filter"
// @Param        active   query     bool    false  "Active alerts only"
// @Param        page     query     int     false  "Page number"
// @Param        limit    query     int     false  "Items per page"
// @Success      200      {object}  map[string]interface{}  "System alerts"
// @Failure      400      {object}  map[string]interface{}  "Bad request"
// @Failure      500      {object}  map[string]interface{}  "Internal server error"
// @Router       /system/alerts [get]
func (h *Handler) GetSystemAlerts(c *gin.Context) {
	// Parse query parameters
	alertType := c.Query("type")
	level := c.Query("level")
	activeStr := c.Query("active")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 1000 {
		limit = 50
	}

	// Build query
	query := h.db.Model(&SystemAlert{})

	if alertType != "" {
		query = query.Where("type = ?", alertType)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if activeStr == "true" {
		query = query.Where("is_active = ?", true)
	}

	// Get total count
	var total int64
	if err := query.Count(&total).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to count alerts", err.Error()))
		return
	}

	// Get paginated results
	var alerts []SystemAlert
	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&alerts).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get alerts", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": alerts,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// GetSystemConfig godoc
// @Summary      Get system configuration
// @Description  Get system monitoring configuration
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "System configuration"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /system/config [get]
func (h *Handler) GetSystemConfig(c *gin.Context) {
	var config SystemConfig
	if err := h.db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default config if none exists
			config = SystemConfig{
				CPULimit:      80.0,
				MemoryLimit:   80.0,
				DiskLimit:     85.0,
				NetworkLimit:  100.0,
				CheckInterval: 60,
				RetentionDays: 30,
				EnableAlerts:  true,
			}
		} else {
			middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get config", err.Error()))
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": config,
	})
}

// UpdateSystemConfig godoc
// @Summary      Update system configuration
// @Description  Update system monitoring configuration
// @Tags         system
// @Accept       json
// @Produce      json
// @Param        config  body      SystemConfig  true  "System configuration"
// @Success      200     {object}  map[string]interface{}  "Updated configuration"
// @Failure      400     {object}  map[string]interface{}  "Bad request"
// @Failure      500     {object}  map[string]interface{}  "Internal server error"
// @Router       /system/config [put]
func (h *Handler) UpdateSystemConfig(c *gin.Context) {
	var config SystemConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Invalid configuration", err.Error()))
		return
	}

	// Validate configuration
	if config.CPULimit < 0 || config.CPULimit > 100 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "CPU limit must be between 0 and 100", ""))
		return
	}
	if config.MemoryLimit < 0 || config.MemoryLimit > 100 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Memory limit must be between 0 and 100", ""))
		return
	}
	if config.DiskLimit < 0 || config.DiskLimit > 100 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Disk limit must be between 0 and 100", ""))
		return
	}
	if config.CheckInterval < 10 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Check interval must be at least 10 seconds", ""))
		return
	}
	if config.RetentionDays < 1 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Retention days must be at least 1", ""))
		return
	}

	// Update or create configuration
	config.UpdatedAt = time.Now()
	if err := h.db.Save(&config).Error; err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to update config", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": config,
		"message": "Configuration updated successfully",
	})
}

// GetSystemDashboard godoc
// @Summary      Get system dashboard
// @Description  Get system dashboard with overview information
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "System dashboard"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /system/dashboard [get]
func (h *Handler) GetSystemDashboard(c *gin.Context) {
	// Get system info
	info, err := h.detector.GetSystemInfo()
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get system info", err.Error()))
		return
	}

	// Get system status
	status, err := h.detector.GetSystemStatus()
	if err != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to get system status", err.Error()))
		return
	}

	// Get recent metrics (last 24 hours)
	var recentMetrics []SystemMetrics
	startTime := time.Now().Add(-24 * time.Hour)
	h.db.Where("timestamp >= ?", startTime).Order("timestamp DESC").Limit(100).Find(&recentMetrics)

	// Get active alerts
	var activeAlerts []SystemAlert
	h.db.Where("is_active = ?", true).Order("created_at DESC").Find(&activeAlerts)

	// Get top processes by CPU usage
	var topProcesses []ProcessInfo
	for _, p := range info.Processes {
		if len(topProcesses) < 10 {
			topProcesses = append(topProcesses, p)
		}
	}

	dashboard := gin.H{
		"system_info": info,
		"system_status": status,
		"recent_metrics": recentMetrics,
		"active_alerts": activeAlerts,
		"top_processes": topProcesses,
		"timestamp": time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dashboard,
	})
}

// ClearOldMetrics godoc
// @Summary      Clear old metrics
// @Description  Clear metrics older than specified days
// @Tags         system
// @Accept       json
// @Produce      json
// @Param        days  query     int  false  "Days to keep (default: 30)"
// @Success      200   {object}  map[string]interface{}  "Cleanup result"
// @Failure      400   {object}  map[string]interface{}  "Bad request"
// @Failure      500   {object}  map[string]interface{}  "Internal server error"
// @Router       /system/metrics/cleanup [post]
func (h *Handler) ClearOldMetrics(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	
	if days < 1 {
		middleware.HandleError(c, middleware.NewError(http.StatusBadRequest, "Days must be at least 1", ""))
		return
	}

	cutoffTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	
	result := h.db.Where("timestamp < ?", cutoffTime).Delete(&SystemMetrics{})
	if result.Error != nil {
		middleware.HandleError(c, middleware.NewError(http.StatusInternalServerError, "Failed to clear old metrics", result.Error.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Old metrics cleared successfully",
		"deleted_count": result.RowsAffected,
		"cutoff_time": cutoffTime,
	})
}
