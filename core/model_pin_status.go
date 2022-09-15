package tfpin

import (
	"time"
)

// PinStatus - Pin object with status
type PinStatus struct {

	// Globally unique identifier of the pin request; can be used to check the status of ongoing pinning, or pin removal
	Requestid string `json:"requestid"`

	Status Status `json:"status"`

	// Immutable timestamp indicating when a pin request entered a pinning service; can be used for filtering results and pagination
	Created time.Time `json:"created"`

	Pin Pin `json:"pin"`

	Delegates Delegates `json:"delegates"`

	Info StatusInfo `json:"info,omitempty"`
}
