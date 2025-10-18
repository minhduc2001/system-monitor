# Go Runner - Microservice Management Platform

A powerful local microservice management platform similar to a "mini Kubernetes" for development environments. Built with Go, Gin, GORM, and WebSocket support.

## Features

- 🚀 **Microservice Management** - Start, stop, restart services with real-time monitoring
- 📊 **Project Groups** - Organize services by teams (backend, frontend, etc.)
- 🔄 **Real-time Logs** - WebSocket-powered live log streaming
- 🏥 **Health Monitoring** - Built-in health checks and status monitoring
- ⚙️ **Flexible Configuration** - YAML config files with environment variable override
- 🗄️ **Multi-database support** - SQLite, PostgreSQL, MySQL
- 🐳 **Docker support** - Ready-to-use Docker and Docker Compose setup
- 🔥 **Hot Reload** - Multiple hot reload options (Air, custom watcher)
- 🛡️ **Advanced Middleware** - CORS, logging, recovery, error handling, validation, rate limiting
- 📱 **Web Interface Ready** - API designed for web/desktop applications

## Quick Start

### Prerequisites

- Go 1.21 or higher
- CGO enabled (required for SQLite support)
- Docker (optional)

**Note:** This project requires CGO to be enabled for SQLite database support. Make sure your Go installation supports CGO.

### Installation

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd go-runner
   ```

2. **Install dependencies**

   ```bash
   make deps
   # or
   go mod download
   ```

3. **Run the application**

   ```bash
   # Normal run
   # Windows (Command Prompt)
   run.bat

   # Windows (PowerShell)
   .\run.ps1

   # Linux/Mac
   ./run.sh

   # Hot reload development
   # Windows
   hot-reload.bat

   # Linux/Mac
   ./hot-reload.sh

   # PowerShell
   .\hot-reload.ps1

   # Or use make
   make run          # Normal run
   make hot-reload   # Hot reload
   make dev-air      # Air hot reload
   ```

4. **Access the API**
   - API: http://localhost:8080/api/v1
   - Health check: http://localhost:8080/health

### Using Docker

1. **Build and run with Docker Compose**

   ```bash
   make up
   # or
   docker-compose up -d
   ```

2. **View logs**

   ```bash
   make logs
   # or
   docker-compose logs -f
   ```

3. **Stop services**
   ```bash
   make down
   # or
   docker-compose down
   ```

## Configuration

The application can be configured using:

1. **YAML config file** (`config.yaml`)
2. **Environment variables**
3. **Command line flags**

### Default Configuration

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "debug"
  read_timeout: 30
  write_timeout: 30

database:
  driver: "sqlite"
  path: "./data/project.db"

logging:
  level: "info"
  format: "json"
  output: "stdout"
```

### Environment Variables

You can override any configuration using environment variables:

```bash
export SERVER_PORT=9000
export DATABASE_DRIVER=postgres
export DATABASE_HOST=localhost
export DATABASE_PORT=5432
export DATABASE_USERNAME=postgres
export DATABASE_PASSWORD=password
export DATABASE_DBNAME=go_runner
```

## API Endpoints

### Health Check

- `GET /health` - Service health status

### Project Groups

- `GET /api/v1/groups` - List all project groups
- `POST /api/v1/groups` - Create a new project group
- `GET /api/v1/groups/:id` - Get project group by ID
- `PUT /api/v1/groups/:id` - Update project group
- `DELETE /api/v1/groups/:id` - Delete project group
- `GET /api/v1/groups/:id/projects` - Get projects in a group

### Microservices (Projects)

- `GET /api/v1/projects` - List all microservices
- `POST /api/v1/projects` - Create a new microservice
- `GET /api/v1/projects/:id` - Get microservice by ID
- `PUT /api/v1/projects/:id` - Update microservice
- `DELETE /api/v1/projects/:id` - Delete microservice
- `GET /api/v1/projects/:id/status` - Get microservice status
- `GET /api/v1/projects/:id/logs` - Get microservice logs (static)
- `GET /api/v1/projects/:id/logs/ws` - WebSocket for real-time logs

### Service Management

- `POST /api/v1/projects/:id/start` - Start microservice
- `POST /api/v1/projects/:id/stop` - Stop microservice
- `POST /api/v1/projects/:id/restart` - Restart microservice
- `GET /api/v1/services/running` - Get all running services

### Example API Usage

**Create a project group:**

```bash
curl -X POST http://localhost:8080/api/v1/groups \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Backend Team",
    "description": "Backend microservices team",
    "color": "#3B82F6"
  }'
```

**Create a microservice:**

```bash
curl -X POST http://localhost:8080/api/v1/projects \
  -H "Content-Type: application/json" \
  -d '{
    "name": "User Service",
    "description": "User management microservice",
    "type": "backend",
    "group_id": 1,
    "path": "/path/to/user-service",
    "command": "go run main.go",
    "port": 3001,
    "environment": "development",
    "editor": "vscode",
    "health_check_url": "http://localhost:3001/health",
    "auto_restart": true
  }'
```

