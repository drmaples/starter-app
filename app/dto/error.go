package dto

// ErrorResponse is the response for an error
type ErrorResponse struct {
	Message string `json:"message"`
}

// NewErrorResp returns a new error response
func NewErrorResp(msg string) ErrorResponse {
	return ErrorResponse{Message: msg}
}
