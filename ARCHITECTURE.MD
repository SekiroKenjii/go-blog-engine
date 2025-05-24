# Architecture

This document outlines the architecture of the Go Blog Engine project.

## Project Structure

```
go-blog-engine/
├── .github/              # GitHub workflows and templates
├── cmd/                  # Command-line tools
│   └── server/           # Main application entry point
│       └── main.go       # Server initialization
├── config/               # Configuration files
│   ├── dev/              # Development configurations
│   └── prod/             # Production configurations
├── core/                 # Core application logic
├── docs/                 # Documentation
│   └── api/              # API documentation
├── environments/         # Environment setup configurations
│   └── docker/           # Docker configuration files and Dockerfiles
├── internal/             # Internal packages (not importable)
│   ├── abstract/         # Interfaces and abstract types
│   ├── db/               # Database connections and models
│   │   ├── migrations/   # Database migration files
│   │   └── sqlc/         # Contains generated Go code for database interactions using SQLC, which creates type-safe Go code from SQL queries
│   ├── middlewares/      # HTTP middleware components
│   ├── modules/          # Business logic modules
│   │   ├── auth/         # Authentication module
│   │   ├── posts/        # Posts management module
│   │   └── users/        # User management module
│   ├── router/           # API routing
│   └── utils/            # Internal utilities
├── pkg/                  # Shared packages (importable)
│   ├── accessor/         # Context access utilities
│   ├── logger/           # Logging utilities
│   ├── response/         # HTTP response utilities
│   ├── scalar/           # Scalar API Docs
│   └── validator/        # Validation utilities
├── scripts/              # Utility scripts
│   ├── build/            # Build scripts
│   └── deploy/           # Deployment scripts
├── storage/              # Storage-related files
│   └── logs/             # Application logs
├── tests/                # Test files
│   ├── integration/      # Integration tests
│   └── unit/             # Unit tests
├── go.mod                # Go module definition
├── go.sum                # Go module checksums
├── Makefile              # Make targets
└── start.sh              # Startup script
```

## Architectural Patterns

The project follows a clean architecture approach with the following layers:

### 1. Presentation Layer (api/)

This layer is responsible for handling HTTP requests, validating inputs, and formatting responses. It includes:

- **Controllers**: Handle HTTP requests and responses
- **Middlewares**: Process requests before they reach controllers (auth, logging, etc.)
- **Routes**: Define API endpoints

### 2. Business Logic Layer (internal/services/)

This layer contains the core business logic of the application:

- Implements use cases
- Coordinates between different components
- Enforces business rules and validation
- Independent of external frameworks and databases

### 3. Data Access Layer (internal/repositories/)

This layer is responsible for data persistence and retrieval:

- Implements repository interfaces defined in the domain layer
- Handles database operations
- Translates between domain models and database models

### 4. Domain Layer (internal/models/)

This layer contains the core business entities and interfaces:

- Domain models representing business concepts
- Repository interfaces for data access
- Service interfaces defining business operations

## Request Flow

1. Client sends a request to a specific endpoint
2. Route is matched and middleware is applied (auth, logging, etc.)
3. Controller handles the request and validates input
4. Service layer processes business logic
5. Repository layer interacts with the database if needed
6. Response is formed and returned to the client

## Dependencies

- [Gin](https://github.com/gin-gonic/gin): Web framework
- [JWT-Go](https://github.com/golang-jwt/jwt): JWT authentication
- [go-playground/validator](https://github.com/go-playground/validator): Request validation
- [Swaggo](https://github.com/swaggo/swag): Swagger documentation generation

## Database Schema

The database schema includes the following main tables:

- **users**: User accounts and authentication information
- **posts**: Blog post content and metadata
- **categories**: Post categorization
- **tags**: Post tagging
- **comments**: User comments on posts
- **media**: Uploaded images and files

## Authentication Flow

The application uses JWT (JSON Web Tokens) for authentication:

1. User provides credentials (username/email and password)
2. Server validates credentials against the database
3. On successful validation, server issues a signed JWT
4. Client includes the JWT in subsequent requests
5. Server validates the JWT on protected routes

## Caching Strategy

The application uses a multi-level caching strategy:

1. In-memory cache for frequently accessed data
2. Redis for distributed caching across multiple instances
3. HTTP caching headers for client-side caching
