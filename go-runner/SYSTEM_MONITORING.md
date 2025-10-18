# System Monitoring

## Tổng quan

Go Runner đã được tích hợp hệ thống monitoring toàn diện để theo dõi tình trạng hệ thống, hiệu suất và cảnh báo.

## Tính năng chính

### 📊 **System Information**

- **CPU**: Usage, cores, model, frequency, load average
- **Memory**: Total, used, free, swap usage
- **Disk**: Space usage, inodes, I/O statistics
- **Network**: Interface details, traffic statistics
- **Processes**: Running processes, resource usage

### 📈 **Metrics Collection**

- Thu thập metrics tự động theo interval
- Lưu trữ historical data
- Cleanup tự động theo retention policy
- Real-time monitoring

### 🚨 **Alert System**

- CPU usage alerts
- Memory usage alerts
- Disk space alerts
- Load average alerts
- Email/webhook notifications

### 📋 **Dashboard**

- System overview
- Real-time metrics
- Active alerts
- Top processes
- Historical charts

## API Endpoints

### System Information

```
GET /api/v1/system/info          # Thông tin hệ thống chi tiết
GET /api/v1/system/status        # Trạng thái hệ thống
GET /api/v1/system/dashboard     # Dashboard tổng quan
```

### Metrics

```
GET /api/v1/system/metrics       # Historical metrics
POST /api/v1/system/metrics/cleanup  # Cleanup old metrics
```

### Alerts

```
GET /api/v1/system/alerts        # System alerts
```

### Configuration

```
GET /api/v1/system/config        # System config
PUT /api/v1/system/config        # Update config
```

## Cấu hình

### SystemConfig

```go
type SystemConfig struct {
    CPULimit       float64 `json:"cpu_limit"`        // CPU threshold (%)
    MemoryLimit    float64 `json:"memory_limit"`     // Memory threshold (%)
    DiskLimit      float64 `json:"disk_limit"`       // Disk threshold (%)
    NetworkLimit   float64 `json:"network_limit"`    // Network threshold (Mbps)
    CheckInterval  int     `json:"check_interval"`   // Check interval (seconds)
    RetentionDays  int     `json:"retention_days"`   // Data retention (days)
    EnableAlerts   bool    `json:"enable_alerts"`    // Enable alerting
    AlertEmail     string  `json:"alert_email"`      // Alert email
    AlertWebhook   string  `json:"alert_webhook"`    // Alert webhook
}
```

### Default Configuration

```json
{
  "cpu_limit": 80.0,
  "memory_limit": 80.0,
  "disk_limit": 85.0,
  "network_limit": 100.0,
  "check_interval": 60,
  "retention_days": 30,
  "enable_alerts": true
}
```

## Sử dụng

### 1. Xem System Information

```bash
curl http://localhost:8080/api/v1/system/info
```

Response:

```json
{
    "data": {
        "hostname": "server-01",
        "platform": "windows",
        "architecture": "amd64",
        "go_version": "go1.21.0",
        "uptime": "2h30m15s",
        "cpu": {
            "usage": 45.2,
            "count": 8,
            "model_name": "Intel(R) Core(TM) i7-8700K",
            "mhz": 3700.0,
            "load_avg": [1.2, 1.5, 1.8]
        },
        "memory": {
            "total": 17179869184,
            "available": 8589934592,
            "used": 8589934592,
            "free": 4294967296,
            "usage": 50.0,
            "swap_total": 4294967296,
            "swap_used": 0,
            "swap_free": 4294967296,
            "swap_usage": 0.0
        },
        "disk": {
            "total": 1000000000000,
            "used": 500000000000,
            "free": 500000000000,
            "usage": 50.0,
            "inodes_total": 1000000,
            "inodes_used": 500000,
            "inodes_free": 500000,
            "inodes_usage": 50.0
        },
        "network": {
            "interfaces": [...],
            "total_bytes_sent": 1000000,
            "total_bytes_received": 2000000
        },
        "processes": [...],
        "timestamp": "2025-10-18T11:00:00Z"
    }
}
```

### 2. Xem System Status

```bash
curl http://localhost:8080/api/v1/system/status
```

Response:

```json
{
  "data": {
    "status": "healthy",
    "message": "System is healthy",
    "last_check": "2025-10-18T11:00:00Z",
    "uptime": "2h30m15s",
    "cpu_status": "healthy",
    "memory_status": "healthy",
    "disk_status": "healthy",
    "network_status": "healthy",
    "active_alerts": 0
  }
}
```

### 3. Xem Dashboard

```bash
curl http://localhost:8080/api/v1/system/dashboard
```

