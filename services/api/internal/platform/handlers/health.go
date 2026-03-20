package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type DBPinger interface {
	Ping(ctx context.Context) error
}

func NewHealthHandler(db DBPinger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if db == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
			})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 2*time.Second)
		defer cancel()

		if err := db.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "unavailable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