**Start a service:**

```bash
curl -X POST http://localhost:8080/api/v1/projects/1/start
```

**Get running services:**

```bash
curl http://localhost:8080/api/v1/services/running
```

## Error Handling & Validation

The API includes comprehensive error handling and validation:

### Error Response Format

```json
{
  "error": "Bad Request",
  "message": "Validation failed",
  "code": 400,
  "details": [
    {
      "field": "name",
      "tag": "required",
      "value": "",
      "message": "name is required"
    }
  ]
}
```

### Validation Rules

- **Project Name**: Required, 1-100 characters
- **Description**: Max 500 characters
- **Port**: Valid port number (1-65535)
- **Environment**: One of: development, staging, production
- **Health Check URL**: Valid URL format
- **Color**: Valid hex color format (#RRGGBB)

### Middleware Stack

1. **Error Handler** - Catches panics and converts to JSON responses
2. **Request Logger** - Logs all requests with timing and details
3. **Error Logger** - Logs all errors with context
4. **Security Headers** - Adds security headers
5. **Rate Limiter** - Prevents abuse (100 requests/minute)
6. **CORS** - Handles cross-origin requests

### Testing Error Handling

Use the `test_error_handling.http` file to test various error scenarios.

## Hot Reload

The project supports multiple hot reload options for development:

### Option 1: Custom Hot Reload Watcher (Recommended)

```bash
# Windows
hot-reload.bat

# Linux/Mac
./hot-reload.sh

# PowerShell
.\hot-reload.ps1

# Or using Make
make hot-reload
```

### Option 2: Air (Third-party)

```bash
# Install Air first
go install github.com/cosmtrek/air@latest

# Run with Air
make dev-air
# or
CGO_ENABLED=1 air
```

### Hot Reload Configuration

Configure hot reload in `config.yaml`:

```yaml
hot_reload:
  enabled: true
  watch_dirs:
    - "."
    - "cmd"
    - "internal"
  exclude_dirs:
    - "tmp"
    - "vendor"
    - "node_modules"
    - ".git"
  include_exts:
    - ".go"
    - ".yaml"
    - ".yml"
    - ".json"
  exclude_exts:
    - ".log"
    - ".tmp"
  delay: 1000 # milliseconds
  build_cmd: "CGO_ENABLED=1 go build -o ./tmp/main cmd/server/main.go"
  run_cmd: "./tmp/main"
  log_level: "info"
```

### Features

- ✅ **File Watching** - Monitors `.go`, `.yaml`, `.json` files
- ✅ **Smart Exclusions** - Ignores `tmp`, `vendor`, `node_modules`
- ✅ **Configurable** - Customize watch directories and file extensions
- ✅ **Build Integration** - Automatic build and restart on changes
- ✅ **Logging** - Detailed logs for debugging

## Database Support

### SQLite (Default)

No additional setup required. Database file will be created automatically.

### PostgreSQL

1. Start PostgreSQL service
2. Update configuration:
   ```yaml
   database:
     driver: "postgres"
     host: "localhost"
     port: 5432
     username: "postgres"
     password: "password"
     dbname: "go_runner"
     sslmode: "disable"
   ```

### MySQL

1. Start MySQL service
2. Update configuration:
   ```yaml
   database:
     driver: "mysql"
     host: "localhost"
     port: 3306
     username: "go_runner"
     password: "password"
     dbname: "go_runner"
   ```

## Development

### Project Structure

```
go-runner/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/
│   ├── app/
│   │   ├── router.go        # Route definitions
│   │   └── server.go        # Server setup
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── db/
│   │   └── db.go           # Database connection
│   ├── middleware/
│   │   ├── cors.go         # CORS middleware
│   │   ├── logger.go       # Logging middleware
│   │   └── recovery.go     # Recovery middleware
│   └── project/
│       ├── handler.go      # Project handlers
│       └── model.go        # Project models
├── config.yaml             # Configuration file
├── docker-compose.yml      # Docker Compose setup
├── Dockerfile             # Docker image
├── Makefile              # Build commands
└── README.md             # This file
```

### Available Commands

```bash
make help                 # Show all available commands
make build               # Build the application
make run                 # Run the application
make dev                 # Run with hot reload
make test                # Run tests
make test-coverage       # Run tests with coverage
make clean               # Clean build artifacts
make fmt                 # Format code
make lint                # Lint code
make docker-build        # Build Docker image
make docker-run          # Run Docker container
make up                  # Start with Docker Compose
make down                # Stop Docker Compose
make logs                # Show logs
```

### Hot Reload Development

For development with automatic restart, install `air`:

```bash
go install github.com/cosmtrek/air@latest
```

Then run:

```bash
make dev
```

## Testing

Run tests:

```bash
make test
```

Run tests with coverage:

```bash
make test-coverage
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run tests and linting
6. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For support and questions, please open an issue in the repository.
