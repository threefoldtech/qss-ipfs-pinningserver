package services

import (
	"context"
	"fmt"
	"time"

	"github.com/threefoldtech/tf-pinning-service/database"

	ipfsController "github.com/threefoldtech/tf-pinning-service/ipfs-controller"
)

const interval = 10 // for testing

func SetSyncService() {
	ctx := context.Background()
	s := GetScheduler()

	s.Every(interval).Minutes().Do(func() {
		fmt.Printf("\nTime: %v Sync service started", time.Now())
		pinsRepo := database.NewPinsRepository()
		cl, _ := ipfsController.NewClusterController()
		pins, _ := pinsRepo.FindByStatus(ctx, []string{"failed", "queued"})

		for _, pin := range pins {
			if pinned, _ := cl.IsPinned(ctx, pin.Cid); pinned {
				pinsRepo.Patch(ctx, pin.UserID, pin.UUID, map[string]interface{}{"status": "pinned"})
				fmt.Printf("\ncid: %v\nstatus: %v\nnew status: %v", pin.Cid, pin.Status, pinned)
			} else {
				elapsed := time.Now().Sub(pin.CreatedAt)
				if elapsed.Hours() > 24*7 {
					fmt.Printf("\ncid: %v still not pinned after more than week!", pin.Cid)
				}
			}
		}
	})
}
