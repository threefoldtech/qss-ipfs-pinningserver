package ipfsController

import (
	"context"

	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

type IpfsController interface {
	Add(ctx context.Context, pin models.Pin) error
	Remove(ctx context.Context, cid string) error
	Delegates(ctx context.Context) ([]string, error)
	SetReplicationFactor(min, max int)
	SetPinMode(mode api.PinMode)
	WaitForPinned(ctx context.Context, cid string) error
	IsPinned(ctx context.Context, cid string) (bool, error)
	Status(ctx context.Context, cid string) (models.Status, error)
	// DagSize(ctx context.Context, key string) (*shell.ObjectStats, error)
	StatusCids(ctx context.Context, cids []string) (map[string]models.Status, error)
}
