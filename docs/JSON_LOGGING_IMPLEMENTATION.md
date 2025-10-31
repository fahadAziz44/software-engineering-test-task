# JSON Logging Implementation

## Overview

This document describes the implementation of structured JSON logging middleware for the CRUDER application.

## Implementation Summary

**Files Modified**:
- `internal/middleware/logger.go` - JSON logging middleware (created)
- `cmd/main.go` - Integrated middleware into application

Approach: Middleware-based logging using Go's standard `log/slog` package

---

## Technical Details


### Features Implemented

✅ **JSON Format**: All logs are valid JSON
✅ **Structured Attributes**: Consistent field names across all logs
✅ **Request Tracking**: Unique request ID for distributed tracing
✅ **Response Headers**: X-Request-ID header added to all responses
✅ **Context Integration**: Logger available in Gin context for controllers
✅ **Log Levels**: Automatic level selection based on HTTP status codes:
- INFO: 2xx status codes
- WARN: 4xx status codes
- ERROR: 5xx status codes and request failures

✅ **Performance Metrics**: Request latency tracking
✅ **Client Information**: IP address and user agent
✅ **Error Logging**: Gin errors captured and logged

---

## Log Format

### Log Fields

Each request generates a JSON log with the following fields:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `time` | string | ISO 8601 timestamp | `2025-10-31T16:58:03.518646+02:00` |
| `level` | string | Log level (INFO/WARN/ERROR) | `INFO` |
| `msg` | string | Log message | `Request completed` |
| `request_id` | string | Unique UUID for request tracing | `5ca149c4-a6cc-4fb4-a151-075828504e48` |
| `method` | string | HTTP method | `GET`, `POST`, `PATCH`, `DELETE` |
| `path` | string | Request path | `/api/v1/users` |
| `status_code` | int | HTTP status code | `200`, `201`, `404`, `400` |
| `latency` | int | Request duration in nanoseconds | `23959166` (≈24ms) |
| `client_ip` | string | Client IP address | `::1` (localhost) |
| `user_agent` | string | Client user agent | `curl/8.7.1` |
| `error` | string | Error message (only for failures) | `validation failed` |

---

## Example Logs

### 1. Successful GET Request (200 OK)
```json
{
  "time": "2025-10-31T16:58:03.518646+02:00",
  "level": "INFO",
  "msg": "Request completed",
  "request_id": "5ca149c4-a6cc-4fb4-a151-075828504e48",
  "method": "GET",
  "path": "/api/v1/users",
  "status_code": 200,
  "latency": 23959166,
  "client_ip": "::1",
  "user_agent": "curl/8.7.1"
}
```

### 2. Resource Not Found (404)
```json
{
  "time": "2025-10-31T16:58:04.78251+02:00",
  "level": "WARN",
  "msg": "Request completed",
  "request_id": "0725e209-19ae-4d1c-a4c2-36134ce0606c",
  "method": "GET",
  "path": "/api/v1/users/username/nonexistent",
  "status_code": 404,
  "latency": 15761667,
  "client_ip": "::1",
  "user_agent": "curl/8.7.1"
}
```

### 3. Resource Created (201)
```json
{
  "time": "2025-10-31T16:58:05.825295+02:00",
  "level": "INFO",
  "msg": "Request completed",
  "request_id": "ce6ff2f2-a716-4078-bade-9c18f34c5d8f",
  "method": "POST",
  "path": "/api/v1/users",
  "status_code": 201,
  "latency": 18072541,
  "client_ip": "::1",
  "user_agent": "curl/8.7.1"
}
```

### 4. Bad Request (400)
```json
{
  "time": "2025-10-31T16:58:06.899096+02:00",
  "level": "WARN",
  "msg": "Request completed",
  "request_id": "83632435-5051-4b31-8185-e298478d2193",
  "method": "GET",
  "path": "/api/v1/users/id/invalid-uuid",
  "status_code": 400,
  "latency": 58458,
  "client_ip": "::1",
  "user_agent": "curl/8.7.1"
}
```

### 5. Application Startup
```json
{
  "time": "2025-10-31T16:57:30.698141+02:00",
  "level": "INFO",
  "msg": "Starting server",
  "address": ":8080"
}
```

---

## Usage Examples

### Viewing Logs in Development

```bash
# Run application
make run

# Logs will appear in JSON format:
{"time":"...","level":"INFO","msg":"Starting server","address":":8080"}
{"time":"...","level":"INFO","msg":"Request completed","request_id":"...","method":"GET","path":"/api/v1/users","status_code":200,...}
```

### Viewing Logs in Docker

```bash
# Start with docker-compose
docker-compose up

# Follow logs
docker-compose logs -f app

# Logs are JSON formatted and easy to filter:
docker-compose logs app | grep '"level":"ERROR"'
docker-compose logs app | grep '"status_code":404'
```

### Using in Controllers

Controllers can access the request-specific logger from context:

```go
func (c *UserController) GetUserByID(ctx *gin.Context) {
	// Get request-specific logger with request_id
	logger, exists := ctx.Get("logger")
	if exists {
		reqLogger := logger.(*slog.Logger)

		// Log additional context
		reqLogger.Info("Fetching user",
			slog.String("user_id", id.String()),
		)
	}

	// ... rest of handler
}
```

