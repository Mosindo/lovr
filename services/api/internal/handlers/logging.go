package handlers

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func logHandlerError(c *gin.Context, operation string, status int, err error) {
	requestID := c.GetString("request_id")
	userID := c.GetString("userID")
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	log.Printf(
		`{"event":"handler_error","request_id":%q,"operation":%q,"method":%q,"path":%q,"user_id":%q,"status":%d,"latency_ms":%d,"error":%q}`,
		requestID,
		operation,
		c.Request.Method,
		path,
		userID,
		status,
		requestLatencyMs(c),
		sanitizeLogValue(err.Error()),
	)
}

func logHandlerEvent(c *gin.Context, operation string, status int, fields map[string]string) {
	requestID := c.GetString("request_id")
	userID := c.GetString("userID")
	path := c.FullPath()
	if path == "" {
		path = c.Request.URL.Path
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	details := ""
	for _, key := range keys {
		details += `,"` + sanitizeLogValue(key) + `":` + `"` + sanitizeLogValue(fields[key]) + `"`
	}
	log.Printf(
		`{"event":"handler_event","request_id":%q,"operation":%q,"method":%q,"path":%q,"user_id":%q,"status":%d,"latency_ms":%d%s}`,
		requestID,
		operation,
		c.Request.Method,
		path,
		userID,
		status,
		requestLatencyMs(c),
		details,
	)
}

func requestLatencyMs(c *gin.Context) int64 {
	startedAt, ok := c.Get("request_started_at")
	if !ok {
		return 0
	}
	start, ok := startedAt.(time.Time)
	if !ok {
		return 0
	}
	return time.Since(start).Milliseconds()
}

func sanitizeLogValue(v string) string {
	trimmed := strings.TrimSpace(v)
	trimmed = strings.ReplaceAll(trimmed, `\`, `\\`)
	trimmed = strings.ReplaceAll(trimmed, `"`, `\"`)
	trimmed = strings.ReplaceAll(trimmed, "\n", " ")
	trimmed = strings.ReplaceAll(trimmed, "\r", " ")
	if len(trimmed) > 256 {
		return trimmed[:256]
	}
	return trimmed
}
