# Metrics Collection Implementation

This document describes the metrics collection system added to the Ensemble project.

## Overview

The metrics collection system tracks performance, token usage, and context management across the orchestration engine. It uses Prometheus for metrics collection and exposition.

## Features

### 1. **Token Tracking**
- **Pre-request estimation**: Estimates token count before sending to LLM (4 chars ≈ 1 token)
- **Post-response actual usage**: Captures actual token usage from provider responses
- **Per-message tracking**: Stores token usage in message metadata
- **Per-session aggregation**: Tracks cumulative tokens per session
- **Per-agent breakdown**: Monitors token usage by individual agents

### 2. **Context Management**
- **Message count tracking**: Monitors number of messages in conversation
- **Context size estimation**: Estimates total tokens in context
- **System prompt size**: Tracks system prompt token count
- **Context pruning**: Automatically prunes messages when exceeding limit (50 messages)
- **Pruning metrics**: Logs when context is pruned

### 3. **Performance Metrics**
- **Orchestration timing**: Measures total orchestration duration
- **Team assembly**: Tracks time to assemble agent teams
- **Moderator decisions**: Measures decision latency
- **Agent turn duration**: Times individual agent responses
- **Tool execution**: Tracks tool execution times (client vs server)

### 4. **Collaboration Metrics**
- **Collaboration messages**: Counts messages between agents
- **Message distribution**: Tracks sender/receiver patterns
- **Collaboration duration**: Measures time for collaboration actions

### 5. **Infrastructure Metrics**
- **HTTP requests**: Request counts, duration, status codes
- **Session lifecycle**: Creation, completion rates
- **Tool execution**: Success/failure rates by tool
- **Token efficiency**: Output/input token ratio

## Metrics Endpoint

Metrics are exposed at `/api/metrics` in Prometheus format.

```bash
curl http://localhost:8080/api/metrics
```

## Key Metrics

### Token Metrics
- `ensemble_tokens_total_per_session{session_id}` - Total tokens per session
- `ensemble_tokens_input_per_turn{agent}` - Input tokens per turn (histogram)
- `ensemble_tokens_output_per_turn{agent}` - Output tokens per turn (histogram)
- `ensemble_tokens_per_message{role,agent}` - Tokens per message (histogram)
- `ensemble_token_efficiency_ratio{agent}` - Output/input ratio (histogram)

### Context Metrics
- `ensemble_context_size_messages{agent}` - Message count in context (gauge)
- `ensemble_context_size_tokens{agent}` - Estimated token count in context (gauge)
- `ensemble_system_prompt_size_tokens{agent}` - System prompt size (histogram)

### Performance Metrics
- `ensemble_orchestration_starts_total` - Total orchestration starts (counter)
- `ensemble_orchestration_completes_total{status}` - Completions by status (counter)
- `ensemble_turn_duration_seconds{agent}` - Agent turn duration (histogram)
- `ensemble_team_assembly_duration_seconds` - Team assembly time (histogram)
- `ensemble_moderator_decision_duration_seconds` - Moderator decision time (histogram)

### HTTP Metrics
- `ensemble_http_requests_total{method,path,status}` - HTTP request counter
- `ensemble_http_request_duration_seconds{method,path}` - Request duration (histogram)

### Tool Metrics
- `ensemble_tool_executions_total{tool,side,status}` - Tool execution counter
- `ensemble_tool_execution_duration_seconds{tool,side}` - Tool execution time (histogram)

### Session Metrics
- `ensemble_sessions_created_total` - Session creation counter
- `ensemble_sessions_completed_total{status}` - Session completion counter

### Collaboration Metrics
- `ensemble_collaboration_messages_total{sender,receiver}` - Collaboration message counter
- `ensemble_collaboration_message_duration_seconds{sender}` - Collaboration duration (histogram)

## Implementation Details

### Token Estimation

Token estimation uses a simple heuristic:
```go
estimatedTokens = len(text) / 4.0
```

This provides a rough estimate before sending requests to the LLM. Actual usage is captured from provider responses.

### Token Attachment to Messages

Token usage is stored in message metadata:
```json
{
  "id": "msg_123",
  "role": "assistant",
  "agent": "developer",
  "content": "...",
  "metadata": {
    "tokens": {
      "input_tokens": 100,
      "output_tokens": 50,
      "total_tokens": 150
    }
  }
}
```

### Context Pruning

The orchestration engine maintains a maximum of 50 messages in the conversation history. When this limit is exceeded, older messages are pruned:

```go
if len(messages) > e.maxMessages {
    pruned := len(messages) - e.maxMessages
    messages = messages[pruned:]
}
```

### WebSocket Streaming

Token usage is streamed to clients via WebSocket:
```json
{
  "type": "agent_message",
  "payload": {
    "agent": "developer",
    "content": "...",
    "timestamp": "2026-01-20T12:00:00Z",
    "tokens": {
      "input_tokens": 100,
      "output_tokens": 50,
      "total_tokens": 150
    }
  }
}
```

## Usage Examples

### Viewing Metrics

```bash
# Get all metrics
curl http://localhost:8080/api/metrics

# Get specific metric
curl http://localhost:8080/api/metrics | grep ensemble_tokens_total
```

### Prometheus Configuration

Add to `prometheus.yml`:
```yaml
scrape_configs:
  - job_name: 'ensemble'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/api/metrics'
```

### Grafana Dashboard

Example PromQL queries:

**Token Usage by Agent**:
```promql
sum(rate(ensemble_tokens_total_per_session[5m])) by (session_id)
```

**Average Turn Duration**:
```promql
avg(ensemble_turn_duration_seconds) by (agent)
```

**Context Size Over Time**:
```promql
ensemble_context_size_tokens{agent="developer"}
```

## Files Modified

### New Files
- `internal/server/metrics/metrics.go` - Core metrics package
- `internal/server/metrics/metrics_test.go` - Metrics tests

### Modified Files
- `internal/server/orchestration/engine.go` - Added metrics instrumentation
- `internal/server/api/middleware.go` - HTTP metrics
- `internal/server/api/router.go` - Metrics endpoint
- `internal/server/api/websocket.go` - Token streaming
- `internal/protocol/message.go` - Added token usage field
- `internal/server/provider/openai/openai.go` - Token tracking
- `internal/server/provider/anthropic/anthropic.go` - Token tracking
- `go.mod` - Added Prometheus client dependency

## Future Enhancements

1. **Advanced Token Estimation**: Use tokenizer libraries (tiktoken) for accurate pre-request estimation
2. **Cost Tracking**: Calculate API costs based on token usage and provider pricing
3. **Alerting**: Set up Prometheus alerts for high token usage or long durations
4. **Dashboards**: Create pre-built Grafana dashboards
5. **Tracing**: Add OpenTelemetry tracing for distributed request tracking
6. **Budget Management**: Implement session-level token budgets
7. **Optimization Recommendations**: Suggest optimizations based on metrics patterns
