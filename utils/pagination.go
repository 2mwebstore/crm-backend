package utils

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PaginationMeta is returned in every paginated response.
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}

// PaginationParams holds page + page_size parsed from query string.
type PaginationParams struct {
	Page     int
	PageSize int
}

// ParsePagination reads ?page= and ?page_size= from the request.
// Defaults: page=1, page_size=20. Max page_size=100.
func ParsePagination(c *gin.Context) PaginationParams {
	page := parseIntQuery(c, "page", 1)
	pageSize := parseIntQuery(c, "page_size", 20)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	return PaginationParams{Page: page, PageSize: pageSize}
}

// Paginate applies OFFSET + LIMIT to a GORM query.
func Paginate(p PaginationParams) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (p.Page - 1) * p.PageSize
		return db.Offset(offset).Limit(p.PageSize)
	}
}

// BuildMeta constructs PaginationMeta from a total count.
func BuildMeta(p PaginationParams, total int64) PaginationMeta {
	totalPages := int(math.Ceil(float64(total) / float64(p.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	return PaginationMeta{
		Page:       p.Page,
		PageSize:   p.PageSize,
		TotalItems: total,
		TotalPages: totalPages,
	}
}

func parseIntQuery(c *gin.Context, key string, fallback int) int {
	if v := c.Query(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
