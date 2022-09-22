package models

// Failure - Response for a failed request
type Failure struct {
	Error FailureError `json:"error"`
}
