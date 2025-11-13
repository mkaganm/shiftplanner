# Shift Planner - React Frontend

## Project Structure

```
frontend-react/
├── src/
│   ├── components/          # Reusable UI components
│   │   ├── MembersSection.jsx
│   │   ├── StatisticsSection.jsx
│   │   ├── PlanningSection.jsx
│   │   ├── CalendarSection.jsx
│   │   └── ProtectedRoute.jsx
│   ├── pages/               # Page components
│   │   ├── Login.jsx
│   │   └── Dashboard.jsx
│   ├── context/             # React Context for state management
│   │   ├── AuthContext.jsx  # Authentication state
│   │   └── AppContext.jsx   # Application state
│   ├── services/            # API service layer
│   │   └── api.js           # API client
│   ├── App.jsx              # Main app component with routing
│   └── main.jsx             # Entry point
├── public/                  # Static assets
├── package.json
└── vite.config.js           # Vite configuration
```

## Getting Started

1. Install dependencies:
```bash
npm install
```

2. Start development server:
```bash
npm run dev
```

The app will run on http://localhost:3000

## Features

- **Authentication**: Login and Register pages
- **Protected Routes**: Routes that require authentication
- **State Management**: Context API for global state
- **API Integration**: Centralized API service layer
- **Responsive Design**: Corporate-themed UI

## Development

- Frontend runs on port 3000
- Backend API runs on port 8080
- Vite proxy configured to forward `/api` requests to backend
