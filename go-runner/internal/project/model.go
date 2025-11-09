package project

import (
	"go-runner/internal/types"
	"time"

	"gorm.io/gorm"
)

// Re-export types for convenience
type ServiceStatus = types.ServiceStatus
type ServiceType = types.ServiceType

const (
	StatusStopped  = types.StatusStopped
	StatusStarting = types.StatusStarting
	StatusRunning  = types.StatusRunning
	StatusStopping = types.StatusStopping
	StatusError    = types.StatusError
	StatusUnknown  = types.StatusUnknown
)

const (
	TypeBackend  = types.TypeBackend
	TypeFrontend = types.TypeFrontend
	TypeWorker   = types.TypeWorker
	TypeDatabase = types.TypeDatabase
	TypeQueue    = types.TypeQueue
	TypeOther    = types.TypeOther
)

// ProjectGroup represents a group of related microservices
type ProjectGroup struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	Name        string `json:"name" gorm:"not null" binding:"required"`
	Description string `json:"description"`
	Color       string `json:"color"` // Hex color for UI
	Projects    []Project `json:"projects" gorm:"foreignKey:GroupID"`
}

// Project represents a microservice/application
type Project struct {
	ID          uint           `json:"id" gorm:"primarykey"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
	
	// Basic info
	Name        string `json:"name" gorm:"not null" binding:"required"`
	Description string `json:"description"`
	Type        ServiceType `json:"type" gorm:"default:'other'"`
	GroupID     *uint  `json:"group_id"`
	Group       *ProjectGroup `json:"group" gorm:"foreignKey:GroupID"`
	
	// Path and execution
	Path        string `json:"path" gorm:"not null" binding:"required"`
	Command     string `json:"command"` // Command to start the service
	Args        string `json:"args"`    // Additional arguments
	WorkingDir  string `json:"working_dir"` // Working directory
	
	// Network and ports
	Port        int    `json:"port"`
	Ports       string `json:"ports"` // JSON array of ports for complex services
	
	// Environment and configuration
	Environment string `json:"environment"` // development, staging, production
	EnvFile     string `json:"env_file"`    // Path to .env file
	EnvVars     string `json:"env_vars"`    // JSON object of environment variables
	
	// IDE and development
	Editor      string `json:"editor"`      // VSCode, IntelliJ, etc.
	EditorArgs  string `json:"editor_args"` // Additional editor arguments
	
	// Service management
	Status      ServiceStatus `json:"status" gorm:"default:'stopped'"`
	PID         int           `json:"pid"`         // Process ID when running
	StartTime   *time.Time    `json:"start_time"`  // When service started
	StopTime    *time.Time    `json:"stop_time"`   // When service stopped
	LastError   string        `json:"last_error"`  // Last error message
	
	// Health check
	HealthCheckURL string `json:"health_check_url"` // URL for health checks
	HealthStatus   string `json:"health_status"`    // healthy, unhealthy, unknown
	
	// Auto-restart settings
	AutoRestart bool `json:"auto_restart" gorm:"default:false"`
	RestartCount int  `json:"restart_count" gorm:"default:0"`
	MaxRestarts  int  `json:"max_restarts" gorm:"default:3"`
	
	// Resource limits
	CPULimit    string `json:"cpu_limit"`    // CPU limit (e.g., "500m")
	MemoryLimit string `json:"memory_limit"` // Memory limit (e.g., "512Mi")
	
	// Logs storage (JSON array of log lines, last 1000 lines)
	Logs string `json:"logs" gorm:"type:text"` // JSON array of log lines
}

// CreateProjectRequest represents the request to create a new project
type CreateProjectRequest struct {
	Name           string      `json:"name" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	Description    string      `json:"description" binding:"max=500" validate:"max=500"`
	Type           ServiceType `json:"type" binding:"oneof=backend frontend worker database queue other" validate:"oneof=backend frontend worker database queue other"`
	GroupID        *uint       `json:"group_id" validate:"omitempty,min=1"`
	Path           string      `json:"path" binding:"required" validate:"required,min=1"`
	Command        string      `json:"command" validate:"max=500"`
	Args           string      `json:"args" validate:"max=500"`
	WorkingDir     string      `json:"working_dir" validate:"max=500"`
	Port           int         `json:"port" binding:"min=1,max=65535" validate:"port"`
	Ports          string      `json:"ports" validate:"max=200"`
	Environment    string      `json:"environment" binding:"oneof=development staging production" validate:"oneof=development staging production"`
	EnvFile        string      `json:"env_file" validate:"max=500"`
	EnvVars        string      `json:"env_vars" validate:"max=2000"`
	Editor         string      `json:"editor" validate:"max=50"`
	EditorArgs     string      `json:"editor_args" validate:"max=500"`
	HealthCheckURL string      `json:"health_check_url" binding:"omitempty,url" validate:"omitempty,url"`
	AutoRestart    bool        `json:"auto_restart"`
	MaxRestarts    int         `json:"max_restarts" binding:"min=0,max=10" validate:"min=0,max=10"`
	CPULimit       string      `json:"cpu_limit" validate:"max=20"`
	MemoryLimit    string      `json:"memory_limit" validate:"max=20"`
}

// UpdateProjectRequest represents the request to update a project
type UpdateProjectRequest struct {
	Name           *string      `json:"name"`
	Description    *string      `json:"description"`
	Type           *ServiceType `json:"type"`
	GroupID        *uint        `json:"group_id"`
	Path           *string      `json:"path"`
	Command        *string      `json:"command"`
	Args           *string      `json:"args"`
	WorkingDir     *string      `json:"working_dir"`
	Port           *int         `json:"port"`
	Ports          *string      `json:"ports"`
	Environment    *string      `json:"environment"`
	EnvFile        *string      `json:"env_file"`
	EnvVars        *string      `json:"env_vars"`
	Editor         *string      `json:"editor"`
	EditorArgs     *string      `json:"editor_args"`
	HealthCheckURL *string      `json:"health_check_url"`
	AutoRestart    *bool        `json:"auto_restart"`
	MaxRestarts    *int         `json:"max_restarts"`
	CPULimit       *string      `json:"cpu_limit"`
	MemoryLimit    *string      `json:"memory_limit"`
}

// CreateProjectGroupRequest represents the request to create a new project group
type CreateProjectGroupRequest struct {
	Name        string `json:"name" binding:"required,min=1,max=100" validate:"required,min=1,max=100"`
	Description string `json:"description" binding:"max=500" validate:"max=500"`
	Color       string `json:"color" binding:"omitempty,hexcolor" validate:"omitempty,hexcolor"`
}

// UpdateProjectGroupRequest represents the request to update a project group
type UpdateProjectGroupRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	Color       *string `json:"color"`
}