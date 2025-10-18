# Installation Guide

## Prerequisites

- Node.js 18+
- npm hoặc yarn
- Go Runner API đang chạy

## Cài đặt Dependencies

```bash
# Cài đặt dependencies chính
npm install

# Cài đặt dependencies cho system monitoring
npm install @ant-design/plots dayjs
```

## Cấu hình Environment

Tạo file `.env` trong thư mục `D:/runner-admin`:

```env
VITE_API_URL=http://localhost:8080
VITE_DEBUG=false
```

## Khởi động Development

### 1. Khởi động Go Runner API

```bash
cd D:/go-runner
go run cmd/hotreload/main.go
```

### 2. Khởi động React Dashboard

```bash
cd D:/runner-admin
npm run dev
```

### 3. Truy cập Dashboard

Mở browser và truy cập: `http://localhost:5173`

## Build Production

```bash
npm run build
```

## Troubleshooting

### Lỗi API Connection

- Kiểm tra Go Runner API có đang chạy không
- Kiểm tra VITE_API_URL trong .env
- Kiểm tra CORS settings trong Go API

### Lỗi Import

- Kiểm tra path aliases trong tsconfig.app.json
- Restart development server

### Lỗi Charts

- Kiểm tra @ant-design/plots đã cài đặt
- Kiểm tra data format từ API
