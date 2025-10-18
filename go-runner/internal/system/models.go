package system

import (
	"time"
)

// SystemInfo represents overall system information
type SystemInfo struct {
	Hostname    string      `json:"hostname"`
	Platform    string      `json:"platform"`
	Architecture string     `json:"architecture"`
	GoVersion   string      `json:"go_version"`
	Uptime      time.Duration `json:"uptime"`
	CPU         CPUInfo     `json:"cpu"`
	Memory      MemoryInfo  `json:"memory"`
	Disk        DiskInfo    `json:"disk"`
	Network     NetworkInfo `json:"network"`
	Processes   []ProcessInfo `json:"processes"`
	Timestamp   time.Time   `json:"timestamp"`
}

// CPUInfo represents CPU usage information
type CPUInfo struct {
	Usage       float64 `json:"usage"`        // CPU usage percentage
	Count       int     `json:"count"`        // Number of CPU cores
	ModelName   string  `json:"model_name"`   // CPU model name
	Mhz         float64 `json:"mhz"`          // CPU frequency in MHz
	LoadAverage []float64 `json:"load_avg"`   // Load average (1min, 5min, 15min)
}

// MemoryInfo represents memory usage information
type MemoryInfo struct {
	Total       uint64  `json:"total"`        // Total memory in bytes
	Available   uint64  `json:"available"`    // Available memory in bytes
	Used        uint64  `json:"used"`         // Used memory in bytes
	Free        uint64  `json:"free"`         // Free memory in bytes
	Usage       float64 `json:"usage"`        // Memory usage percentage
	SwapTotal   uint64  `json:"swap_total"`   // Total swap in bytes
	SwapUsed    uint64  `json:"swap_used"`    // Used swap in bytes
	SwapFree    uint64  `json:"swap_free"`    // Free swap in bytes
	SwapUsage   float64 `json:"swap_usage"`   // Swap usage percentage
}

// DiskInfo represents disk usage information
type DiskInfo struct {
	Total       uint64  `json:"total"`        // Total disk space in bytes
	Used        uint64  `json:"used"`         // Used disk space in bytes
	Free        uint64  `json:"free"`         // Free disk space in bytes
	Usage       float64 `json:"usage"`        // Disk usage percentage
	InodesTotal uint64  `json:"inodes_total"` // Total inodes
	InodesUsed  uint64  `json:"inodes_used"`  // Used inodes
	InodesFree  uint64  `json:"inodes_free"`  // Free inodes
	InodesUsage float64 `json:"inodes_usage"` // Inodes usage percentage
}

// NetworkInfo represents network interface information
type NetworkInfo struct {
	Interfaces []NetworkInterface `json:"interfaces"`
	TotalBytesSent     uint64 `json:"total_bytes_sent"`
	TotalBytesReceived uint64 `json:"total_bytes_received"`
	TotalPacketsSent   uint64 `json:"total_packets_sent"`
	TotalPacketsReceived uint64 `json:"total_packets_received"`
}

// NetworkInterface represents a network interface
type NetworkInterface struct {
	Name         string `json:"name"`
	MTU          int    `json:"mtu"`
	HardwareAddr string `json:"hardware_addr"`
	Flags        string `json:"flags"`
	Addrs        []string `json:"addrs"`
	BytesSent    uint64 `json:"bytes_sent"`
	BytesReceived uint64 `json:"bytes_received"`
	PacketsSent  uint64 `json:"packets_sent"`
	PacketsReceived uint64 `json:"packets_received"`
	ErrorsIn     uint64 `json:"errors_in"`
	ErrorsOut    uint64 `json:"errors_out"`
	DropIn       uint64 `json:"drop_in"`
	DropOut      uint64 `json:"drop_out"`
}

// ProcessInfo represents process information
type ProcessInfo struct {
	PID         int32   `json:"pid"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	CPUPercent  float64 `json:"cpu_percent"`
	MemoryPercent float64 `json:"memory_percent"`
	MemoryRSS   uint64  `json:"memory_rss"`   // Resident Set Size
	MemoryVMS   uint64  `json:"memory_vms"`   // Virtual Memory Size
	CreateTime  int64   `json:"create_time"`
	Username    string  `json:"username"`
	Command     string  `json:"command"`
}

// SystemMetrics represents historical system metrics
type SystemMetrics struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Timestamp time.Time `json:"timestamp"`
	CPUUsage  float64   `json:"cpu_usage"`
	MemoryUsage float64 `json:"memory_usage"`
	DiskUsage float64   `json:"disk_usage"`
	LoadAvg1  float64   `json:"load_avg_1"`
	LoadAvg5  float64   `json:"load_avg_5"`
	LoadAvg15 float64   `json:"load_avg_15"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SystemAlert represents system alerts
type SystemAlert struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Type        string    `json:"type"`        // cpu, memory, disk, network
	Level       string    `json:"level"`       // info, warning, error, critical
	Message     string    `json:"message"`
	Value       float64   `json:"value"`
	Threshold   float64   `json:"threshold"`
	IsActive    bool      `json:"is_active"`
	ResolvedAt  *time.Time `json:"resolved_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// SystemConfig represents system monitoring configuration
type SystemConfig struct {
	ID                    uint    `json:"id" gorm:"primaryKey"`
	CPULimit              float64 `json:"cpu_limit"`               // CPU usage threshold (%)
	MemoryLimit           float64 `json:"memory_limit"`            // Memory usage threshold (%)
	DiskLimit             float64 `json:"disk_limit"`              // Disk usage threshold (%)
	NetworkLimit          float64 `json:"network_limit"`           // Network usage threshold (Mbps)
	CheckInterval         int     `json:"check_interval"`          // Check interval in seconds
	RetentionDays         int     `json:"retention_days"`          // Metrics retention in days
	EnableAlerts          bool    `json:"enable_alerts"`           // Enable alerting
	AlertEmail            string  `json:"alert_email"`             // Alert email address
	AlertWebhook          string  `json:"alert_webhook"`           // Alert webhook URL
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

// SystemStatus represents current system status
type SystemStatus struct {
	Status      string    `json:"status"`      // healthy, warning, critical
	Message     string    `json:"message"`
	LastCheck   time.Time `json:"last_check"`
	Uptime      time.Duration `json:"uptime"`
	CPUStatus   string    `json:"cpu_status"`
	MemoryStatus string   `json:"memory_status"`
	DiskStatus  string    `json:"disk_status"`
	NetworkStatus string  `json:"network_status"`
	ActiveAlerts int      `json:"active_alerts"`
}
