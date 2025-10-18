# Swagger API Documentation

## Tổng quan

Go Runner API đã được tích hợp Swagger để tạo documentation tự động và interactive API testing.

## Truy cập Swagger UI

Sau khi khởi động server, bạn có thể truy cập Swagger UI tại:

```
http://localhost:8080/swagger/index.html
```

## Các API Endpoints có sẵn

### Thông tin cơ bản

- **GET /**: Thông tin về API
- **GET /health**: Health check endpoint

### Projects API

- **GET /api/v1/projects**: Lấy danh sách tất cả projects
- **POST /api/v1/projects**: Tạo project mới
- **GET /api/v1/projects/:id**: Lấy thông tin project theo ID
- **PUT /api/v1/projects/:id**: Cập nhật project
- **DELETE /api/v1/projects/:id**: Xóa project

### Project Operations

- **POST /api/v1/projects/:id/start**: Khởi động project
- **POST /api/v1/projects/:id/stop**: Dừng project
- **POST /api/v1/projects/:id/restart**: Khởi động lại project
- **GET /api/v1/projects/:id/status**: Lấy trạng thái project
- **GET /api/v1/projects/:id/logs**: Lấy logs của project

### Project Groups

- **GET /api/v1/groups**: Lấy danh sách groups
- **POST /api/v1/groups**: Tạo group mới
- **GET /api/v1/groups/:id**: Lấy thông tin group
- **PUT /api/v1/groups/:id**: Cập nhật group
- **DELETE /api/v1/groups/:id**: Xóa group

## Tạo và cập nhật Swagger Documentation

### Cài đặt Swagger CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

### Generate Swagger docs

```bash
# Generate với parsing dependencies và internal packages
go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs --parseDependency --parseInternal

# Hoặc nếu đã cài swag
swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal
```

### Thêm Swagger Annotations vào API

Swagger sử dụng comments đặc biệt để generate documentation. Ví dụ:

```go
// GetProjects godoc
// @Summary      Get all projects
// @Description  Get a list of all projects
// @Tags         projects
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "List of projects"
// @Failure      500  {object}  map[string]interface{}  "Internal server error"
// @Router       /projects [get]
func (h *Handler) GetProjects(c *gin.Context) {
    // Handler implementation
}
```

### Cấu hình chính trong main.go

```go
// @title           Go Runner API
// @version         1.0
// @description     A microservices management API built with Go and Gin
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  MIT
// @license.url   https://opensource.org/licenses/MIT

// @host      localhost:8080
// @BasePath  /api/v1
```

## Build và Run

### Build với CGO disabled (khuyến nghị cho Windows)

```bash
# PowerShell
$env:CGO_ENABLED='1'; go build -o tmp/main.exe cmd/server/main.go

# Bash
CGO_ENABLED=1 go build -o tmp/main cmd/server/main.go
```

### Run server

```bash
# Windows
tmp\main.exe

# Linux/Mac
./tmp/main
```

## Sử dụng Swagger UI

1. Mở trình duyệt và truy cập `http://localhost:8080/swagger/index.html`
2. Xem danh sách các API endpoints
3. Click vào bất kỳ endpoint nào để xem chi tiết
4. Click "Try it out" để test API trực tiếp
5. Nhập parameters (nếu có) và click "Execute"
6. Xem response trả về từ server

## Swagger Annotations Reference

### Common Annotations

- `@Summary`: Mô tả ngắn gọn về endpoint
- `@Description`: Mô tả chi tiết
- `@Tags`: Nhóm các endpoint liên quan
- `@Accept`: Format dữ liệu nhận vào (json, xml, etc.)
- `@Produce`: Format dữ liệu trả về
- `@Param`: Định nghĩa parameters
- `@Success`: Response thành công
- `@Failure`: Response lỗi
- `@Router`: Đường dẫn và HTTP method

### Parameter Types

- `path`: URL path parameter (e.g., /users/:id)
- `query`: Query string parameter (e.g., ?name=value)
- `header`: HTTP header
- `body`: Request body
- `formData`: Form data

## Troubleshooting

### Swagger UI không hiển thị

1. Kiểm tra xem docs package đã được import:

   ```go
   import _ "go-runner/docs"
   ```

2. Kiểm tra router có đăng ký endpoint:

   ```go
   r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
   ```

3. Regenerate docs:
   ```bash
   go run github.com/swaggo/swag/cmd/swag@latest init -g cmd/server/main.go -o docs --parseDependency --parseInternal
   ```

### CGO Error khi build

Sử dụng CGO_ENABLED=0 để build mà không cần C compiler:

```bash
# PowerShell
$env:CGO_ENABLED='0'; go build -o tmp/main.exe cmd/server/main.go

# Bash
CGO_ENABLED=0 go build -o tmp/main cmd/server/main.go
```

## Resources

- [Swaggo Documentation](https://github.com/swaggo/swag)
- [Swagger Specification](https://swagger.io/specification/)
- [Gin Swagger](https://github.com/swaggo/gin-swagger)
