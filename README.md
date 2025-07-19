# TaskFlow API 🚀

A production-ready task management REST API built with Go, Fiber, PostgreSQL, and Docker. TaskFlow provides comprehensive project and task management capabilities with JWT authentication, CRUD operations, and proper database relationships.

## 🏗️ Architecture

- **Backend Framework**: Go with Fiber v2
- **Database**: PostgreSQL with GORM ORM
- **Authentication**: JWT tokens
- **Migrations**: Goose migration tool
- **Containerization**: Docker & Docker Compose
- **Build Tool**: Makefile

## 📁 Project Structure

```
taskflow-api/
├── cmd/main.go                 # Application entry point
├── internal/
│   ├── config/config.go        # Configuration management
│   ├── models/                 # Data models and DTOs
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── task.go
│   │   └── common.go
│   ├── handlers/               # HTTP request handlers
│   │   ├── user_handler.go
│   │   ├── project_handler.go
│   │   └── task_handler.go
│   ├── routes/routes.go        # Route definitions
│   └── middleware/auth.go      # JWT authentication
├── migrations/                 # Database migrations
├── docker-compose.yml          # Docker services
├── Dockerfile                  # API container
├── Dockerfile.migrations       # Migration container
├── Makefile                    # Development commands
├── .env.example               # Environment template
└── README.md                  # This file
```

## 🗄️ Database Schema

### Users Table
- `id` (UUID, primary key)
- `email` (unique, not null)
- `password_hash` (bcrypt)
- `first_name`, `last_name`
- `avatar_url` (optional)
- `is_active` (boolean)
- `created_at`, `updated_at`

### Projects Table
- `id` (UUID, primary key)
- `name` (not null)
- `description` (text)
- `color` (hex color code)
- `owner_id` (foreign key to users)
- `status` (enum: active, archived, completed)
- `created_at`, `updated_at`

### Tasks Table
- `id` (UUID, primary key)
- `title` (not null)
- `description` (text)
- `project_id` (foreign key to projects)
- `assignee_id` (foreign key to users, nullable)
- `status` (enum: todo, in_progress, done, cancelled)
- `priority` (enum: low, medium, high, urgent)
- `due_date` (timestamp, nullable)
- `completed_at` (timestamp, nullable)
- `created_at`, `updated_at`

## 🚀 Quick Start

### Prerequisites
- Go 1.21+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### 1. Clone and Setup
```bash
git clone <repository-url>
cd taskflow-api
make setup  # or cp .env.example .env && go mod tidy
```

### 2. Configure Environment
Edit `.env` file with your settings:
```bash
# Server
PORT=8080
ENV=development

# Database
DB_HOST=localhost
DB_PORT=5433
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=taskflow
DB_SSL_MODE=disable

# JWT
JWT_SECRET=your_super_secret_jwt_key
JWT_EXPIRY=24h
```

### 3. Start with Docker (Recommended)
```bash
make start  # or docker-compose up -d
```

### 4. Start for Development
```bash
make dev    # Starts DB and runs API locally
```

### 5. Health Check
```bash
curl http://localhost:8080/health
```

## 📡 API Endpoints

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user

### Users (Protected)
- `GET /api/v1/users` - List users (paginated)
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user (soft delete)

### Projects (Protected)
- `POST /api/v1/projects` - Create project
- `GET /api/v1/projects` - List user's projects
- `GET /api/v1/projects/:id` - Get project with tasks
- `PUT /api/v1/projects/:id` - Update project
- `DELETE /api/v1/projects/:id` - Delete project

### Tasks (Protected)
- `POST /api/v1/projects/:project_id/tasks` - Create task
- `GET /api/v1/projects/:project_id/tasks` - List project tasks
- `GET /api/v1/tasks/:id` - Get task details
- `PUT /api/v1/tasks/:id` - Update task
- `DELETE /api/v1/tasks/:id` - Delete task
- `PATCH /api/v1/tasks/:id/status` - Update task status

### Health Check
- `GET /health` - API health status

## 🔐 Authentication

The API uses JWT tokens for authentication. Include the token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

### Example Registration
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123",
    "first_name": "John",
    "last_name": "Doe"
  }'
