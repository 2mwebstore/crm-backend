package utils

import (
	"fmt"
	"math"
	"strings"

	"github.com/gin-gonic/gin"
)

// ParamUint extracts a uint path parameter from the Gin context.
// Returns 0 and false if the parameter is missing or not a valid uint.
func ParamUint(c *gin.Context, key string) (uint, bool) {
	val := c.Param(key)
	if val == "" {
		return 0, false
	}
	var id uint
	_, err := fmt.Sscanf(val, "%d", &id)
	return id, err == nil
}

// SortDir validates a sort direction value and returns "ASC" or "DESC".
func SortDir(dir string) string {
	if strings.ToUpper(dir) == "ASC" {
		return "ASC"
	}
	return "DESC"
}

// SanitizeSort ensures sortBy is in an allowlist to prevent SQL injection.
func SanitizeSort(sortBy string, allowed map[string]string, defaultCol string) string {
	if col, ok := allowed[sortBy]; ok {
		return col
	}
	return defaultCol
}

// PtrString returns a pointer to a string value.
func PtrString(s string) *string { return &s }

// PtrUint returns a pointer to a uint value.
func PtrUint(u uint) *uint { return &u }

// PtrBool returns a pointer to a bool value.
func PtrBool(b bool) *bool { return &b }

// RoundFloat rounds a float64 to the specified number of decimal places.
func RoundFloat(val float64, precision int) float64 {
	factor := math.Pow(10, float64(precision))
	return math.Round(val*factor) / factor
}

// MaskPhone masks a phone number showing only last 4 digits: *****8755
func MaskPhone(phone string) string {
	if len(phone) <= 4 {
		return "****"
	}
	masked := ""
	for i := 0; i < len(phone)-4; i++ {
		masked += "*"
	}
	return masked + phone[len(phone)-4:]
}
