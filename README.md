# Blog Post Generator AI Agent 📝

A conversational AI agent built with Go, LangChain, and Groq that generates professional blog posts through natural conversation.

## Features

🤖 **Telex.im Compatible** - Works as an AI coworker workflow

## Prerequisites

- Go 1.21 or higher
- Groq API key (get one free at [console.groq.com](https://console.groq.com/keys))

## Installation

### 1. Clone the repository

```bash
git clone <your-repo-url>
cd blog-generator
```

### 2. Install dependencies

```bash
go mod download
```

### 3. Set environment variables

```bash
export GROQ_API_KEY="gsk_your_key_here"
export PORT="8080"  # Optional, defaults to 8080
```

Or create a `.env` file:

```
GROQ_API_KEY=gsk_your_key_here
PORT=8080
```

### 4. Run the server

```bash
go run main.go
```

You should see:
```
🚀 Server running on port 8080
```

## Usage

### Testing with curl

Send a A2A-RPC request:

```bash
curl -X POST http://localhost:8080 \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": "req-1",
    "method": "message/send",
    "params": {
      "message": {
        "parts": [{"kind": "text", "text": "[title and content of blog"}],
        "taskId": "task-123"
      }
    }
  }'
```
**Response**
```
✅ Blog post generated!

[Full blog post content here...]

```

## Integrating with Telex.im

### 1. Deploy your agent

Deploy to any hosting platform (Heroku, Railway, Leapcell, etc.) and get your public URL.

### 2. Create a Telex workflow

In Telex.im, create a new workflow with this config:

```json
{
  "active": true,
  "category": "utilities",
  "description": "A workflow that creates blog posts",
  "id": "your-blogpost-id",
  "name": "blogpost_agent",
  "short_description": "Creates blog posts from title and content",
  "nodes": [
    {
      "id": "blogpost_agent",
      "name": "blogpost agent",
      "type": "a2a/mastra-a2a-node",
      "url": "https://your-deployed-url.com/agent"
    }
  ]
}
```
Chat with your agent 
