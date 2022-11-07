package ipfsController

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"github.com/ipfs-cluster/ipfs-cluster/api"
	"github.com/ipfs-cluster/ipfs-cluster/api/rest/client"
	shell "github.com/ipfs/go-ipfs-api"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

const DefaultPinMode = api.PinModeRecursive

type clusterController struct {
	Client client.Client
	// Attributes
	ReplicationFactorMin int
	ReplicationFactorMax int
	PinMode              api.PinMode
}

func GetClusterController(cfg config.ClusterConfig) (IpfsController, error) {
	// ProxyAddr, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/5001")
	client, err := client.NewDefaultClient(&client.Config{
		Host:     cfg.Host,
		Port:     cfg.Port,
		Username: cfg.Username,
		Password: cfg.Password,
		Timeout:  time.Duration(cfg.IpfsClusterTimeout) * time.Second,
	})

	if err != nil {
		return nil, &ControllerError{
			Type: CONNECTION_ERROR,
			Err:  err,
		}
	}
	c := &clusterController{
		Client:               client,
		ReplicationFactorMin: cfg.ReplicationFactorMin,
		ReplicationFactorMax: cfg.ReplicationFactorMax,
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
	for i := 0; i < 3; i++ {
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
		if err == nil {
			break
		}
	}
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
	for i := 0; i < 3; i++ {
		_, err = c.Client.Unpin(ctx, id)
		if err == nil {
			break
		}
	}
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
	for _, addr := range clientId.IPFS.Addresses {
		v, err := ParseIPFromMultiaddr(addr.Multiaddr)
		if err == nil && !v.IsLoopback() {
			delegates = append(delegates, addr.String())
		}
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
	if err != nil {
		return models.FAILED, &ControllerError{
			Type: CONNECTION_ERROR,
			Err:  err,
		}
	}
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

func (c *clusterController) StatusCids(ctx context.Context, cids []string) (map[string]models.Status, error) {
	res := make(map[string]models.Status)
	var cids_decoded []api.Cid
	for _, cid := range cids {
		cid_decoded, err := api.DecodeCid(cid)
		if err != nil {
			return res, &ControllerError{
				Type: INVALID_CID,
				Err:  err,
			}
		}
		cids_decoded = append(cids_decoded, cid_decoded)
	}
	pinsInfo := make(chan api.GlobalPinInfo)
	go c.Client.StatusCids(ctx, cids_decoded, false, pinsInfo)

	done := make(chan struct{})
	go func() {
		for pinInfo := range pinsInfo {
			if pinInfo.Match(api.TrackerStatusPinning) {
				res[pinInfo.Cid.String()] = models.PINNING
			} else if pinInfo.Match(api.TrackerStatusQueued) {
				res[pinInfo.Cid.String()] = models.QUEUED
			} else if pinInfo.Match(api.TrackerStatusPinned) {
				res[pinInfo.Cid.String()] = models.PINNED
			} else {
				res[pinInfo.Cid.String()] = models.FAILED
			}
		}
		done <- struct{}{}
	}()
	<-done
	return res, nil
}

func (c *clusterController) IsPinned(ctx context.Context, cid string) (bool, error) {
	status, err := c.Status(ctx, cid)
	if err != nil {
		return false, err
	}
	return status == models.PINNED, nil
}

// cluster peer not aware or care about the dag size (we need it for the billing, right?), for that piece of information we need reach to the ipfs peer
// reaching to ipfs peer directly would require ipfs proxy port to be exposed to work properly
// usually/by default it running only on localhost, problem is that it should not be exposed without an authentication mechanism on top (nginx etc…).
// and by default it provides no authentication nor encryption
// maybe we could query a public ipfs gateway to fetch this information instead ?
func (c *clusterController) DagSize(ctx context.Context, cid string) (*shell.ObjectStats, error) {
	r, err := c.Client.IPFS(ctx).ObjectStat(cid)
	if err != nil {
		return nil, &ControllerError{
			Type: CONNECTION_ERROR,
			Err:  err,
		}
	}
	return r, nil
}

func (c *clusterController) Alerts(ctx context.Context) ([]api.Alert, error) {
	alerts, err := c.Client.Alerts(ctx)
	if err != nil {
		return nil, &ControllerError{
			Type: CONNECTION_ERROR,
			Err:  err,
		}
	}
	return alerts, nil
}

func (c *clusterController) Peers(ctx context.Context) ([]api.ID, error) {
	messages := make(chan api.ID, 1)
	err := c.Client.Peers(ctx, messages)
	if err != nil {
		return nil, &ControllerError{
			Type: CONNECTION_ERROR,
			Err:  err,
		}
	}
	var peers []api.ID
	done := make(chan struct{})
	go func() {
		for elem := range messages {
			peers = append(peers, elem)
		}
		done <- struct{}{}
	}()

	<-done
	return peers, nil
}

func ParseIPFromMultiaddr(addr ma.Multiaddr) (net.IP, error) {
	ErrInvalidMultiaddrFormat := errors.New("invalid multiaddr format")
	s := addr.String()
	parts := strings.Split(s, "/")
	if parts[0] != "" {
		return nil, ErrInvalidMultiaddrFormat
	}
	if len(parts) < 3 {
		return nil, ErrInvalidMultiaddrFormat
	}
	isip := parts[1] == "ip4" || parts[1] == "ip6"
	if !isip {
		return nil, ErrInvalidMultiaddrFormat
	}
	return net.ParseIP(parts[2]), nil
}
