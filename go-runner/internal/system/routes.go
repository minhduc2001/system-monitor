package system

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RegisterRoutes registers system monitoring routes
func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB) {
	handler := NewHandler(db)
	
	// System information routes
	system := r.Group("/system")
	{
		// Basic system info
		system.GET("/info", handler.GetSystemInfo)
		system.GET("/status", handler.GetSystemStatus)
		system.GET("/dashboard", handler.GetSystemDashboard)
		
		// Metrics and monitoring
		system.GET("/metrics", handler.GetSystemMetrics)
		system.POST("/metrics/cleanup", handler.ClearOldMetrics)
		
		// Alerts
		system.GET("/alerts", handler.GetSystemAlerts)
		
		// Configuration
		system.GET("/config", handler.GetSystemConfig)
		system.PUT("/config", handler.UpdateSystemConfig)
	}
}
