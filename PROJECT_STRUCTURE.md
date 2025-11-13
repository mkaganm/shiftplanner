# Shift Planner - Project Structure

## Overview

This project follows a clean architecture with clear separation between frontend and backend.

## Frontend Structure (React + Vite)

```
frontend-react/
├── src/
│   ├── components/              # Reusable UI components
│   │   ├── MembersSection.jsx   # Team members management
│   │   ├── StatisticsSection.jsx # Statistics display
│   │   ├── PlanningSection.jsx  # Shift planning form
│   │   ├── CalendarSection.jsx   # Calendar view
│   │   └── ProtectedRoute.jsx   # Route protection
│   ├── pages/                   # Page components
│   │   ├── Login.jsx            # Login/Register page
│   │   └── Dashboard.jsx        # Main dashboard
│   ├── context/                 # React Context for state
│   │   ├── AuthContext.jsx      # Authentication state
│   │   └── AppContext.jsx       # Application state
│   ├── services/                # API service layer
│   │   └── api.js               # API client functions
│   ├── App.jsx                  # Main app with routing
│   └── main.jsx                 # Entry point
├── public/                      # Static assets
├── package.json                 # Dependencies
└── vite.config.js               # Vite configuration
```

### Key Features:
- **Component-based**: Modular, reusable components
- **Context API**: Global state management
- **Service Layer**: Centralized API calls
- **Protected Routes**: Authentication-based routing
- **CSS Modules**: Scoped styling per component

## Backend Structure (Go + Fiber)

```
backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/                    # Private application code
│   ├── api/                     # HTTP handlers
│   │   ├── handlers.go         # Main API handlers
│   │   ├── auth_handlers.go    # Auth endpoints
│   │   └── middleware.go      # Auth middleware
│   ├── auth/                    # Authentication logic
│   │   └── auth.go             # Session management
│   ├── database/                # Database layer
│   │   └── db.go               # DB initialization
│   ├── models/                  # Data models
│   │   ├── user.go
│   │   ├── member.go
│   │   ├── shift.go
│   │   └── holiday.go          # Holiday definitions
│   ├── scheduler/               # Business logic
│   │   └── planner.go          # Shift planning algorithm
│   └── storage/                 # Data access layer
│       ├── storage.go          # CRUD operations
│       └── user_storage.go     # User operations
├── pkg/                         # Public packages (if needed)
├── go.mod                       # Go module
└── go.sum                       # Dependency checksums
```

### Key Features:
- **Standard Layout**: Follows Go project layout conventions
- **Separation of Concerns**: Clear layers (API, Business Logic, Data Access)
- **Middleware**: Authentication middleware for protected routes
- **Pure Go**: Uses modernc.org/sqlite (no CGO)

## Development Workflow

### Frontend (React)
```bash
cd frontend-react
npm install
npm run dev          # Runs on http://localhost:3000
```

### Backend (Go)
```bash
cd backend
go run ./cmd/server  # Runs on http://localhost:8080
```

### Both
```bash
make dev             # Runs both (if Makefile configured)
```

## Architecture Principles

1. **Separation of Concerns**: Frontend and backend are completely separate
2. **API-First**: Backend provides RESTful API, frontend consumes it
3. **State Management**: React Context for global state
4. **Type Safety**: Go's type system + React PropTypes/TypeScript (optional)
5. **Scalability**: Modular structure allows easy expansion

