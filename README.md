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

## System Design

### Architecture Overview

The system is designed to ensure high scalability and reliability:

```
┌─────────┐     ┌──────────────┐     ┌────────────────┐
│         │     │              │     │                │
│ Clients ├────►│ Load Balancer├────►│ API Gateway    │
│         │     │   (Nginx)    │     │                │
└─────────┘     └──────────────┘     └────────┬───────┘
                                              │
                                              ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│  ┌───────────────┐   ┌───────────────┐   ┌───────────────┐  │
│  │               │   │               │   │               │  │
│  │ Chat Server   │   │ Chat Server   │   │ Chat Server   │  │
│  │ Instance 1    │   │ Instance 2    │   │ Instance N    │  │
│  │               │   │               │   │               │  │
│  └───────┬───────┘   └───────┬───────┘   └───────┬───────┘  │
│          │                   │                   │          │
└──────────┼───────────────────┼───────────────────┼──────────┘
           │                   │                   │
           ▼                   ▼                   ▼
┌─────────────────────────────────────────────────────────────┐
│                                                             │
│                   Redis Pub/Sub Cluster                     │
│                                                             │
└─────────────────────────────────────────────────────────────┘
           │                   │                   │
           ▼                   ▼                   ▼
┌─────────────┐         ┌─────────────┐      ┌────────────────┐
│             │         │             │      │                │
│ MySQL       │         │ Redis Cache │      │ Cloudinary     │
│ Database    │         │             │      │ (File Storage) │
│             │         │             │      │                │
└─────────────┘         └─────────────┘      └────────────────┘
```

### Request Flow

1. **Client Layer**: Mobile apps, web browsers, desktop applications
2. **Load Balancer (Nginx)**: 
   - Distributes traffic evenly across server instances
   - Sticky sessions for WebSocket connections
   - Health checking and auto-healing
3. **API Gateway**: 
   - Request routing and authentication
   - Rate limiting and logging
   - API versioning
4. **Server Instances**: 
   - Multiple Go server instances handling REST API and WebSocket
   - Auto-scaling based on system load
5. **Redis Pub/Sub**: 
   - Synchronizes messages between server instances
   - Manages presence (online/offline) and typing status
6. **Data Layer**:
   - **MySQL**: Stores users, messages, chat rooms with sharding capability
   - **Redis Cache**: Session data, JWT tokens, recent messages
   - **Cloudinary**: Image and file storage

### WebSocket Message Flow

1. Client sends message via WebSocket to Server Instance X
2. Server Instance X validates and saves message to MySQL
3. Message is published to Redis Pub/Sub channel
4. Redis distributes message to all Server Instances
5. Each Server Instance forwards message to connected recipients
6. Push notifications sent to offline users

### Clean Architecture

The codebase follows Clean Architecture principles with clear separation:

1. **Interface Layer**: HTTP handlers, WebSocket controllers, middleware
2. **Usecase Layer**: Business logic for auth, chat, friends, profile
3. **Repository Layer**: Data access for MySQL and Redis
4. **Domain Layer**: Core entities (User, ChatRoom, Message, Friendship)

### Scalability Features

- **Horizontal Scaling**: Add server instances as load increases
- **Database Sharding**: Partition data by users or chat rooms
- **Read Replicas**: Separate read/write operations
- **Multi-level Caching**: Redis caching with appropriate TTL
- **Microservices Ready**: Can be split into smaller services

### Security

- JWT authentication with RSA-256
- Password encryption with bcrypt
- Rate limiting for brute-force protection
- CORS, XSS, and CSRF protection
- Input validation and sanitization

## Tech Stack

- **Go**: Main programming language
- **MySQL**: Primary database
- **Redis**: Caching and WebSocket message distribution
- **WebSockets**: Real-time communication
- **Docker**: Containerization
- **Nginx**: Load balancing
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

- Go 1.24+
- MySQL 8.0+
- Redis 6.0+
- Docker and Docker Compose (optional)

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/Realtime_Chat_App_Backend.git
   cd Realtime_Chat_App_Backend
   ```

2. Copy the environment file and configure it:
   ```bash
   cp .env.sample .env
   # Edit .env with your configuration
   ```

3. Start required services using Docker:
   ```bash
   docker-compose up -d
   ```

4. Run database migrations:
   ```bash
   make migrate-up
   ```

5. Start the server:
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
- `make debug`: Start the server in debug mode
- `make dev`: Start the server for development
- `make rand-user`: Seed the database with random users
- `make build`: Build the application
- `make test`: Run tests
- `make lint`: Run linter
- `make migrate-up`: Apply database migrations
- `make migrate-down`: Revert database migrations
- `make create-migration name=migration_name`: Create a new migration file
- `make swagger`: Generate Swagger documentation
- `make docker-build`: Build Docker images
- `make docker-up`: Start Docker containers

## WebSocket Communication

The application uses WebSockets for real-time messaging. Connect to the WebSocket endpoint after authentication:

```
ws://localhost:8080/ws
```

### Message Format

Messages are exchanged in JSON format with the following structure:

```json
{
  "type": "SEND_MESSAGE",
  "sender_id": "user-123",
  "timestamp": 1686123456789,
  "data": {
    "chat_room_id": "room-123",
    "content": "Hello world!",
    "mime_type": "text/plain"
  }
}
```

### Message Types

The server supports the following message types:

| Type | Direction | Description |
|------|-----------|-------------|
| SEND_MESSAGE | Client → Server | Send a new message to a chat room |
| JOIN_ROOM | Client → Server | Join a chat room for active viewing |
| LEAVE_ROOM | Client → Server | Leave a chat room's active view |
| TYPING | Client → Server | Indicate typing status |
| READ_RECEIPT | Client → Server | Mark messages as read |
| PING | Client → Server | Keep connection alive |
| NEW_MESSAGE | Server → Client | Notify about new messages |
| USERS | Server → Client | List of active users in a room |
| JOIN_SUCCESS | Server → Client | Room join confirmation |
| USER_JOINED | Server → Client | User joined room notification |
| USER_LEFT | Server → Client | User left room notification |

## User Status Management

The system tracks online/offline user status and stores it in Redis:
- Online users: Display "Active now"
- Offline users: Display "Last seen" with timestamp

## Caching and Performance

The system uses Redis for caching:
- Chat room details
- User's chat room lists
- Authentication tokens
- User status information

## Authentication

The API uses JWT-based authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your_token>
```

## License

[MIT License](LICENSE)

## Contact

For questions or support, please open an issue in the repository.
