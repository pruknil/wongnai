package model

// RestaurantStatus represents the opening status of a restaurant
type RestaurantStatus struct {
	RestaurantID string `json:"restaurant_id"`
	Name         string `json:"name"`
	IsOpen       bool   `json:"is_open"`
	Status       string `json:"status"` // e.g., "เปิดอยู่", "ปิดแล้ว"
	OpenUntil    string `json:"open_until,omitempty"`
	Message      string `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}
