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

func SetDagService(interval int, log *logrus.Logger, pinsRepo database.PinsRepository, ipfsClusterConfig config.ClusterConfig) {
	ctx := context.Background()
	s := GetScheduler()
	s.Every(interval).Minutes().Do(func() {
		loggerContext := log.WithFields(logger.Fields{
			"topic": "Service-Dag",
		})
		loggerContext.Info("Dag service started")
		strated_time := time.Now()
		cl, err := ipfsController.GetClusterController(ipfsClusterConfig)
		if err != nil {
			loggerContext.WithFields(logger.Fields{
				"from_error": err.Error(),
			}).Error("Can't get cluster controller")
			return
		}
		pins, _ := pinsRepo.ProcessByStatus(ctx, []string{"pinned"})
		for patch := range pins {
			for _, pin := range patch {
				innerContext := loggerContext.WithFields(logger.Fields{
					"cid":        pin.Cid,
					"status":     pin.Status,
					"user_id":    pin.UserID,
					"request_id": pin.UUID})
				stat, err := cl.DagSize(ctx, pin.Cid)
				if err != nil {
					innerContext.WithFields(logger.Fields{
						"from_error": err.Error(),
					}).Warn("Can't get the dag size from the cluster peer!")
					continue
				}
				if stat.CumulativeSize != 0 {
					err = pinsRepo.Patch(ctx, pin.UserID, pin.UUID, map[string]interface{}{"dag_size": stat.CumulativeSize})
					if err != nil {
						innerContext.WithFields(logger.Fields{
							"from_error": err.Error(),
						}).Warn("Can't update the dag size on the db!")
						continue
					}
					innerContext.WithFields(logger.Fields{
						"dag size": pin.DagSize,
					}).Info("dag size updated")
				}
			}
		}
		loggerContext.Info("Dag service finished. took ", time.Since(strated_time))
	})
}
