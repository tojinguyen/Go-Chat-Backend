# Realtime Chat Application Backend

A robust backend service for a real-time chat application built with Go, featuring WebSocket support for instant messaging, friend management, and authentication.

## Features

- **Real-time Communication**: WebSocket implementation for instant messaging
- **User Authentication**: Complete JWT-based authentication system
- **Friend Management**: Add, accept, reject friend requests and manage friendships
- **Chat Rooms**: Create and manage chat rooms for group conversations
- **File Storage**: Cloudinary integration for image and file uploads
- **Scalable Architecture**: Designed with clean architecture principles
- **Database Support**: MySQL for persistent storage
- **Caching**: Redis for improved performance
- **API Documentation**: Swagger documentation included

## Tech Stack

- **Go**: Main programming language
- **MySQL**: Primary database
- **Redis**: Caching and WebSocket message distribution
- **WebSockets**: Real-time communication
- **Docker**: Containerization
- **Swagger**: API documentation
- **JWT**: Authentication
- **Cloudinary**: File storage
- **Kafka**: Message queuing (prepared infrastructure)

## Project Structure

```
├── cmd/                  # Application entry points
│   ├── seed/             # Database seeder
│   └── server/           # Main server application
├── config/               # Configuration files
├── docs/                 # API documentation (Swagger)
├── error/                # Error handling
├── internal/             # Internal application code
│   ├── domain/           # Domain models
│   ├── handler/          # HTTP handlers
│   ├── infra/            # Infrastructure layer
│   ├── middleware/       # HTTP middleware
│   ├── repository/       # Data access layer
│   ├── router/           # HTTP routing
│   ├── socket/           # WebSocket implementation
│   ├── usecase/          # Business logic
│   └── validations/      # Request validation
├── migrations/           # Database migrations
└── pkg/                  # Shared packages
```

## Getting Started

### Prerequisites

- Go 1.16+
- Docker and Docker Compose
- MySQL
- Redis

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/Realtime_Chat_App_Backend.git
   cd Realtime_Chat_App_Backend
   ```

2. Start required services using Docker:
   ```bash
   docker-compose up -d
   ```

3. Run database migrations:
   ```bash
   make migrate-up
   ```

4. Run the server:
   ```bash
   make run
   ```

The server will be available at `http://localhost:8080` by default.

### API Documentation

After starting the server, access the Swagger documentation at:
```
http://localhost:8080/swagger/index.html
```

## Development

### Available Make Commands

- `make run`: Start the server
- `make build`: Build the application
- `make test`: Run tests
- `make migrate-up`: Apply database migrations
- `make migrate-down`: Revert database migrations
- `make swagger`: Generate Swagger documentation

## WebSocket Communication

The application uses WebSockets for real-time messaging. Connect to the WebSocket endpoint:

```
ws://localhost:8080/ws
```

Messages are exchanged in JSON format with the following structure:

```json
{
  "type": "message",
  "data": {
    "room_id": "room-123",
    "content": "Hello world!",
    "file_url": "https://cloudinary.com/file.jpg" // Optional
  }
}
```

## Authentication

The API uses JWT-based authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your_token>
```

## License

[MIT License](LICENSE)

## Contact

For questions or support, please open an issue in the repository.
