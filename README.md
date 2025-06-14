# Realtime Chat Application Backend

A robust backend service for a real-time chat application built with Go, featuring WebSocket support for instant messaging, friend management, and authentication.

## Features

- **Real-time Communication**: WebSocket implementation for instant messaging
- **User Authentication**: Complete JWT-based authentication system
- **Friend Management**: Add, accept, reject friend requests and manage friendships
- **Chat Rooms**: Create and manage chat rooms for group conversations
- **File Storage**: Cloudinary integration for image and file uploads with signed uploads
- **Message Queuing**: Kafka integration for reliable message delivery and event processing
- **User Status Tracking**: Real-time online/offline status with "last seen" timestamps
- **Scalable Architecture**: Designed with clean architecture principles
- **Database Support**: MySQL for persistent storage with migration support
- **Caching**: Redis for improved performance
- **API Documentation**: Comprehensive Swagger documentation included

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
│                       Kafka Cluster                         │
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
5. **Kafka**: 
   - Handles messaging and event processing between server instances
   - Manages chat messages, notifications, and system events
   - Ensures reliable message delivery even during high load
6. **Data Layer**:
   - **MySQL**: Stores users, messages, chat rooms with sharding capability
   - **Redis Cache**: Session data, JWT tokens, user presence information
   - **Cloudinary**: Image and file storage

### WebSocket Message Flow

1. Client sends message via WebSocket to Server Instance X
2. Server Instance X validates and saves message to MySQL
3. Message is published to Kafka topic
4. Kafka distributes message to all consumer Server Instances
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

## File Storage

The application uses Cloudinary for secure file storage and management:

### Cloudinary Integration

- **Avatar uploads**: Profile picture management with automatic optimization
- **Chat attachments**: Images, documents, and other media files in chat
- **Secure uploads**: Client-side signed uploads for improved security
- **Folder structure**: Organized content with proper access controls
- **Image transformations**: On-the-fly resizing and optimization
- **Direct upload**: Client can upload directly to Cloudinary after server authentication

### Upload Process

1. Server generates a signed upload token with specific permissions
2. Client receives the token and uploads directly to Cloudinary
3. Cloudinary validates the signature and processes the upload
4. Server receives the upload notification and updates the database

## Tech Stack

- **Go 1.24+**: Main programming language with modern features and performance
- **MySQL 8.0+**: Primary relational database for persistent storage
- **Redis 6.0+**: Caching and session management
- **Kafka 3.0+**: Message queuing and distributed event streaming platform
- **WebSockets**: Real-time bidirectional communication protocol
- **Docker & Docker Compose**: Application containerization and orchestration
- **Nginx**: HTTP server and load balancer
- **Swagger**: API documentation with OpenAPI specification
- **JWT**: JSON Web Token based authentication
- **Cloudinary**: Cloud-based image and video management service
- **bcrypt**: Password hashing algorithm
- **Gorilla WebSocket**: WebSocket implementation for Go
- **GORM**: Object-relational mapping library for Go
- **go-redis**: Redis client for Go
- **Logrus**: Structured logging
- **Viper**: Configuration management

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
- Kafka 3.0+ (optional, for event-driven features)
- Cloudinary account (for file storage features)
- Docker and Docker Compose (recommended)

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

### Environment Configuration

The `.env` file should include the following configuration:

```env
# Server
SERVER_PORT=8080
SERVER_MODE=development

# Database
DB_HOST=localhost
DB_PORT=3306
DB_USER=root
DB_PASSWORD=yourpassword
DB_NAME=chat_app

# Redis
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key
JWT_EXPIRATION=24h

# Cloudinary
CLOUDINARY_CLOUD_NAME=your-cloud-name
CLOUDINARY_API_KEY=your-api-key
CLOUDINARY_API_SECRET=your-api-secret

# Kafka (optional)
KAFKA_BROKERS=localhost:9092
KAFKA_CHAT_TOPIC=chat-events
KAFKA_STATUS_TOPIC=status-events
KAFKA_CONSUMER_GROUP=chat-consumer-group
```

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

## Message Queuing System

The application relies on Kafka as the primary message queuing system for reliable message delivery and event processing:

### Kafka Integration

- **Producer/Consumer Model**: Asynchronous communication between system components
- **Event Types**:
  - `message_sent`: New messages in chat rooms
  - `typing_started`/`typing_stopped`: User typing indicators
  - `user_online`/`user_offline`: User presence events
  - `user_joined_room`/`user_left_room`: Room participation events
- **Benefits**:
  - Guaranteed message delivery with persistence
  - High throughput message processing
  - Horizontal scalability for handling traffic spikes
  - Message replay capabilities for system recovery
  - Topic partitioning for ordered message processing
- **Implementation**: Producer/Consumer implementations with error handling and retry mechanisms

### Message Flow

1. Events generated in the application are published to Kafka topics
2. Consumer groups process events based on their responsibilities
3. Events are partitioned by chat room ID for ordered processing
4. Consumers implement idempotent operations for reliability

## Authentication

The API uses JWT-based authentication. Include the JWT token in the Authorization header:

```
Authorization: Bearer <your_token>
```

## License

[MIT License](LICENSE)

## Contact

For questions or support, please open an issue in the repository.
