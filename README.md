# WobotAi

WobotAi is a simple and efficient Real-Time Tunnel Relay Service. It helps in sending messages instantly between different devices or users connected to it.

## What it does
- **Real-Time Communication**: Connects multiple users (clients) at once.
- **Instant Messaging**: Messages sent are delivered immediately.
- **Broadcasting**: Send a message to everyone at once.
- **Direct Messaging**: Send a message privately to a specific user.

## How to Run
You can start the server with a simple command:
```bash
go run .
```

## How to Use

### 1. Check Server Status
Open your browser or use curl to see if the server is running:
`http://localhost:8080/ping`

### 2. Connect to WebSocket
Connect your application or a WebSocket client (like Postman) to:
`ws://localhost:8080/ws`

### 3. See Connected Clients
Check who is currently connected:
`http://localhost:8080/clients`

## Message Format
To send a message, use this JSON format:

**Broadcast (to everyone):**
```json
{
    "type": "broadcast",
    "content": "Hello Everyone!"
}
```

**Direct Message (to specific person):**
```json
{
    "type": "direct",
    "target": "CLIENT_ID_HERE",
    "content": "Secret message for you"
}
```