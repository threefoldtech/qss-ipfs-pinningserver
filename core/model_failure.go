package tfpin

// Failure - Response for a failed request
type Failure struct {
	Error FailureError `json:"error"`
}
