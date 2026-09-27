# Go Starter App

A production-ready Go starter application with **Email & Password Authentication**, **JWT Access and Refresh Tokens**, **PostgreSQL (GORM)**, and a clean **layered architecture** structured after [SahilKarwasra/Trello-App](https://github.com/SahilKarwasra/Trello-App).

All API endpoints strictly follow the unified response wrapper schema:
```json
{
  "statusCode": 200,
  "data": {},
  "isSuccess": true,
  "message": "Operation successful"
}
```

---

## 📁 Project Structure

```
Go-Starter-App/
├── apps/
│   └── api/
│       ├── cmd/
│       │   ├── migrate/
│       │   │   └── main.go                 # Database migration entrypoint
│       │   └── server/
│       │       ├── handler/
│       │       │   ├── auth_handler.go     # HTTP request handlers using response wrapper
│       │       │   └── auth_handler_test.go
│       │       ├── middleware/
│       │       │   └── auth_middleware.go  # JWT Bearer token validation middleware
│       │       ├── repository/
│       │       │   └── user_repository.go  # Database query layer (GORM)
│       │       ├── routes/
│       │       │   └── routes.go           # Gin router setup and endpoint registration
│       │       ├── services/
│       │       │   ├── auth_dto.go         # Request / Response DTO structs
│       │       │   ├── auth_service.go     # Business logic, token generation & rotation
│       │       │   └── auth_service_test.go
│       │       ├── utils/
│       │       │   ├── errors.go           # Common sentinel errors
│       │       │   ├── jwt.go              # JWT access & refresh token helper
│       │       │   ├── password.go         # bcrypt hashing and checking
│       │       │   ├── response.go         # APIResponse wrapper (Success, Error, BadRequest, etc.)
│       │       │   ├── jwt_test.go
│       │       │   └── password_test.go
│       │       └── main.go                 # API server entrypoint
│       └── go.mod
├── packages/
│   ├── config/
│   │   ├── config.go                       # Environment variable loader (godotenv)
│   │   └── go.mod
│   └── database/
│       ├── migrate/
│       │   └── migrate.go                  # Auto-migration schema definitions
│       ├── models/
│       │   └── users.go                    # GORM User model
│       ├── postgres.go                     # Database connection pool setup
│       └── go.mod
├── .env.example                            # Sample environment variables
├── .env                                    # Local development environment
├── docker-compose.yml                      # PostgreSQL service
├── go.work                                 # Go workspace configuration
├── Makefile                                # Handy commands
└── README.md
```

---

## 🚀 Getting Started

### 1. Prerequisites
- **Go 1.25+** (tested with Go 1.27)
- **Docker** & **Docker Compose** (for PostgreSQL)

### 2. Start PostgreSQL
You can start the PostgreSQL database container with Docker Compose:
```bash
make docker-up
# or: docker compose up -d
```

### 3. Configure Environment Variables
Copy `.env.example` to `.env` (already done for default local dev):
```env
PORT=8080
DATABASE_URL=postgres://postgres:postgres@localhost:5432/starter_db?sslmode=disable
JWT_SECRET=starter_jwt_access_secret_key_12345
JWT_REFRESH_SECRET=starter_jwt_refresh_secret_key_67890
ACCESS_TOKEN_EXPIRY=15m
REFRESH_TOKEN_EXPIRY=168h
```

### 4. Run Database Migrations
Create the database tables automatically:
```bash
make migrate
# or: go run apps/api/cmd/migrate/main.go
```

### 5. Start the Server
Launch the API server:
```bash
make run
# or: go run apps/api/cmd/server/main.go
```
The server will start listening at `http://localhost:8080`.

### 6. Run Tests
Execute the unit and integration tests:
```bash
make test
# or: go test ./apps/api/... -v
```

---

## 📡 API Endpoints

### Response Wrapper Schema
Every response returns the standard `APIResponse`:
```json
{
  "statusCode": <int>,
  "data": <object>,
  "isSuccess": <bool>,
  "message": <string>
}
```

### 1. Health Check
- **Endpoint**: `GET /health`
- **Access**: Public
- **Response**:
```json
{
  "statusCode": 200,
  "data": { "status": "UP" },
  "isSuccess": true,
  "message": "Service is healthy"
}
```

---

### 2. Sign Up
- **Endpoint**: `POST /api/v1/auth/sign-up`
- **Access**: Public
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "Password123!",
  "name": "Sahil Karwasra"
}
```
- **Response (201 Created)**:
```json
{
  "statusCode": 201,
  "data": {
    "accessToken": "eyJhbGciOi...",
    "refreshToken": "eyJhbGciOi...",
    "user": {
      "id": "e4b1a43a-7965-4f35-9f5b-6f8d0ab9182a",
      "email": "user@example.com",
      "name": "Sahil Karwasra",
      "createdAt": "2026-09-26T23:30:00Z",
      "updatedAt": "2026-09-26T23:30:00Z"
    }
  },
  "isSuccess": true,
  "message": "User registered successfully"
}
```

---

### 3. Sign In
- **Endpoint**: `POST /api/v1/auth/sign-in`
- **Access**: Public
- **Request Body**:
```json
{
  "email": "user@example.com",
  "password": "Password123!"
}
```
- **Response (200 OK)**:
```json
{
  "statusCode": 200,
  "data": {
    "accessToken": "eyJhbGciOi...",
    "refreshToken": "eyJhbGciOi...",
    "user": {
      "id": "e4b1a43a-7965-4f35-9f5b-6f8d0ab9182a",
      "email": "user@example.com",
      "name": "Sahil Karwasra",
      "createdAt": "2026-09-26T23:30:00Z",
      "updatedAt": "2026-09-26T23:30:00Z"
    }
  },
  "isSuccess": true,
  "message": "Signed in successfully"
}
```

---

### 4. Refresh Token
Refreshes the access token using a valid refresh token. Rotates the refresh token for security.
- **Endpoint**: `POST /api/v1/auth/refresh-token`
- **Access**: Public
- **Request Body**:
```json
{
  "refreshToken": "eyJhbGciOi..."
}
```
- **Response (200 OK)**:
```json
{
  "statusCode": 200,
  "data": {
    "accessToken": "eyJhbGciOi...(new)",
    "refreshToken": "eyJhbGciOi...(new)"
  },
  "isSuccess": true,
  "message": "Tokens refreshed successfully"
}
```

---

### 5. Get User Profile
- **Endpoint**: `GET /api/v1/auth/me`
- **Access**: Protected (`Authorization: Bearer <accessToken>`)
- **Response (200 OK)**:
```json
{
  "statusCode": 200,
  "data": {
    "id": "e4b1a43a-7965-4f35-9f5b-6f8d0ab9182a",
    "email": "user@example.com",
    "name": "Sahil Karwasra",
    "createdAt": "2026-09-26T23:30:00Z",
    "updatedAt": "2026-09-26T23:30:00Z"
  },
  "isSuccess": true,
  "message": "Profile retrieved successfully"
}
```

---

### 6. Logout
Revokes the refresh token in the database.
- **Endpoint**: `POST /api/v1/auth/logout`
- **Access**: Protected (`Authorization: Bearer <accessToken>`)
- **Response (200 OK)**:
```json
{
  "statusCode": 200,
  "data": {},
  "isSuccess": true,
  "message": "Logged out successfully"
}
```
