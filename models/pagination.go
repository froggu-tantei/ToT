package models

// PaginatedResponse provides a standard structure for paginated data
type PaginatedResponse struct {
	Success    bool       `json:"success"`
	Data       any        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination holds pagination metadata
type Pagination struct {
	Total       int `json:"total"`
	PerPage     int `json:"per_page"`
	CurrentPage int `json:"current_page"`
	LastPage    int `json:"last_page"`
	From        int `json:"from"`
	To          int `json:"to"`
}

// NewPaginatedResponse creates a standard paginated response
func NewPaginatedResponse(data any, total, perPage, currentPage int) PaginatedResponse {
	// Calculate last page (ceiling division)
	lastPage := (total + perPage - 1) / perPage

	// Calculate from/to indexes
	from := (currentPage-1)*perPage + 1
	to := from + perPage - 1

	if total == 0 {
		from = 0
		to = 0
	} else if to > total {
		to = total
	}

	return PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Total:       total,
			PerPage:     perPage,
			CurrentPage: currentPage,
			LastPage:    lastPage,
			From:        from,
			To:          to,
		},
	}
}
