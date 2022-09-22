package ipfsController

type ErrorType string

// List of Status
const (
	INVALID_CID      ErrorType = "invalid_cid"
	INVALID_ORIGINS  ErrorType = "invalid_origins"
	CONNECTION_ERROR ErrorType = "connection_error"
	PIN_ERROR        ErrorType = "pin_error"
	UNPIN_ERROR      ErrorType = "unpin_error"
)

type ControllerError struct {
	Type ErrorType
	Err  error
}

func (r *ControllerError) Error() string {
	return r.Err.Error()
}
