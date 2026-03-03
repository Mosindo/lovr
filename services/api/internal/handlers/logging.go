package handlers

import (
	"log"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

func logHandlerError(c *gin.Context, operation string, err error) {
	requestID := c.GetString("request_id")
	userID := c.GetString("userID")
	log.Printf(
		`{"event":"handler_error","request_id":%q,"operation":%q,"method":%q,"path":%q,"user_id":%q,"error":%q}`,
		requestID,
		operation,
		c.Request.Method,
		c.FullPath(),
		userID,
		sanitizeLogValue(err.Error()),
	)
}

func logHandlerEvent(c *gin.Context, operation string, fields map[string]string) {
	requestID := c.GetString("request_id")
	userID := c.GetString("userID")
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
		`{"event":"handler_event","request_id":%q,"operation":%q,"method":%q,"path":%q,"user_id":%q%s}`,
		requestID,
		operation,
		c.Request.Method,
		c.FullPath(),
		userID,
		details,
	)
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
