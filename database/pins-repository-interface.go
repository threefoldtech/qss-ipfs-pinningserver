package database

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

// Pins represents an interface to the Pins database
type PinsRepository interface {
	// Set adds or updates a Pin
	InsertOrGet(ctx *gin.Context, pinStatus models.PinStatus) (models.PinStatus, error)
	// Patch patches the fields of a Pin according to the given ID
	Patch(ctx *gin.Context, id string, fields map[string]interface{}) error
	// Get returns the Pin status for a given ID
	FindByID(ctx *gin.Context, id string) (models.PinStatus, error)
	// Find returns a list of Pins for the given parameters
	Find(ctx *gin.Context,
		cids, statuses []string,
		name string,
		before, after time.Time,
		match string,
		limit int,
	) (models.PinResults, error)
	// Delete removes the Pin according to the given ID
	Delete(ctx *gin.Context, id string) error
	CIDRefrenceCount(ctx *gin.Context, cid string) int64
}
