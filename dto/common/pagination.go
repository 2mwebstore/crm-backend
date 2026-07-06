package common

// SortQuery holds common sort query params.
type SortQuery struct {
	SortBy  string `form:"sort_by"`
	SortDir string `form:"sort_dir"`
}
