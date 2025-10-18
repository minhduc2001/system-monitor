package types

import (
	"context"
	"time"
)

// ServiceStatus represents the current status of a microservice
type ServiceStatus string

const (
	StatusStopped  ServiceStatus = "stopped"
	StatusStarting ServiceStatus = "starting"
	StatusRunning  ServiceStatus = "running"
	StatusStopping ServiceStatus = "stopping"
	StatusError    ServiceStatus = "error"
	StatusUnknown  ServiceStatus = "unknown"
)

// ServiceType represents the type of microservice
type ServiceType string

const (
	TypeBackend  ServiceType = "backend"
	TypeFrontend ServiceType = "frontend"
	TypeWorker   ServiceType = "worker"
	TypeDatabase ServiceType = "database"
	TypeQueue    ServiceType = "queue"
	TypeOther    ServiceType = "other"
)

// ProcessInfo holds information about a running process
type ProcessInfo struct {
	ProjectID uint
	Process   interface{} // *exec.Cmd - using interface to avoid circular import
	Context   context.Context
	Cancel    context.CancelFunc
	StartTime time.Time
	Logs      chan string
}

// ServiceStats represents runtime statistics for a service
type ServiceStats struct {
	ProjectID     uint      `json:"project_id"`
	CPUUsage      float64   `json:"cpu_usage"`
	MemoryUsage   int64     `json:"memory_usage"`
	Uptime        int64     `json:"uptime"` // in seconds
	RequestCount  int64     `json:"request_count"`
	ErrorCount    int64     `json:"error_count"`
	LastUpdated   time.Time `json:"last_updated"`
}
