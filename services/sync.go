package services

import (
	"context"
	"tf-ipfs-pinning-service/logger"
	"time"

	"github.com/threefoldtech/tf-pinning-service/database"

	ipfsController "github.com/threefoldtech/tf-pinning-service/ipfs-controller"
)

const interval = 10 // for testing

func SetSyncService() {
	ctx := context.Background()
	s := GetScheduler()
	log := logger.GetDefaultLogger()
	s.Every(interval).Minutes().Do(func() {
		loggerContext := log.WithFields(logger.Fields{
			"topic": "Services-Sync",
		})
		loggerContext.Info("Sync service started")
		pinsRepo := database.NewPinsRepository()
		cl, err := ipfsController.NewClusterController()
		if err != nil {
			loggerContext.WithFields(logger.Fields{
				"from_error": err.Error(),
			}).Error("Can't get cluster controller")
		}
		pins, _ := pinsRepo.FindByStatus(ctx, []string{"failed", "queued"}) // TODO: Use rows iteration for optimal memory usage

		for _, pin := range pins {
			if pinned, _ := cl.IsPinned(ctx, pin.Cid); pinned {
				pinsRepo.Patch(ctx, pin.UserID, pin.UUID, map[string]interface{}{"status": "pinned"})
				loggerContext.WithFields(logger.Fields{
					"cid":        pin.Cid,
					"status":     pin.Status,
					"new status": pinned,
				}).Info("Status updated")
			} else {
				elapsed := time.Now().Sub(pin.CreatedAt)
				if elapsed.Hours() > 24*7 {
					loggerContext.WithFields(logger.Fields{
						"cid":        pin.Cid,
						"status":     pin.Status,
						"new status": "",
					}).Info("CID stuck for a week+")
				}
			}
		}
	})
}