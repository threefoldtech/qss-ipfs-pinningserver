package ipfsController

import (
	"context"
	"fmt"
	"time"

	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

const (
	DefaultReplicationFactorMin = 1
	DefaultReplicationFactorMax = 3
	DefaultPinMode              = api.PinModeRecursive
)

type clusterController struct {
	Client client.Client
	// Attributes
	ReplicationFactorMin int
	ReplicationFactorMax int
	PinMode              api.PinMode
}

func NewClusterController() (ipfsController, error) {
	// ProxyAddr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5001")
	client, err := client.NewDefaultClient(&client.Config{
		Host: config.CFG.Cluster.Host,
		Port: config.CFG.Cluster.Port,
	})

	if err != nil {
		return nil, &ControllerError{
			Type: CONNECTION_ERROR,
			Err:  err,
		}
	}
	c := &clusterController{
		Client:               client,
		ReplicationFactorMin: DefaultReplicationFactorMin,
		ReplicationFactorMax: DefaultReplicationFactorMax,
		PinMode:              DefaultPinMode,
	}
	return c, nil
}

func (c *clusterController) Add(ctx context.Context, pin models.Pin) error {
	id, err := api.DecodeCid(pin.Cid)
	if err != nil {
		return &ControllerError{
			Type: INVALID_CID,
			Err:  err,
		}
	}
	origins := []api.Multiaddr{}
	for _, s := range pin.Origins {
		m, err := api.NewMultiaddr(s)
		if err != nil {
			return &ControllerError{
				Type: INVALID_ORIGINS,
				Err:  err,
			}
		}
		origins = append(origins, m)
	}
	_, err = c.Client.Pin(ctx, id,
		api.PinOptions{
			ReplicationFactorMin: c.ReplicationFactorMin,
			ReplicationFactorMax: c.ReplicationFactorMax,
			Name:                 pin.Name,
			Mode:                 c.PinMode,
			ShardSize:            api.DefaultShardSize,
			Metadata:             pin.Meta,
			Origins:              origins,
		},
	)
	if err != nil {
		return &ControllerError{
			Type: PIN_ERROR,
			Err:  err,
		}
	}

	return nil
}

func (c *clusterController) Remove(ctx context.Context, cid string) error {
	id, err := api.DecodeCid(cid)
	if err != nil {
		return &ControllerError{
			Type: INVALID_CID,
			Err:  err,
		}
	}
	_, err = c.Client.Unpin(ctx, id)
	if err != nil {
		return &ControllerError{
			Type: UNPIN_ERROR,
			Err:  err,
		}
	}
	return nil
}
func (c *clusterController) Delegates(ctx context.Context) ([]string, error) {
	clientId, err := c.Client.ID(ctx)
	if err != nil {
		return []string{}, err
	}
	delegates := []string{}
	// TODO: All multiaddrs MUST end with `/p2p/{peerID}` and SHOULD be fully resolved and confirmed to be disable from the public internet. Avoid sending addresses from local networks.
	for _, addr := range clientId.IPFS.Addresses {
		delegates = append(delegates, addr.String())

	}
	return delegates, nil
}

func (c *clusterController) SetReplicationFactor(min, max int) {

	c.ReplicationFactorMin = min
	c.ReplicationFactorMax = max
}

func (c *clusterController) SetPinMode(mode api.PinMode) {
	c.PinMode = mode
}

func (c *clusterController) WaitForPinned(ctx context.Context, cid string) error {
	cid_decoded, err := api.DecodeCid(cid)
	if err != nil {
		return &ControllerError{
			Type: INVALID_CID,
			Err:  err,
		}
	}
	_, err = client.WaitFor(ctx, c.Client, client.StatusFilterParams{
		Cid:       cid_decoded,
		Local:     false,
		Target:    api.TrackerStatusPinned,
		CheckFreq: 10 * time.Second,
	})
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (c *clusterController) Status(ctx context.Context, cid string) (models.Status, error) {
	cid_decoded, err := api.DecodeCid(cid)
	if err != nil {
		return models.FAILED, &ControllerError{
			Type: INVALID_CID,
			Err:  err,
		}
	}
	pinInfo, err := c.Client.Status(ctx, cid_decoded, false)
	if pinInfo.Match(api.TrackerStatusPinning) {
		return models.PINNING, nil
	} else if pinInfo.Match(api.TrackerStatusQueued) {
		return models.QUEUED, nil
	} else if pinInfo.Match(api.TrackerStatusPinned) {
		return models.PINNED, nil
	} else {
		return models.FAILED, nil
	}
}

func (c *clusterController) IsPinned(ctx context.Context, cid string) (bool, error) {
	status, err := c.Status(ctx, cid)
	if err != nil {
		return false, err
	}
	return status == models.PINNED, nil
}

func (c *clusterController) DagSize(ctx context.Context, cid string) (*shell.ObjectStats, error) {
	r, err := c.Client.IPFS(ctx).ObjectStat(cid)
	if err != nil {
		return nil, err
	}
	return r, nil
}
