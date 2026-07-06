package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ── Response envelope ────────────────────────────────────────────────────────

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Meta    *PaginationMeta `json:"meta,omitempty"`
}

// ── Success helpers ──────────────────────────────────────────────────────────

func OK(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Success: true, Message: message, Data: data})
}

func Created(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusCreated, Response{Success: true, Message: message, Data: data})
}

func OKPaginated(c *gin.Context, data interface{}, meta PaginationMeta) {
	c.JSON(http.StatusOK, Response{Success: true, Data: data, Meta: &meta})
}

// ── Error helpers ────────────────────────────────────────────────────────────

func BadRequest(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, Response{Success: false, Error: err})
}

func Unauthorized(c *gin.Context, err string) {
	c.JSON(http.StatusUnauthorized, Response{Success: false, Error: err})
}

func Forbidden(c *gin.Context) {
	c.JSON(http.StatusForbidden, Response{Success: false, Error: "you do not have permission to perform this action"})
}

func NotFound(c *gin.Context, resource string) {
	c.JSON(http.StatusNotFound, Response{Success: false, Error: resource + " not found"})
}

func Conflict(c *gin.Context, err string) {
	c.JSON(http.StatusConflict, Response{Success: false, Error: err})
}

func InternalError(c *gin.Context, err error) {
	// Never expose internal error details to clients
	c.JSON(http.StatusInternalServerError, Response{Success: false, Error: "internal server error"})
}