```

### Example Login
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

## 🛠️ Development Commands

### Application
```bash
make run           # Run the application
make build         # Build the application
make test          # Run tests
make clean         # Clean build artifacts
make format        # Format code
```

### Docker
```bash
make docker-up     # Start all containers
make docker-down   # Stop all containers
make docker-build  # Build Docker image
make logs          # Show container logs
```

### Database
```bash
make db-up         # Start only database
make db-shell      # Connect to database
make db-backup     # Backup database
make db-reset      # Reset database (WARNING: deletes data)
```

### Migrations
```bash
make migrate       # Run all migrations
make migrate-down  # Rollback last migration
make migrate-status # Show migration status
make migrate-create NAME=migration_name # Create new migration
```

### Quick Commands
```bash
make start         # Quick start (docker-up)
make stop          # Quick stop (docker-down)
make dev           # Development mode (db + local API)
make health        # Check API health
make info          # Show project information
make help          # Show all available commands
```

## 📊 API Response Format

### Success Response
```json
{
  "message": "Success message",
  "data": { ... }
}
```

### Error Response
```json
{
  "error": "Error Type",
  "message": "Detailed error message",
  "code": 400
}
```

### Paginated Response
```json
{
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100,
    "total_pages": 10
  }
}
```

## 🔧 Configuration

### Environment Variables
| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `ENV` | Environment (development/production) | development |
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5433 |
| `DB_USER` | Database user | postgres |
| `DB_PASSWORD` | Database password | password |
| `DB_NAME` | Database name | taskflow |
| `JWT_SECRET` | JWT signing secret | (required) |
| `JWT_EXPIRY` | Token expiry duration | 24h |

### Database Connection Pool
- Max Idle Connections: 10
- Max Open Connections: 100
- Connection Max Lifetime: 1 hour

## 🧪 Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# View coverage report
open coverage.html
```

## 📦 Production Deployment

### Docker Production Build
```bash
# Build production image
docker build -t taskflow-api:latest .

# Run with production environment
docker run -d \
  --name taskflow-api \
  -p 8080:8080 \
  -e ENV=production \
  -e DB_HOST=your-db-host \
  -e JWT_SECRET=your-production-secret \
  taskflow-api:latest
```

### Database Migrations in Production
```bash
# Run migrations
docker run --rm \
  -e DB_HOST=your-db-host \
  -e DB_USER=your-db-user \
  -e DB_PASSWORD=your-db-password \
  -e DB_NAME=your-db-name \
  taskflow-migrations:latest \
  -dir /migrations postgres "postgres://user:pass@host:port/dbname?sslmode=require" up
```

## 🛡️ Security Features

- **Password Hashing**: bcrypt with default cost
- **JWT Authentication**: Secure token-based auth
- **Input Validation**: Request validation with struct tags
- **SQL Injection Protection**: GORM ORM with parameterized queries
- **CORS Support**: Configurable cross-origin requests
- **Rate Limiting**: Basic rate limiting middleware
- **Graceful Shutdown**: Proper connection cleanup

## 🚦 Health Monitoring

The API includes a health check endpoint that monitors:
- Application status
- Database connectivity
- Service dependencies

Access at: `GET /health`

## 🔍 Troubleshooting

### Common Issues

1. **Database Connection Failed**
   ```bash
   # Check if database is running
   make db-up
   
   # Verify connection settings in .env
   make info
   ```

2. **Migration Errors**
   ```bash
   # Check migration status
   make migrate-status
   
   # Reset and re-run migrations
   make migrate-reset
   make migrate
   ```

3. **JWT Token Issues**
   ```bash
   # Ensure JWT_SECRET is set in .env
   # Check token expiry (default 24h)
   ```

### Logs
```bash
# View application logs
make logs

# View database logs
docker logs taskflow-db
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 👥 Authors

- **Your Name** - *Initial work* - [YourGitHub](https://github.com/yourusername)

## 🙏 Acknowledgments

- [Fiber](https://gofiber.io/) - Express-inspired web framework
- [GORM](https://gorm.io/) - Fantastic ORM library for Go
- [Goose](https://github.com/pressly/goose) - Database migration tool
- [PostgreSQL](https://www.postgresql.org/) - World's most advanced open source database

---

**TaskFlow API** - Built with ❤️ using Go and modern technologies