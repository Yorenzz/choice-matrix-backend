package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func respondSuccess(c *gin.Context, status int, data any, message string) {
	if message == "" {
		message = http.StatusText(status)
	}

	c.JSON(status, gin.H{
		"code":    http.StatusOK,
		"data":    data,
		"message": message,
	})
}

func respondError(c *gin.Context, status int, message string) {
	if message == "" {
		message = http.StatusText(status)
	}

	c.JSON(status, gin.H{
		"code":    status,
		"message": message,
	})
}