### 4. Xem Metrics

```bash
# Lấy metrics 24h gần nhất
curl "http://localhost:8080/api/v1/system/metrics?hours=24&page=1&limit=100"

# Lấy metrics với pagination
curl "http://localhost:8080/api/v1/system/metrics?page=1&limit=50"
```

### 5. Xem Alerts

```bash
# Tất cả alerts
curl http://localhost:8080/api/v1/system/alerts

# Chỉ active alerts
curl "http://localhost:8080/api/v1/system/alerts?active=true"

# Filter theo type
curl "http://localhost:8080/api/v1/system/alerts?type=cpu&level=warning"
```

### 6. Cập nhật Configuration

```bash
curl -X PUT http://localhost:8080/api/v1/system/config \
  -H "Content-Type: application/json" \
  -d '{
    "cpu_limit": 85.0,
    "memory_limit": 85.0,
    "disk_limit": 90.0,
    "check_interval": 30,
    "retention_days": 7,
    "enable_alerts": true,
    "alert_email": "admin@example.com"
  }'
```

## Alert Levels

### 🟢 **Healthy**

- CPU usage < 70%
- Memory usage < 80%
- Disk usage < 85%
- Load average < 2x CPU cores

### 🟡 **Warning**

- CPU usage 70-90%
- Memory usage 80-90%
- Disk usage 85-95%
- Load average 2-4x CPU cores

### 🔴 **Critical**

- CPU usage > 90%
- Memory usage > 90%
- Disk usage > 95%
- Load average > 4x CPU cores

## Database Schema

### SystemMetrics

```sql
CREATE TABLE system_metrics (
    id INTEGER PRIMARY KEY,
    timestamp DATETIME,
    cpu_usage REAL,
    memory_usage REAL,
    disk_usage REAL,
    load_avg_1 REAL,
    load_avg_5 REAL,
    load_avg_15 REAL,
    created_at DATETIME,
    updated_at DATETIME
);
```

### SystemAlert

```sql
CREATE TABLE system_alerts (
    id INTEGER PRIMARY KEY,
    type VARCHAR(50),
    level VARCHAR(20),
    message TEXT,
    value REAL,
    threshold REAL,
    is_active BOOLEAN,
    resolved_at DATETIME,
    created_at DATETIME,
    updated_at DATETIME
);
```

### SystemConfig

```sql
CREATE TABLE system_configs (
    id INTEGER PRIMARY KEY,
    cpu_limit REAL,
    memory_limit REAL,
    disk_limit REAL,
    network_limit REAL,
    check_interval INTEGER,
    retention_days INTEGER,
    enable_alerts BOOLEAN,
    alert_email VARCHAR(255),
    alert_webhook VARCHAR(500),
    created_at DATETIME,
    updated_at DATETIME
);
```

## Background Services

### Metrics Collector

- Chạy mỗi `check_interval` giây
- Thu thập system metrics
- Lưu vào database
- Auto cleanup old data

### Alert Checker

- Chạy mỗi `check_interval` giây
- Kiểm tra thresholds
- Tạo alerts khi cần
- Gửi notifications

## Monitoring Best Practices

### 1. **Thresholds**

- CPU: 70% warning, 90% critical
- Memory: 80% warning, 90% critical
- Disk: 85% warning, 95% critical
- Load: 2x cores warning, 4x cores critical

### 2. **Retention**

- Metrics: 30 days default
- Alerts: Keep resolved alerts 7 days
- Logs: Rotate daily

### 3. **Alerting**

- Enable email notifications
- Setup webhook for Slack/Discord
- Test alert system regularly

### 4. **Performance**

- Monitor collector performance
- Adjust check_interval based on load
- Use pagination for large datasets

## Troubleshooting

### Metrics không được thu thập

1. Kiểm tra database connection
2. Kiểm tra system permissions
3. Xem logs của collector service

### Alerts không hoạt động

1. Kiểm tra `enable_alerts` = true
2. Kiểm tra thresholds configuration
3. Kiểm tra alert email/webhook settings

### Performance issues

1. Tăng `check_interval`
2. Giảm `retention_days`
3. Cleanup old metrics manually

## Integration với Swagger

Tất cả endpoints đều có Swagger documentation tại:

```
http://localhost:8080/swagger/index.html
```

## Dependencies

- `github.com/shirou/gopsutil/v3` - System information collection
- `gorm.io/gorm` - Database ORM
- `github.com/gin-gonic/gin` - HTTP framework

## Security

- System info chỉ accessible từ localhost
- Có thể thêm authentication middleware
- Rate limiting cho API endpoints
- Input validation cho configuration
