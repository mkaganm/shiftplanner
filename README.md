# Shift Planning System

Web application that creates fair shift schedules for team members, taking into account public holidays.

## Project Structure

```
shiftplanner/
├── backend/              # Backend code
│   ├── cmd/
│   │   └── server/       # Main application entry point
│   ├── internal/         # Internal packages (not accessible from outside)
│   │   ├── api/         # HTTP handlers
│   │   ├── database/    # Database connection and schema
│   │   ├── models/      # Data models
│   │   ├── scheduler/   # Shift planning algorithm
│   │   └── storage/     # Database operations
│   └── pkg/             # Public packages
├── frontend/            # Frontend files (HTML, CSS, JS)
├── data/                # Database file (auto-generated)
├── go.mod
└── go.sum
```

## Features

- ✅ Public holiday management (2025-2026)
- ✅ Weekend check
- ✅ Fair shift distribution (based on past shift days and long shift count)
- ✅ Long shift management (shifts before holidays/weekends)
- ✅ Calendar view
- ✅ Statistics
- ✅ User authentication (registration, login, session management)

## Installation

1. Go 1.24+ must be installed
2. Install dependencies:
   ```bash
   go mod download
   ```

## Running

To run the backend:

```bash
cd backend/cmd/server
go run main.go
```

or to run the compiled version:

```bash
cd backend/cmd/server
go build -o ../../../shiftplanner.exe .
../../../shiftplanner.exe
```

The server will run at `http://localhost:8080`.

## Usage

1. Go to `http://localhost:8080` in your browser
2. Register or login
3. Add team members
4. Select start and end dates
5. Click "Create Plan" button
6. View the shift plan in calendar view
7. View past shifts and statistics

## API Endpoints

### Authentication (Unprotected)
- `POST /api/auth/register` - User registration
- `POST /api/auth/login` - User login
- `POST /api/auth/logout` - User logout

### API (Protected - Authorization header required)
- `GET /api/members` - List all members
- `POST /api/members` - Add new member
- `DELETE /api/members/:id` - Delete member
- `GET /api/shifts` - Get shifts (with start_date, end_date query parameters)
- `POST /api/shifts/generate` - Create new shift plan
- `GET /api/holidays` - List public holidays
- `GET /api/stats` - Get member statistics

**Note:** All protected endpoints require a token in the `Authorization` header.

## Technologies

- **Backend**: Go (Golang)
- **Database**: SQLite
- **Frontend**: HTML, CSS, JavaScript (Vanilla)

## License

MIT
