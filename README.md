# Cautious Tree - AI Chat Session Management API

A Go-based REST API for managing AI chat sessions with authentication and channel-based conversation management.

## Overview

Cautious Tree is a robust backend service built in Go that provides a structured way to manage AI chat sessions, user authentication, and conversation channels. The project uses MongoDB for data persistence and implements JWT-based authentication.

## Features

- **User Management**
  - User registration and authentication
  - Email-based account activation
  - JWT token-based authentication
  - Password management and reset functionality

- **Channel Management**
  - Create and manage conversation channels
  - Tree-based session organization
  - Session history tracking
  - Multi-session support within channels

- **Session Management**
  - Create and manage chat sessions
  - Support for parent-child session relationships
  - Message history tracking
  - Context management between sessions

- **Security**
  - JWT-based secure authentication
  - Account activation through tokens
  - Password encryption
  - Error handling and validation

## Technical Stack

- **Language**: Go (97%)
- **Database**: MongoDB
- **Authentication**: JWT (JSON Web Tokens)
- **API Framework**: Custom HTTP router using `github.com/julienschmidt/httprouter`
- **Dependencies**:
  - `go.mongodb.org/mongo-driver` - MongoDB driver
  - `github.com/pascaldekloe/jwt` - JWT handling
  - Other internal packages for data models and validation

## API Endpoints

### Authentication
- `POST /v1/users` - Register new user
- `PUT /v1/users/activated` - Activate user account
- `PUT /v1/users/password` - Update user password
- `POST /v1/tokens/authentication` - Create authentication token
- `POST /v1/tokens/activations` - Create activation token
- `POST /v1/tokens/password-reset` - Create password reset token

### Channels
- `GET /v1/channels/:id` - Get channel information
- `GET /v1/channels/:id/sessions` - Get all sessions in a channel
- `POST /v1/channels/` - Create new channel

### Sessions
- `POST /v1/sessions/` - Create new session
- `GET /v1/sessions/:id` - Get session information
- `POST /v1/sessions/copy` - Copy existing session
- `PUT /v1/sessions/:id` - Append context to session
- `DELETE /v1/sessions/:id` - Delete session
- `GET /v1/sessions/:id/messages` - Get all messages in a session
- `POST /v1/sessions/message` - Send message in session

### System
- `GET /v1/health` - Health check endpoint

## Configuration

The application accepts the following configuration parameters:

- `port` - API server port (default: 4000)
- `env` - Environment (development/production)
- `mongo-uri` - MongoDB connection URI
- `db-name` - Database name
- `gemini-api-key` - Google Gemini API key
- `db-max-pool-size` - Maximum database pool size (default: 100)
- `db-min-pool-size` - Minimum database pool size (default: 10)
- `db-max-idle-time` - Maximum idle time for database connections (default: 15m)
- `jwt-secret` - Secret key for JWT token generation

## Getting Started

1. Clone the repository
2. Set up MongoDB instance
3. Configure environment variables
4. Build and run the application:
   ```bash
   make build
   ./bin/api
   ```

## Development

The project follows a clean architecture pattern with the following structure:
```
cautious-tree/
├── cmd/
│   └── api/
│       ├── main.go           # Application entry point and configuration
│       ├── server.go         # HTTP server implementation
│       ├── routes.go         # API route definitions
│       ├── middleware.go     # HTTP middleware functions
│       ├── errors.go         # Error handling utilities
│       ├── helpers.go        # Helper functions
│       ├── channels.go       # Channel-related handlers
│       ├── sessions.go       # Session-related handlers
│       ├── users.go          # User management handlers
│       ├── tokens.go         # Authentication token handlers
│       └── healthcheck.go    # Health check endpoint
│
├── internal/
│   ├── data/
│   │   ├── models.go        # Database models initialization
│   │   ├── users.go         # User model operations
│   │   ├── sessions.go      # Session model operations
│   │   ├── channels.go      # Channel model operations
│   │   ├── trees.go         # Tree structure model operations
│   │   └── tokens.go        # Token model operations
│   │
│   ├── validator/
│   │   └── validator.go     # Input validation utilities
│   │
│   ├── jsonlog/
│   │   └── jsonlog.go       # JSON logging functionality
│   │
│   └── llm/
│       └── gemini.go        # Google Gemini AI integration
│
└── Makefile                 # Build and development commands
```

## License

This project is licensed under appropriate open-source license.