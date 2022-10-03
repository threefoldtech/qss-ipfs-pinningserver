package models

import "net/http"

// Failure - Response for a failed request
type Failure struct {
	Error FailureError `json:"error"`
}

func NewAPIError(Status int, Details string) Failure {
	return Failure{
		FailureError{
			Reason:  http.StatusText(Status),
			Details: Details,
		},
	}
}
