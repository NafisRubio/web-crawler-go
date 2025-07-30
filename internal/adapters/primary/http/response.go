package http

// Response represents the response structure for the API
type Response struct {
	Status     string      `json:"status"`
	Message    string      `json:"message,omitempty"`
	Data       interface{} `json:"data"` // Use interface{} to handle different data types
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Pagination represents the pagination metadata
type Pagination struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalItems int64  `json:"total_items"`
	TotalPages int    `json:"total_pages"`
	NextPage   string `json:"next_page,omitempty"`
	PrevPage   string `json:"prev_page,omitempty"`
}
