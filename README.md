# hkChat

*hkChat* - is a real-time messaging web application. Users can register, select a chat room of interest, and start communicating with other participants.

## Interface

The application is implemented as a single-page interface, ensuring fast response times without page reloads. After registration and login, users gain access to the chat list and can exchange messages with others without revealing their identity.

## Technologies

### _Frontend_
The client side is built as a single-page application using React with JavaScript. It supports HTTPS and enhances security with cookies.
- React, JavaScript, HTML

### _Backend_
The backend consists of three microservices responsible for authentication, connection handling, and request proxying. Each request is processed asynchronously, achieving high efficiency.
- Golang, Nginx, WebSocket

### _Database_
Persistent user data (logins, passwords, etc.) is stored in PostgreSQL. Temporary data, such as authentication tokens, is stored in Redis.
- PostgreSQL, Redis

## Running the Application

The application is deployed in Docker containers using `docker compose`. Before running, you must fill in the required fields in the `datagate/.example.dbenv`, `postgres/postgres.config.example`, `redis/redis.config.example` files and remove ".example" from the file name.

Start the application from the project root with:
```bash
make up
```

To shut down the application, run:
```bash
make down
```