# Gemini CLI Bridge

This Node.js service provides a bridge between the Go-based Ensemble server and the `ai-sdk-provider-gemini-cli` package, enabling OAuth-based authentication with Google Gemini models.

## Why Use the CLI Bridge?

The Gemini CLI bridge offers several advantages:

1. **OAuth Authentication** - Use your Google account instead of API keys
2. **Gemini Code Assist Subscription** - Use your existing subscription
3. **AI SDK Integration** - Full access to Vercel AI SDK features
4. **No API Key Management** - Credentials managed by the CLI

## Setup

### 1. Install Gemini CLI Globally

```bash
npm install -g @google/gemini-cli
```

### 2. Authenticate with Google

```bash
gemini
# Follow the interactive authentication flow
# This will open your browser and prompt you to sign in with Google
```

### 3. Install Bridge Dependencies

```bash
cd internal/server/provider/gemini/bridge
npm install
```

### 4. Start the Bridge Service

```bash
npm start
# Bridge will start on http://localhost:3001
```

Or with custom port:

```bash
GEMINI_BRIDGE_PORT=3002 npm start
```

### 5. Configure Ensemble Server

Edit `config/server.yaml`:

```yaml
providers:
  gemini:
    use_cli: true                            # Enable CLI mode
    bridge_url: "http://localhost:3001"      # Bridge URL
    default_model: "gemini-2.0-flash-exp"
```

### 6. Start Ensemble Server

```bash
./bin/ensemble-server
```

## Architecture

```
┌─────────────────┐
│ Ensemble Server │ (Go)
│  (Port 8080)    │
└────────┬────────┘
         │ HTTP/SSE
         ▼
┌─────────────────┐
│ Gemini Bridge   │ (Node.js)
│  (Port 3001)    │
└────────┬────────┘
         │ ai-sdk-provider-gemini-cli
         ▼
┌─────────────────┐
│ Google Gemini   │
│    API          │
└─────────────────┘
```

## API Endpoints

### `POST /v1/completions`

Stream completions from Gemini models.

**Request:**
```json
{
  "model": "gemini-2.0-flash-exp",
  "messages": [
    { "role": "user", "content": "Hello!" }
  ],
  "temperature": 0.7,
  "maxTokens": 4096,
  "tools": []
}
```

**Response:** Server-Sent Events (SSE)

```
data: {"type":"content","content":"Hello!"}
data: {"type":"done","usage":{"inputTokens":5,"outputTokens":2,"totalTokens":7}}
```

### `GET /health`

Health check endpoint.

**Response:**
```json
{
  "status": "ok",
  "service": "gemini-cli-bridge"
}
```

## Authentication Options

The bridge supports multiple authentication methods via `ai-sdk-provider-gemini-cli`:

### OAuth (Default)
```javascript
const gemini = createGeminiProvider({
  authType: 'oauth-personal',
});
```

### API Key
```javascript
const gemini = createGeminiProvider({
  authType: 'api-key',
  apiKey: process.env.GEMINI_API_KEY,
});
```

### Vertex AI
```javascript
const gemini = createGeminiProvider({
  authType: 'vertex-ai',
  vertexAI: {
    projectId: 'my-project',
    location: 'us-central1',
  },
});
```

## Troubleshooting

### Bridge Won't Start

**Error:** `Cannot find module 'ai-sdk-provider-gemini-cli'`

**Solution:** Run `npm install` in the bridge directory

### Authentication Fails

**Error:** `No credentials found`

**Solution:** Run `gemini` to authenticate via the CLI

### Port Already in Use

**Error:** `EADDRINUSE: address already in use`

**Solution:** Change the port:
```bash
GEMINI_BRIDGE_PORT=3002 npm start
```

Then update `config/server.yaml`:
```yaml
gemini:
  bridge_url: "http://localhost:3002"
```

### Connection Refused

**Error:** `gemini-cli: bridge not available`

**Solution:** Ensure the bridge is running:
```bash
cd internal/server/provider/gemini/bridge
npm start
```

## Development

### Running in Development

```bash
# Terminal 1: Start bridge with auto-reload
cd internal/server/provider/gemini/bridge
npm install --save-dev nodemon
npx nodemon server.js

# Terminal 2: Start Ensemble server
./bin/ensemble-server
```

### Environment Variables

- `GEMINI_BRIDGE_PORT` - Port for the bridge server (default: 3001)
- `GEMINI_API_KEY` - API key if using api-key auth type

## Production Deployment

For production, consider:

1. **Process Manager**: Use PM2 or similar to manage the Node.js process
2. **Reverse Proxy**: Put the bridge behind nginx/caddy
3. **SSL/TLS**: Enable HTTPS for the bridge
4. **Health Checks**: Monitor `/health` endpoint
5. **Log Aggregation**: Collect logs from both Go and Node.js services

### Example PM2 Config

```javascript
// ecosystem.config.js
module.exports = {
  apps: [{
    name: 'gemini-bridge',
    script: './server.js',
    instances: 1,
    autorestart: true,
    watch: false,
    max_memory_restart: '200M',
    env: {
      NODE_ENV: 'production',
      GEMINI_BRIDGE_PORT: 3001
    }
  }]
};
```

Start with PM2:
```bash
pm2 start ecosystem.config.js
pm2 save
```

## References

- [Gemini CLI](https://www.npmjs.com/package/@google/gemini-cli)
- [ai-sdk-provider-gemini-cli](https://github.com/ben-vargas/ai-sdk-provider-gemini-cli)
- [Vercel AI SDK](https://sdk.vercel.ai/docs)
