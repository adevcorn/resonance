---
name: web-access
description: Fetch URLs and search the web for information
category: capability
capabilities:
  - fetch_url
  - web_search
---

# Web Access

Retrieve content from the web using HTTP requests and search.

## Fetching URLs

### When to Use
- Retrieve API documentation
- Download specifications
- Check API responses
- Fetch web pages
- Access REST endpoints

### How to Execute

```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {
    "url": "https://example.com/api/data"
  }
}
```

### Returns

```json
{
  "url": "https://example.com/api/data",
  "status_code": 200,
  "headers": {
    "content-type": "application/json",
    "content-length": "1234"
  },
  "body": "{\"key\": \"value\"}"
}
```

### Important Notes
- Only HTTP and HTTPS protocols supported
- 30 second timeout
- Response body limited to 1MB
- Automatically sets User-Agent header
- Returns status code, headers, and body

### Examples

Fetch API documentation:
```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {"url": "https://api.example.com/docs"}
}
```

Check API endpoint:
```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {"url": "https://api.example.com/v1/status"}
}
```

Download specification:
```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {"url": "https://raw.githubusercontent.com/user/repo/main/spec.yaml"}
}
```

### Common Use Cases

**Check API Health:**
```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {"url": "https://api.service.com/health"}
}
```

**Fetch OpenAPI Spec:**
```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {"url": "https://petstore.swagger.io/v2/swagger.json"}
}
```

**Get Latest Release Info:**
```json
{
  "action": "execute",
  "capability": "fetch_url",
  "parameters": {"url": "https://api.github.com/repos/owner/repo/releases/latest"}
}
```

## Web Search

### Status
**Currently a placeholder implementation.** To enable real web search, configure a search API provider (Google Custom Search, Brave Search, DuckDuckGo, etc.).

### How to Execute

```json
{
  "action": "execute",
  "capability": "web_search",
  "parameters": {
    "query": "golang error handling best practices",
    "limit": 5
  }
}
```

### Parameters
- `query` (required): Search query string
- `limit` (optional): Maximum results (default: 5, max: 10)

### Returns (when configured)

```json
{
  "query": "golang error handling best practices",
  "results": [
    {
      "title": "Error Handling in Go",
      "url": "https://example.com/go-errors",
      "snippet": "Best practices for handling errors in Go..."
    }
  ]
}
```

### Configuration Required

To enable web search, you need to:
1. Choose a search API provider (Google, Brave, DuckDuckGo)
2. Get API credentials
3. Update the WebSearchCapability implementation
4. Configure API keys in server config

## Workflow Examples

### Researching an API

1. **Search for the skill**:
   ```json
   {"action": "search_skills", "query": "fetch url web"}
   ```

2. **Load this skill**:
   ```json
   {"action": "load_skill", "skill_name": "web-access"}
   ```

3. **Fetch API documentation**:
   ```json
   {"action": "execute", "capability": "fetch_url", "parameters": {"url": "https://api.example.com/docs"}}
   ```

4. **Test an endpoint**:
   ```json
   {"action": "execute", "capability": "fetch_url", "parameters": {"url": "https://api.example.com/v1/users"}}
   ```

### Checking Multiple Services

Check multiple service health endpoints:

```json
{"action": "execute", "capability": "fetch_url", "parameters": {"url": "https://api1.example.com/health"}}
```

```json
{"action": "execute", "capability": "fetch_url", "parameters": {"url": "https://api2.example.com/health"}}
```

```json
{"action": "execute", "capability": "fetch_url", "parameters": {"url": "https://api3.example.com/health"}}
```

## Understanding HTTP Status Codes

- **200**: Success - request completed successfully
- **201**: Created - resource created successfully
- **204**: No Content - success but no body returned
- **400**: Bad Request - invalid request parameters
- **401**: Unauthorized - authentication required
- **403**: Forbidden - no permission to access
- **404**: Not Found - resource doesn't exist
- **500**: Server Error - server-side error
- **503**: Service Unavailable - server overloaded or down

## Security Considerations

- Only fetches data, doesn't POST/PUT/DELETE
- Uses GET method only
- 1MB response size limit to prevent memory issues
- 30 second timeout to prevent hanging
- User-Agent header identifies as Ensemble-Agent
- No authentication headers sent automatically

## Best Practices

### URL Validation
- Always use HTTPS when possible
- Verify URLs are from trusted sources
- Check status codes before parsing body
- Handle errors gracefully

### Response Handling
- Check status_code first (200 = success)
- Parse headers for content-type
- Validate JSON before parsing
- Handle large responses appropriately

### Error Handling
- Expect network errors (timeouts, DNS failures)
- Handle non-200 status codes
- Parse error messages from body when available
- Retry failed requests if appropriate

## Troubleshooting

**Timeout errors**
- URL takes >30 seconds to respond
- Try a different endpoint
- Check if service is online
- Increase timeout if possible

**403 Forbidden**
- API requires authentication
- May need API key or token
- Check API documentation

**SSL/TLS errors**
- Certificate validation failed
- Use HTTPS URLs only
- Check certificate validity

**Response too large**
- Response exceeds 1MB limit
- Use pagination parameters
- Filter response fields
- Download in chunks if API supports it

## Related Skills

- **shell-execution**: For using curl or wget commands
- **filesystem-operations**: For saving fetched content to files

## Future Enhancements

Planned capabilities:
- POST/PUT/DELETE methods
- Custom headers support
- Authentication (Bearer tokens, API keys)
- Request body support
- Streaming for large responses
- Real web search integration
