# Shift Planner - Go Backend

## Project Structure

```
backend/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── internal/                # Private application code
│   ├── api/                 # HTTP handlers
│   │   ├── handlers.go      # Main API handlers
│   │   ├── auth_handlers.go # Authentication handlers
│   │   └── middleware.go   # HTTP middleware
│   ├── auth/               # Authentication logic
│   │   └── auth.go         # Session management
│   ├── database/           # Database layer
│   │   └── db.go           # Database initialization
│   ├── models/             # Data models
│   │   ├── user.go
│   │   ├── member.go
│   │   ├── shift.go
│   │   └── holiday.go       # Holiday definitions
│   ├── scheduler/          # Business logic
│   │   └── planner.go       # Shift planning algorithm
│   └── storage/            # Data access layer
│       ├── storage.go      # CRUD operations
│       └── user_storage.go # User-related operations
├── pkg/                    # Public packages (if any)
├── go.mod                  # Go module definition
└── go.sum                  # Dependency checksums
```

## Getting Started

1. Install dependencies:
```bash
cd backend
go mod download
```

2. Run the server:
```bash
go run ./cmd/server
```

The server will run on http://localhost:8080

## API Endpoints

### Authentication (Public)
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `POST /api/auth/logout` - Logout user

### Members (Protected)
- `GET /api/members` - Get all members
- `POST /api/members` - Create member
- `DELETE /api/members/:id` - Delete member

### Shifts (Protected)
- `GET /api/shifts` - Get shifts (query: start_date, end_date)
- `POST /api/shifts/generate` - Generate shift plan

### Holidays (Public)
- `GET /api/holidays` - Get all holidays

### Statistics (Protected)
- `GET /api/stats` - Get shift statistics

## Architecture

- **Framework**: Fiber (Fast HTTP framework)
- **Database**: SQLite (modernc.org/sqlite - pure Go)
- **Authentication**: Token-based sessions
- **Project Layout**: Standard Go project layout

