package system

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"gorm.io/gorm"
)

// Service handles system monitoring business logic
type Service struct {
	db       *gorm.DB
	detector *Detector
	config   *SystemConfig
}

// NewService creates a new system service
func NewService(db *gorm.DB) *Service {
	service := &Service{
		db:       db,
		detector: NewDetector(),
	}

	// Load configuration
	service.loadConfig()

	// Start background tasks
	go service.startMetricsCollector()
	go service.startAlertChecker()

	return service
}

// loadConfig loads system configuration
func (s *Service) loadConfig() {
	var config SystemConfig
	if err := s.db.First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create default configuration
			config = SystemConfig{
				CPULimit:      80.0,
				MemoryLimit:   80.0,
				DiskLimit:     85.0,
				NetworkLimit:  100.0,
				CheckInterval: 60,
				RetentionDays: 30,
				EnableAlerts:  true,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			s.db.Create(&config)
		} else {
			log.Printf("Failed to load system config: %v", err)
			// Use default config
			config = SystemConfig{
				CPULimit:      80.0,
				MemoryLimit:   80.0,
				DiskLimit:     85.0,
				NetworkLimit:  100.0,
				CheckInterval: 60,
				RetentionDays: 30,
				EnableAlerts:  true,
			}
		}
	}
	s.config = &config
}

