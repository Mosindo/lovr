package handlers

import (
	"log"

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
		err.Error(),
	)
}

func logHandlerEvent(c *gin.Context, operation string, fields map[string]string) {
	requestID := c.GetString("request_id")
	userID := c.GetString("userID")
	details := ""
	for key, value := range fields {
		details += `,"` + key + `":` + `"` + value + `"`
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
