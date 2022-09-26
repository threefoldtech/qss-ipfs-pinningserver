package database

import "context"

type UsersRepository interface {
	// Set adds or updates a Pin
	Insert(ctx context.Context, token string) error
	// Patch patches the fields of a Pin according to the given ID
	Patch(ctx context.Context, id string, fields map[string]interface{}) error
	// Get returns the Pin status for a given ID
	FindByID(ctx context.Context, id string) (User, error)
	// Find returns a list of Pins for the given parameters
	FindByTokenHash(ctx context.Context, hash string) (User, error)
	// Delete removes the Pin according to the given ID
	Delete(ctx context.Context, id string) error
}
