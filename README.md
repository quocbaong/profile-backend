# Portfolio Backend API (Golang)

Production-grade RESTful API backend for Nguyễn Quốc Bảo's personal portfolio & administration dashboard, built with **Go**, **Gin Web Framework**, and **GORM**.

---

## 🏗 Architecture & Project Structure

Follows the **Standard Go Project Layout**:

```
profile-backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entrypoint
├── internal/                    # Private application code
│   ├── config/                  # Environment configuration
│   ├── database/                # GORM DB connection & Auto-migration
│   ├── models/                  # Database entity models
│   ├── repository/              # Data access layer (CRUD queries)
│   ├── service/                 # Business logic & authentication
│   ├── handler/                 # HTTP controller handlers (Gin)
│   ├── middleware/              # Auth (JWT) & CORS middlewares
│   └── router/                  # API routes registration
├── pkg/                         # Reusable public packages
│   ├── jwt/                     # JWT token generator & validator
│   └── response/                # Unified JSON response envelope
├── .env.example                 # Environment variables template
├── Dockerfile                   # Multi-stage lightweight Docker build (~15MB)
├── go.mod                       # Go modules definition
└── README.md                    # Project documentation
```

---

## ⚡ Tech Stack

- **Language:** Go 1.23+
- **Framework:** [Gin](https://github.com/gin-gonic/gin)
- **ORM:** [GORM](https://gorm.io/)
- **Databases Supported:**
  - **SQLite** (Pure-Go via `github.com/glebarez/sqlite` - Zero configuration for local dev)
  - **PostgreSQL** (Production-ready via `gorm.io/driver/postgres`)
- **Authentication:** JWT (JSON Web Tokens via `golang-jwt/jwt/v5`) & `bcrypt` password hashing
- **Deployment:** Docker / Render / Railway / Fly.io

---

## 🚀 Getting Started

### 1. Prerequisites
- Go 1.23 or newer installed (`go version`)

### 2. Run Locally
```bash
# Clone the repository
git clone https://github.com/quocbaong/profile-backend.git
cd profile-backend

# Download dependencies
go mod tidy

# Run server (default uses local SQLite: portfolio.db)
go run ./cmd/api
```
The server will start at `http://localhost:8080`.

On first run, GORM automatically runs table migrations and seeds the default admin user:
- **Username:** `admin`
- **Password:** `Admin@123456`

---

## 📡 API Endpoints

### Public Endpoints (Client / Visitor)
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/health` | Health check endpoint |
| `GET` | `/api/projects` | Get all active projects for portfolio |
| `POST` | `/api/contact` | Submit a new contact message from website form |
| `POST` | `/api/auth/login` | Admin login to receive JWT Bearer Token |

### Protected Endpoints (Admin Dashboard - Requires `Authorization: Bearer <token>`)
| Method | Endpoint | Description |
| :--- | :--- | :--- |
| `GET` | `/api/admin/me` | Get current logged-in admin profile |
| `POST` | `/api/admin/change-password` | Change admin account password |
| `GET` | `/api/admin/projects` | List all projects (including inactive) |
| `GET` | `/api/admin/projects/:id` | Get single project detail |
| `POST` | `/api/admin/projects` | Create a new project |
| `PUT` | `/api/admin/projects/:id` | Update project by ID |
| `DELETE` | `/api/admin/projects/:id` | Delete project by ID |
| `GET` | `/api/admin/messages` | List all contact messages from visitors |
| `GET` | `/api/admin/messages/unread-count` | Get count of unread messages (for badge) |
| `PATCH` | `/api/admin/messages/:id/read` | Mark message as read |
| `DELETE` | `/api/admin/messages/:id` | Delete contact message |

---

## 🐳 Docker Deployment

```bash
# Build Docker image
docker build -t quocbao-portfolio-api .

# Run container
docker run -p 8080:8080 quocbao-portfolio-api
```
The resulting Docker image is only **~15MB**.
