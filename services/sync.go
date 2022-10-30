package services

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/threefoldtech/tf-pinning-service/config"
	"github.com/threefoldtech/tf-pinning-service/database"
	"github.com/threefoldtech/tf-pinning-service/logger"

	ipfsController "github.com/threefoldtech/tf-pinning-service/ipfs-controller"
)

func SetSyncService(interval int, log *logrus.Logger, pinsRepo database.PinsRepository, ipfsClusterConfig config.ClusterConfig) {
	ctx := context.Background()
	s := GetScheduler()
	s.Every(interval).Minutes().Do(func() {
		loggerContext := log.WithFields(logger.Fields{
			"topic": "Service-Sync",
		})
		loggerContext.Info("Sync service started")
		strated_time := time.Now()
		cl, err := ipfsController.GetClusterController(ipfsClusterConfig)
		if err != nil {
			loggerContext.WithFields(logger.Fields{
				"from_error": err.Error(),
			}).Error("Can't get cluster controller")
			return
		}
		done := make(chan bool, 1)
		pins, _ := pinsRepo.ProcessByStatus(ctx, []string{"failed", "queued"}, done) // TODO: Use rows iteration for optimal memory usage
		for pin := range pins {
			innerContext := loggerContext.WithFields(logger.Fields{
				"cid":        pin.Cid,
				"status":     pin.Status,
				"user_id":    pin.UserID,
				"request_id": pin.UUID})
			pinned, err := cl.IsPinned(ctx, pin.Cid)
			if err != nil {
				innerContext.WithFields(logger.Fields{
					"from_error": err.Error(),
				}).Warn("Can't get the pin status from the cluster peer!")
				done <- pinned
				continue
			}
			if pinned {
				pin.Status = database.PINNED
				innerContext.WithFields(logger.Fields{
					"new status": "pinned",
				}).Info("Status updated")
			} else {
				elapsed := time.Since(time.Unix(int64(pin.CreatedAt), 0))
				if elapsed.Hours() > 24*7 {
					innerContext.WithFields(logger.Fields{
						"new status": "",
					}).Warn("CID stuck for a week+")
					// too many retry attempts can generate a lot of extra requests and extra load on the system.
					// Should we delete the request on behalf of the user?
				}
			}
			done <- pinned
		}
		loggerContext.Info("Sync service finished. took ", time.Since(strated_time))
	})
}
