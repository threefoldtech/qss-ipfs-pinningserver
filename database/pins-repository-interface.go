package database

import (
	"context"
	"time"

	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

// Pins represents an interface to the Pins database
type PinsRepository interface {
	LockByCID(string)
	UnlockByCID(string)
	// Set adds or updates a Pin
	InsertOrGet(ctx context.Context, userID uint, pinStatus models.PinStatus) (models.PinStatus, error)
	// Patch patches the fields of a Pin according to the given ID
	Patch(ctx context.Context, userID uint, id string, fields map[string]interface{}) error
	// Get returns the Pin status for a given ID
	FindByID(ctx context.Context, userID uint, id string) (models.PinStatus, error)
	// Find returns a list of Pins for the given parameters
	Find(ctx context.Context,
		userID uint,
		cids, statuses []string,
		name string,
		before, after time.Time,
		match string,
		limit int,
	) (models.PinResults, error)
	// Delete removes the Pin according to the given ID
	Delete(ctx context.Context, userID uint, id string) error
	CIDRefrenceCount(ctx context.Context, cid string) (int64, error)
	FindByStatus(ctx context.Context, statuses []string) ([]PinDTO, error)
	ProcessByStatus(ctx context.Context, statuses []string, c chan bool) (chan *PinDTO, error)
	/*
		 	Lock()
			Unlock()
			Begin() *pins
			Rollback()
			Commit()
	*/
}
