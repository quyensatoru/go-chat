package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
}

// 200 OK
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

// 201 Created
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{
		Success: true,
		Data:    data,
	})
}

// 400 Bad Request
func BadRequest(c *gin.Context, err interface{}) {
	c.JSON(http.StatusBadRequest, Response{
		Success: false,
		Error:   err,
	})
}

// 500 Internal Server Error
func InternalError(c *gin.Context, err interface{}) {
	c.JSON(http.StatusInternalServerError, Response{
		Success: false,
		Error:   err,
	})
}

// Forbidden
func Forbidden(c *gin.Context, m string) {
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": m})
}

// 404 Not Found
func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{
		Success: false,
		Message: message,
	})
}
