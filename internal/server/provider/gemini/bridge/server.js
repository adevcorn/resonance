#!/usr/bin/env node

/**
 * Gemini CLI Bridge Server
 * 
 * This Node.js service provides a bridge between the Go-based Ensemble server
 * and the ai-sdk-provider-gemini-cli package. It handles streaming completions
 * with Gemini models using CLI authentication.
 */

import { createServer } from 'http';
import { createGeminiProvider } from 'ai-sdk-provider-gemini-cli';
import { streamText } from 'ai';

const PORT = process.env.GEMINI_BRIDGE_PORT || 3001;

// Initialize Gemini provider with OAuth authentication
const gemini = createGeminiProvider({
  authType: 'oauth-personal',
  // Can be configured via environment variables or config
});

/**
 * Handle completion requests from Go
 */
async function handleCompletion(requestBody) {
  const { model, messages, tools, temperature, maxTokens } = requestBody;
  
  // Convert messages to AI SDK format
  // The AI SDK expects: system, user, assistant (with optional toolCalls/toolResults)
  // Tool results should be part of assistant messages, not separate "tool" role messages
  const formattedMessages = messages
    .filter(msg => msg.role !== 'tool')  // Filter out tool role messages
    .map(msg => {
      const formatted = {
        role: msg.role === 'model' ? 'assistant' : msg.role,
        content: msg.content || '',  // Ensure content is never undefined
      };
      
      // Add tool calls if present
      if (msg.toolCalls && msg.toolCalls.length > 0) {
        formatted.toolCalls = msg.toolCalls;
      }
      
      // Add tool results if present  
      if (msg.toolResults && msg.toolResults.length > 0) {
        formatted.toolResults = msg.toolResults;
      }
      
      return formatted;
    });
  
  // Convert tools to AI SDK format
  const formattedTools = tools ? tools.reduce((acc, tool) => {
    acc[tool.name] = {
      description: tool.description,
      // Parameters is already an object when sent from Go (json.RawMessage unmarshals)
      parameters: tool.parameters || {}
    };
    return acc;
  }, {}) : undefined;
  
  // Stream completion
  const result = await streamText({
    model: gemini(model),
    messages: formattedMessages,
    tools: formattedTools,
    temperature: temperature || 0.7,
    maxTokens: maxTokens || 4096,
  });
  
  return result;
}

/**
 * Create HTTP server to handle requests from Go
 */
const server = createServer(async (req, res) => {
  // Enable CORS
  res.setHeader('Access-Control-Allow-Origin', '*');
  res.setHeader('Access-Control-Allow-Methods', 'POST, OPTIONS');
  res.setHeader('Access-Control-Allow-Headers', 'Content-Type');
  
  if (req.method === 'OPTIONS') {
    res.writeHead(200);
    res.end();
    return;
  }
  
  if (req.method === 'POST' && req.url === '/v1/completions') {
    let body = '';
    
    req.on('data', chunk => {
      body += chunk.toString();
    });
    
    req.on('end', async () => {
      try {
        const requestData = JSON.parse(body);
        
        console.log('Received completion request:', {
          model: requestData.model,
          messageCount: requestData.messages?.length || 0,
          toolCount: requestData.tools?.length || 0
        });
        
        // Handle streaming response
        res.writeHead(200, {
          'Content-Type': 'text/event-stream',
          'Cache-Control': 'no-cache',
          'Connection': 'keep-alive'
        });
        
        try {
          const result = await handleCompletion(requestData);
          
          // Stream text chunks
          for await (const chunk of result.textStream) {
            const event = {
              type: 'content',
              content: chunk,
            };
            res.write(`data: ${JSON.stringify(event)}\n\n`);
          }
          
          // Get tool calls (it's a Promise, not an async iterable)
          const toolCalls = await result.toolCalls;
          if (toolCalls && toolCalls.length > 0) {
            for (const toolCall of toolCalls) {
              const event = {
                type: 'tool_call',
                toolCall: {
                  id: toolCall.toolCallId,
                  toolName: toolCall.toolName,
                  arguments: toolCall.args
                }
              };
              res.write(`data: ${JSON.stringify(event)}\n\n`);
            }
          }
          
          // Send final usage stats
          const usage = await result.usage;
          const doneEvent = {
            type: 'done',
            usage: {
              inputTokens: usage.promptTokens,
              outputTokens: usage.completionTokens,
              totalTokens: usage.totalTokens
            }
          };
          res.write(`data: ${JSON.stringify(doneEvent)}\n\n`);
          res.end();
          
        } catch (streamError) {
          console.error('Error during streaming:', streamError);
          
          // Send error as SSE event
          const errorMessage = streamError instanceof Error 
            ? streamError.message 
            : String(streamError);
            
          const errorEvent = {
            type: 'error',
            error: errorMessage
          };
          res.write(`data: ${JSON.stringify(errorEvent)}\n\n`);
          res.end();
        }
        
      } catch (error) {
        console.error('Error parsing request or initializing stream:', error);
        
        // Ensure error is properly serialized
        const errorMessage = error instanceof Error 
          ? error.message 
          : (typeof error === 'string' ? error : JSON.stringify(error));
        
        // For errors before streaming starts, send HTTP error response
        if (!res.headersSent) {
          res.writeHead(500, { 'Content-Type': 'application/json' });
          res.end(JSON.stringify({
            error: errorMessage,
            type: error.name || 'Error',
            stack: process.env.NODE_ENV === 'development' ? error.stack : undefined
          }));
        } else {
          // If headers already sent, send error as SSE
          const errorEvent = {
            type: 'error',
            error: errorMessage
          };
          res.write(`data: ${JSON.stringify(errorEvent)}\n\n`);
          res.end();
        }
      }
    });
  } else if (req.method === 'GET' && req.url === '/health') {
    res.writeHead(200, { 'Content-Type': 'application/json' });
    res.end(JSON.stringify({ status: 'ok', service: 'gemini-cli-bridge' }));
  } else {
    res.writeHead(404);
    res.end('Not Found');
  }
});

server.listen(PORT, () => {
  console.log(`Gemini CLI Bridge running on port ${PORT}`);
  console.log(`Health check: http://localhost:${PORT}/health`);
});

// Handle graceful shutdown
process.on('SIGTERM', () => {
  console.log('Shutting down gracefully...');
  server.close(() => {
    console.log('Server closed');
    process.exit(0);
  });
});
