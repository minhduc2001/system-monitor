# System Monitoring Dashboard

## Tổng quan

React dashboard để hiển thị system monitoring và logs từ Go Runner API. Dashboard được xây dựng với React, TypeScript, Ant Design và React Query.

## Tính năng chính

### 📊 **System Overview**

- Real-time system information
- CPU, Memory, Disk, Network usage
- System status indicators
- Process count và uptime
- Load average monitoring

### 🚨 **System Alerts**

- Active alerts với filtering
- Alert levels (info, warning, error, critical)
- Real-time alert updates
- Alert resolution actions

### 📈 **System Metrics**

- Historical metrics charts
- Resource usage trends
- Configurable time ranges
- Export functionality
- Average statistics

### ⚡ **Top Processes**

- Process resource usage
- CPU và Memory sorting
- Process status indicators
- Command line display
- User information

## Cấu trúc Project

```
src/
├── components/system/
│   ├── SystemOverview.tsx      # System overview component
│   ├── SystemAlerts.tsx        # Alerts management
│   ├── SystemMetrics.tsx       # Metrics charts
│   └── TopProcesses.tsx        # Process monitoring
├── pages/system/
│   ├── SystemOverview.tsx      # Overview page
│   ├── SystemAlerts.tsx        # Alerts page
│   ├── SystemMetrics.tsx       # Metrics page
│   └── SystemProcesses.tsx     # Processes page
├── api/
│   └── system.ts               # System API client
├── hooks/queries/
│   └── use-system.query.ts     # React Query hooks
├── types/
│   └── system.ts               # TypeScript types
└── utils/
    └── system.ts               # Utility functions
```

## Cài đặt Dependencies

```bash
npm install @ant-design/plots dayjs
```

## Cấu hình Environment

Tạo file `.env`:

```env
VITE_API_URL=http://localhost:8080
```

## Sử dụng

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

```
http://localhost:5173
```

## Components

### SystemOverview

Hiển thị tổng quan hệ thống với:

- System status và uptime
- Resource usage (CPU, Memory, Disk, Network)
- System information
- Status indicators

### SystemAlerts

Quản lý alerts với:

- Filtering theo type, level, status
- Real-time updates
- Alert actions
- Pagination

### SystemMetrics

Hiển thị metrics với:

- Line charts cho resource usage
- Time range selection
- Average statistics
- Export functionality

### TopProcesses

Monitor processes với:

- CPU/Memory sorting
- Process details
- Status indicators
- Command line display

## API Integration

### React Query Hooks

- `useSystemInfo()` - System information
- `useSystemStatus()` - System status
- `useSystemDashboard()` - Dashboard data
- `useSystemMetrics()` - Historical metrics
- `useSystemAlerts()` - System alerts
- `useSystemConfig()` - Configuration

### Auto-refresh

- System info: 30 seconds
- System status: 10 seconds
- Dashboard: 15 seconds
- Metrics: 60 seconds
- Alerts: 30 seconds

## Styling

### Ant Design Components

- Cards cho layout
- Tables cho data display
- Progress bars cho usage
- Tags cho status
- Charts cho metrics

### Color Coding

- 🟢 Healthy: Green (#52c41a)
- 🟡 Warning: Orange (#faad14)
- 🔴 Critical: Red (#ff4d4f)

## Responsive Design

- Mobile-first approach
- Responsive grid system
- Collapsible sidebar
- Mobile-optimized tables

## Performance

### Optimization

- React Query caching
- Component memoization
- Lazy loading
- Pagination
- Debounced search

### Data Management

- Automatic refetching
- Stale-while-revalidate
- Background updates
- Error handling

## Development

### Scripts

```bash
npm run dev          # Development server
npm run build        # Production build
npm run preview      # Preview build
npm run lint         # Lint code
```

### Code Structure

- TypeScript cho type safety
- ESLint cho code quality
- Prettier cho formatting
- Path aliases (@/)

## Deployment

### Build

```bash
npm run build
```

### Environment Variables

```env
VITE_API_URL=https://api.example.com
```

### Nginx Configuration

```nginx
location / {
  try_files $uri $uri/ /index.html;
}
```

## Troubleshooting

### Common Issues

1. **API Connection Failed**

   - Kiểm tra VITE_API_URL
   - Kiểm tra CORS settings
   - Kiểm tra Go Runner API

2. **Charts Not Loading**

   - Kiểm tra @ant-design/plots
   - Kiểm tra data format
   - Kiểm tra console errors

3. **Real-time Updates Not Working**
   - Kiểm tra React Query config
   - Kiểm tra refetch intervals
   - Kiểm tra network connection

### Debug Mode

```bash
# Enable debug logging
VITE_DEBUG=true npm run dev
```

## Features Roadmap

### Phase 1 ✅

- [x] System overview
- [x] Alerts management
- [x] Metrics charts
- [x] Process monitoring

### Phase 2 🔄

- [ ] Real-time WebSocket updates
- [ ] Alert notifications
- [ ] Custom dashboards
- [ ] Export functionality

### Phase 3 📋

- [ ] System configuration
- [ ] User management
- [ ] Role-based access
- [ ] Mobile app

## Contributing

1. Fork repository
2. Create feature branch
3. Make changes
4. Add tests
5. Submit pull request

## License

MIT License - see LICENSE file for details
