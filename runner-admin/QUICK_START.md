# Quick Start Guide

## 🚀 Cài đặt nhanh

### 1. Cài đặt dependencies

```bash
cd D:/runner-admin
npm install
```

### 2. Tạo file .env

Tạo file `.env` trong thư mục `D:/runner-admin`:

```env
VITE_API_URL=http://localhost:8080
VITE_DEBUG=false
```

### 3. Khởi động Go Runner API

```bash
cd D:/go-runner
go run cmd/hotreload/main.go
```

### 4. Khởi động React Dashboard

```bash
cd D:/runner-admin
npm run dev
```

### 5. Truy cập Dashboard

Mở browser: `http://localhost:5173`

## 📊 Tính năng có sẵn

### System Overview

- Real-time system metrics
- CPU, Memory, Disk, Network usage
- System status indicators
- Process count và uptime

### System Alerts

- Active alerts với filtering
- Alert levels (info, warning, error, critical)
- Real-time updates

### System Metrics

- Historical charts
- Time range selection
- Export functionality

### Top Processes

- Process resource usage
- CPU/Memory sorting
- Process details

## 🔧 Troubleshooting

### Lỗi API Connection

- Kiểm tra Go Runner API có chạy không
- Kiểm tra VITE_API_URL trong .env
- Kiểm tra CORS settings

### Lỗi Icons

- Đã sử dụng lucide-react thay vì antd icons
- Tất cả icons đã được fix

### Lỗi Build

- Chạy `npm run build` để kiểm tra
- Kiểm tra TypeScript errors

## 📁 Cấu trúc

```
src/
├── components/system/     # System monitoring components
├── pages/system/         # System monitoring pages
├── api/                  # API clients
├── hooks/queries/        # React Query hooks
├── types/                # TypeScript types
└── utils/                # Utility functions
```

## 🎯 Next Steps

1. Tùy chỉnh dashboard theo nhu cầu
2. Thêm authentication nếu cần
3. Deploy lên production
4. Thêm real-time WebSocket updates