// startMetricsCollector starts the metrics collection background task
func (s *Service) startMetricsCollector() {
	ticker := time.NewTicker(time.Duration(s.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.collectMetrics()
	}
}

// collectMetrics collects and stores system metrics
func (s *Service) collectMetrics() {
	info, err := s.detector.GetSystemInfo()
	if err != nil {
		log.Printf("Failed to collect system metrics: %v", err)
		return
	}

	// Create metrics record
	metrics := SystemMetrics{
		Timestamp:  time.Now(),
		CPUUsage:   info.CPU.Usage,
		MemoryUsage: info.Memory.Usage,
		DiskUsage:  info.Disk.Usage,
		LoadAvg1:   info.CPU.LoadAverage[0],
		LoadAvg5:   info.CPU.LoadAverage[1],
		LoadAvg15:  info.CPU.LoadAverage[2],
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	// Store metrics
	if err := s.db.Create(&metrics).Error; err != nil {
		log.Printf("Failed to store system metrics: %v", err)
		return
	}

	// Clean up old metrics
	s.cleanupOldMetrics()
}

// cleanupOldMetrics removes old metrics based on retention policy
func (s *Service) cleanupOldMetrics() {
	cutoffTime := time.Now().Add(-time.Duration(s.config.RetentionDays) * 24 * time.Hour)
	
	result := s.db.Where("timestamp < ?", cutoffTime).Delete(&SystemMetrics{})
	if result.Error != nil {
		log.Printf("Failed to cleanup old metrics: %v", result.Error)
	} else {
		log.Printf("Cleaned up %d old metrics", result.RowsAffected)
	}
}

// startAlertChecker starts the alert checking background task
func (s *Service) startAlertChecker() {
	if !s.config.EnableAlerts {
		return
	}

	ticker := time.NewTicker(time.Duration(s.config.CheckInterval) * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		s.checkAlerts()
	}
}

// checkAlerts checks for system alerts
func (s *Service) checkAlerts() {
	info, err := s.detector.GetSystemInfo()
	if err != nil {
		log.Printf("Failed to check alerts: %v", err)
		return
	}

	// Check CPU alert
	s.checkCPUAlert(info.CPU.Usage)

	// Check memory alert
	s.checkMemoryAlert(info.Memory.Usage)

	// Check disk alert
	s.checkDiskAlert(info.Disk.Usage)

	// Check load average alert
	if len(info.CPU.LoadAverage) > 0 {
		s.checkLoadAlert(info.CPU.LoadAverage[0])
	}
}

// checkCPUAlert checks for CPU usage alerts
func (s *Service) checkCPUAlert(usage float64) {
	if usage >= s.config.CPULimit {
		level := "warning"
		if usage >= 95 {
			level = "critical"
		}

		alert := SystemAlert{
			Type:      "cpu",
			Level:     level,
			Message:   fmt.Sprintf("CPU usage is %.2f%% (threshold: %.2f%%)", usage, s.config.CPULimit),
			Value:     usage,
			Threshold: s.config.CPULimit,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.createAlert(&alert)
	}
}

// checkMemoryAlert checks for memory usage alerts
func (s *Service) checkMemoryAlert(usage float64) {
	if usage >= s.config.MemoryLimit {
		level := "warning"
		if usage >= 95 {
			level = "critical"
		}

		alert := SystemAlert{
			Type:      "memory",
			Level:     level,
			Message:   fmt.Sprintf("Memory usage is %.2f%% (threshold: %.2f%%)", usage, s.config.MemoryLimit),
			Value:     usage,
			Threshold: s.config.MemoryLimit,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.createAlert(&alert)
	}
}

// checkDiskAlert checks for disk usage alerts
func (s *Service) checkDiskAlert(usage float64) {
	if usage >= s.config.DiskLimit {
		level := "warning"
		if usage >= 95 {
			level = "critical"
		}

		alert := SystemAlert{
			Type:      "disk",
			Level:     level,
			Message:   fmt.Sprintf("Disk usage is %.2f%% (threshold: %.2f%%)", usage, s.config.DiskLimit),
			Value:     usage,
			Threshold: s.config.DiskLimit,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.createAlert(&alert)
	}
}

// checkLoadAlert checks for load average alerts
func (s *Service) checkLoadAlert(load1 float64) {
	// Simple load alert based on CPU count
	cpuCount := runtime.NumCPU()
	threshold := float64(cpuCount) * 2.0 // Alert if load > 2x CPU count

	if load1 >= threshold {
		level := "warning"
		if load1 >= float64(cpuCount)*4.0 {
			level = "critical"
		}

		alert := SystemAlert{
			Type:      "load",
			Level:     level,
			Message:   fmt.Sprintf("Load average is %.2f (threshold: %.2f)", load1, threshold),
			Value:     load1,
			Threshold: threshold,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		s.createAlert(&alert)
	}
}

// createAlert creates a new alert if it doesn't already exist
func (s *Service) createAlert(alert *SystemAlert) {
	// Check if similar alert already exists
	var existingAlert SystemAlert
	err := s.db.Where("type = ? AND level = ? AND is_active = ? AND created_at > ?", 
		alert.Type, alert.Level, true, time.Now().Add(-5*time.Minute)).First(&existingAlert).Error

	if err == gorm.ErrRecordNotFound {
		// Create new alert
		if err := s.db.Create(alert).Error; err != nil {
			log.Printf("Failed to create alert: %v", err)
		} else {
			log.Printf("Created %s alert: %s", alert.Level, alert.Message)
			// TODO: Send notification (email, webhook, etc.)
		}
	}
}

// UpdateConfig updates system configuration
func (s *Service) UpdateConfig(config *SystemConfig) error {
	config.UpdatedAt = time.Now()
	if err := s.db.Save(config).Error; err != nil {
		return err
	}
	s.config = config
	return nil
}

// GetConfig returns current system configuration
func (s *Service) GetConfig() *SystemConfig {
	return s.config
}

// GetMetrics returns system metrics with pagination
func (s *Service) GetMetrics(page, limit int, hours int) ([]SystemMetrics, int64, error) {
	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)
	
	var metrics []SystemMetrics
	var total int64

	query := s.db.Model(&SystemMetrics{}).Where("timestamp >= ?", startTime)
	
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("timestamp DESC").Offset(offset).Limit(limit).Find(&metrics).Error; err != nil {
		return nil, 0, err
	}

	return metrics, total, nil
}

// GetAlerts returns system alerts with filtering
func (s *Service) GetAlerts(alertType, level string, activeOnly bool, page, limit int) ([]SystemAlert, int64, error) {
	query := s.db.Model(&SystemAlert{})

	if alertType != "" {
		query = query.Where("type = ?", alertType)
	}
	if level != "" {
		query = query.Where("level = ?", level)
	}
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	var alerts []SystemAlert
	var total int64

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Order("created_at DESC").Offset(offset).Limit(limit).Find(&alerts).Error; err != nil {
		return nil, 0, err
	}

	return alerts, total, nil
}

// ResolveAlert resolves an alert
func (s *Service) ResolveAlert(alertID uint) error {
	now := time.Now()
	return s.db.Model(&SystemAlert{}).Where("id = ?", alertID).Updates(map[string]interface{}{
		"is_active":   false,
		"resolved_at": &now,
		"updated_at":  now,
	}).Error
}

// ClearOldMetrics clears metrics older than specified days
func (s *Service) ClearOldMetrics(days int) (int64, error) {
	cutoffTime := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	result := s.db.Where("timestamp < ?", cutoffTime).Delete(&SystemMetrics{})
	return result.RowsAffected, result.Error
}
