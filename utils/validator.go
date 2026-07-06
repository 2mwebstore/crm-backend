package utils

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// BindJSON binds JSON body and returns a formatted error string on failure.
func BindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		return fmt.Errorf("%s", formatBindError(err))
	}
	return nil
}

// BindQuery binds query params and returns a formatted error string on failure.
func BindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		return fmt.Errorf("%s", formatBindError(err))
	}
	return nil
}

func formatBindError(err error) string {
	msg := err.Error()
	// Strip internal Go type paths for cleaner client messages
	if idx := strings.LastIndex(msg, "."); idx >= 0 {
		// keep as-is for simple messages
	}
	return msg
}
