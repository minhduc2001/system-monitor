# Hướng dẫn cấu hình Projects

## Tổng quan

Bạn có thể import nhiều projects cùng lúc bằng cách sử dụng file cấu hình YAML hoặc JSON. Điều này giúp bạn quản lý toàn bộ microservices một cách dễ dàng.

## Cấu trúc file cấu hình

### Format YAML

File cấu hình YAML có cấu trúc như sau:

```yaml
projects:
  - name: "Project Name"
    description: "Project description"
    type: "backend"  # backend, frontend, worker, database, queue, other
    path: "/path/to/project"
    command: "go run main.go"
    port: 3001
    environment: "development"
    # ... các trường khác

groups:
  - name: "Group Name"
    description: "Group description"
    color: "#3B82F6"
```

### Format JSON

File cấu hình JSON có cấu trúc tương tự:

```json
{
  "projects": [
    {
      "name": "Project Name",
      "description": "Project description",
      "type": "backend",
      "path": "/path/to/project",
      "command": "go run main.go",
      "port": 3001,
      "environment": "development"
    }
  ],
  "groups": [
    {
      "name": "Group Name",
      "description": "Group description",
      "color": "#3B82F6"
    }
  ]
}
```

## Các trường bắt buộc

### Project

- **name** (string, required): Tên của project
- **path** (string, required): Đường dẫn tuyệt đối đến thư mục project
- **type** (string, required): Loại service, một trong các giá trị:
  - `backend`: Backend service (Go, Node.js, Python, etc.)
  - `frontend`: Frontend application (React, Vue, Angular, etc.)
  - `worker`: Background worker (Celery, Bull, etc.)
  - `database`: Database service (PostgreSQL, MySQL, etc.)
  - `queue`: Message queue (RabbitMQ, Redis, etc.)
  - `other`: Loại khác

## Các trường tùy chọn

### Project

- **description** (string): Mô tả về project
- **group_id** (number): ID của project group (sẽ được tạo nếu chưa tồn tại)
- **command** (string): Lệnh để chạy service. Nếu không có, hệ thống sẽ tự động chọn dựa trên type:
  - `backend`: `go run main.go` hoặc `npm start`
  - `frontend`: `npm start`
  - `worker`: `node worker.js`
- **args** (string): Các tham số bổ sung cho command
- **working_dir** (string): Thư mục làm việc (mặc định là path)
- **port** (number): Port mà service chạy trên
- **ports** (string): JSON array string cho nhiều ports, ví dụ: `["3000", "3001"]`
- **environment** (string): Môi trường (`development`, `staging`, `production`)
- **env_file** (string): Đường dẫn đến file .env
- **env_vars** (string): JSON string chứa environment variables, ví dụ: `{"KEY": "value"}`
- **editor** (string): Editor để mở project (vscode, intellij, etc.)
- **editor_args** (string): Tham số bổ sung cho editor
- **health_check_url** (string): URL để kiểm tra health của service
- **auto_restart** (boolean): Tự động restart khi service crash (mặc định: false)
- **max_restarts** (number): Số lần restart tối đa (mặc định: 3)
- **cpu_limit** (string): Giới hạn CPU, ví dụ: `500m`, `1`
- **memory_limit** (string): Giới hạn memory, ví dụ: `512Mi`, `1Gi`

### Project Group

- **name** (string, required): Tên của group
- **description** (string): Mô tả về group
- **color** (string): Màu hex cho UI, ví dụ: `#3B82F6`

## Ví dụ đầy đủ

### YAML Example

Xem file `examples/project-config.example.yaml` để xem ví dụ đầy đủ.

### JSON Example

Xem file `examples/project-config.example.json` để xem ví dụ đầy đủ.

## Cách sử dụng

1. **Tạo file cấu hình**: Tạo file YAML hoặc JSON theo cấu trúc trên
2. **Import trong UI**: 
   - Vào trang Projects
   - Click nút "Import Config"
   - Chọn file cấu hình
   - Xác nhận import
3. **Import qua API**:
   ```bash
   curl -X POST http://localhost:8080/api/v1/projects/import \
     -H "Content-Type: application/json" \
     -d @project-config.json
   ```

## Lưu ý

- Đường dẫn `path` phải là đường dẫn tuyệt đối
- Nếu `group_id` không tồn tại, hệ thống sẽ bỏ qua và không gán group
- Nếu project đã tồn tại (theo name), hệ thống sẽ bỏ qua hoặc cập nhật (tùy cấu hình)
- File cấu hình có thể chứa cả `projects` và `groups`, hoặc chỉ một trong hai

## Best Practices

1. **Tổ chức theo nhóm**: Sử dụng project groups để tổ chức các services liên quan
2. **Đặt tên rõ ràng**: Sử dụng tên mô tả rõ ràng cho projects
3. **Cấu hình environment**: Sử dụng các file .env riêng cho từng môi trường
4. **Health checks**: Luôn cấu hình health_check_url để monitoring
5. **Resource limits**: Đặt giới hạn CPU và memory để tránh resource exhaustion
6. **Auto restart**: Bật auto_restart cho các services quan trọng

## Troubleshooting

### Lỗi "Path not found"
- Đảm bảo đường dẫn `path` là đường dẫn tuyệt đối và tồn tại
- Kiểm tra quyền truy cập vào thư mục

### Lỗi "Invalid type"
- Đảm bảo `type` là một trong các giá trị hợp lệ: `backend`, `frontend`, `worker`, `database`, `queue`, `other`

### Lỗi "Port already in use"
- Kiểm tra xem port đã được sử dụng bởi service khác chưa
- Thay đổi port trong cấu hình

### Lỗi "Group not found"
- Nếu `group_id` không tồn tại, hệ thống sẽ bỏ qua và không gán group
- Tạo group trước khi import projects, hoặc import groups cùng với projects

