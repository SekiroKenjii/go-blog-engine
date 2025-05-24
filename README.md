# Go Blog Engine

A modern, high-performance blog engine written in Go using the Gin framework.

## Overview

Go Blog Engine is a lightweight, scalable blog platform designed for speed and simplicity. It provides a RESTful API for managing blog content, user authentication, and content delivery.

## Features

- RESTful API for content management
- User authentication and authorization
- Markdown support for blog posts
- Category and tag management
- Search functionality
- Image upload and management
- Responsive design
- Caching for improved performance

## Installation

### Prerequisites

- Go 1.20 or higher
- PostgreSQL (or your preferred database)
- Redis (optional, for caching)

### Getting Started

1. Clone the repository:

```bash
git clone https://github.com/yourusername/go-blog-engine.git
cd go-blog-engine
```

2. Install dependencies:

```bash
go mod download
```

3. Set up environment variables (create a `develop.yml` file and `docker.yml` file in `./config/env/` folder base on `example.yml` file)

4. Set up docker (create a `docker-compose.yml` file and `docker-compose-db.yml` file in `./environments/docker/` folder base on `docker-compose-example.yml` file)
   - docker-compose.yml: contains two services - a PostgreSQL database and an API service
   - docker-compose-db.yml: contains only the PostgreSQL database service

5. Generate swagger files

```bash
make swagger-gen
```

5. Run the application:

- start db:
```bash
make docker-build -db
```

- start Go API:
```bash
make server
```

### Documentation
- [Architecture Documentation](ARCHITECTURE.md)

### API Documentation
API documentation is available at /docs/swagger/index.html or /docs/scalar when running the server.

### License
This project is licensed under the MIT License - see the LICENSE file for details.
