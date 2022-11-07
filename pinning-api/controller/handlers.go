package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/google/uuid"

	"github.com/threefoldtech/tf-pinning-service/config"
	db "github.com/threefoldtech/tf-pinning-service/database"
	ipfsController "github.com/threefoldtech/tf-pinning-service/ipfs-controller"
	"github.com/threefoldtech/tf-pinning-service/logger"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

func getUserIdFromContext(ctx *gin.Context) uint {
	user_id := ctx.GetUint("userID")
	return user_id
}

type Handlers struct {
	Log       *logrus.Logger
	PinsRepo  db.PinsRepository
	UsersRepo db.UsersRepository
	Config    config.Config
}

// AddPin - Add pin object
func (h *Handlers) AddPin(c *gin.Context) {
	log := h.Log

	var pin models.Pin

	if err := c.ShouldBindJSON(&pin); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, err.Error()))
		return
	}
	loggerContext := log.WithFields(logger.Fields{
		"topic":   "Handler-AddPin",
		"user_id": getUserIdFromContext(c),
		"cid":     pin.Cid,
	})
	pinsRepo := h.PinsRepo

	loggerContext.Debug("Trying to acquire the lock")
	pinsRepo.LockByCID(pin.Cid)
	loggerContext.Debug("Lock acquired")

	defer func() {
		pinsRepo.UnlockByCID(pin.Cid)
		loggerContext.Debug("Lock released")
	}()

	cl, err := ipfsController.GetClusterController(h.Config.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}

	request_id := uuid.New().String()
	// log delegates error
	pinStatus := models.PinStatus{
		Pin:       pin,
		Requestid: request_id,
		Status:    models.QUEUED,
		Created:   time.Now(),
	}
	pinStatus, err = pinsRepo.InsertOrGet(c, getUserIdFromContext(c), pinStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	delegates, err := cl.Delegates(c)
	if err != nil {
		loggerContext.WithFields(logger.Fields{
			"from_error": err.Error(),
		}).Error("Can't get delegates. set to empty!")
	}
	pinStatus.Delegates = delegates
	isNew := pinStatus.Requestid == request_id
	loggerContext = loggerContext.WithFields(logger.Fields{
		"request_id": pinStatus.Requestid,
	})
	err = cl.Add(c, pin) // use hooks to roll back in case request can't make it to the cluster?
	if err != nil {
		ce, ok := err.(*ipfsController.ControllerError)
		if ok {
			switch ce.Type {
			case ipfsController.INVALID_CID, ipfsController.INVALID_ORIGINS:
				c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, ce.Error()))
				return
			default:
				c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
				return
			}
		}
	}
	isPinned, err := cl.IsPinned(c, pin.Cid)
	if err != nil {
		loggerContext.WithFields(logger.Fields{
			"from_error": err.Error(),
		}).Error("Can't get pin status")
	} else if !isPinned && isNew {
		// The cluster-pinning stage is relatively fast,
		// but the ipfs-pinning stage can take much longer depending on the amount of things being pinned and the sizes of data
		go func(ctx *gin.Context) {
			err := cl.WaitForPinned(ctx, pin.Cid) // TODO: this will timeout if no progress could be made, but the cluster will retry later. need to sync db in that case.
			if err != nil {
				loggerContext.WithFields(logger.Fields{
					"from_error": err.Error(),
				}).Error("First attempt for pin failed, cluster will keep retry")
				// pinsRepo.Patch(ctx, getUserIdFromContext(ctx), pinStatus.Requestid, map[string]interface{}{"status": db.FAILED})
				return
			}
			pinsRepo.Patch(ctx, getUserIdFromContext(ctx), pinStatus.Requestid, map[string]interface{}{"status": db.PINNED})
			loggerContext.WithFields(logger.Fields{
				"new_status": "pinned",
			}).Debug("Status updated")
			// last_status := models.QUEUED
			// for end := time.Now().Add(time.Minute * 10); ; {
			// 	status, _ := cl.Status(c, pin.Cid)
			// 	if last_status != status {
			// 		err = pinsRepo.Patch(ctx, pinStatus.Requestid, map[string]interface{}{"status": db.Status(status)})
			// 		last_status = status
			// 		if status == models.PINNED || status == models.FAILED {
			// 			return
			// 		}
			// 	}
			// 	if time.Now().After(end) {
			// 		break
			// 	}
			// 	time.Sleep(1 * time.Second) // to fast ?
			// }

		}(c.Copy())
	}
	if isPinned {
		pinStatus.Status = models.PINNED
		pinsRepo.Patch(c, getUserIdFromContext(c), pinStatus.Requestid, map[string]interface{}{"status": db.PINNED})
	}
	loggerContext.WithFields(logger.Fields{
		"status": pinStatus.Status,
	}).Debug("Request Received and handled")
	c.JSON(http.StatusAccepted, pinStatus)
}

// DeletePinByRequestId - Remove pin object
func (h *Handlers) DeletePinByRequestId(c *gin.Context) {
	log := h.Log
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `requestid` query parameter"))
		return
	}
	pinsRepo := h.PinsRepo
	pin_status, err := pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}
	cid := pin_status.Pin.Cid
	loggerContext := log.WithFields(logger.Fields{
		"topic":      "Handler-DeletePin",
		"user_id":    getUserIdFromContext(c),
		"request_id": pin_status.Requestid,
		"cid":        cid,
	})
	loggerContext.Debug("Trying to acquire the lock")
	pinsRepo.LockByCID(cid)
	loggerContext.Debug("Lock acquired")

	defer func() {
		pinsRepo.UnlockByCID(cid)
		loggerContext.Debug("Lock released")
	}()
	// workaround
	pin_status, err = pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}
	count, err := pinsRepo.CIDRefrenceCount(c, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}

	err = pinsRepo.Delete(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}

	if count == 1 {
		loggerContext.Debug("This cid no longer referenced by service records, and will be unpinned.")
		cl, err := ipfsController.GetClusterController(h.Config.Cluster)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
			return
		}
		err = cl.Remove(c, cid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
			return
		}
	}

	c.JSON(http.StatusAccepted, gin.H{})
}

// GetPinByRequestId - Get pin object
func (h *Handlers) GetPinByRequestId(c *gin.Context) {
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `requestid` query parameter"))
		return
	}

	pinsRepo := h.PinsRepo
	pin_status, err := pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}
	c.JSON(http.StatusOK, pin_status)
}

// GetPins - List pin objects
func (h *Handlers) GetPins(c *gin.Context) {
	limit := c.Query("limit")
	status := c.Query("status")
	name := c.Query("name")
	cid := c.Query("cid")
	before := c.Query("before")
	after := c.Query("after")
	match := c.Query("match")
	var cids, statuses []string
	if status != "" {
		statuses = strings.Split(status, ",")
	}
	if cid != "" {
		cids = strings.Split(cid, ",")
	}

	limit_int, err := strconv.Atoi(limit)
	if err != nil || limit_int < 1 || limit_int > 1000 {
		limit_int = 10
	}
	var before_t, after_t time.Time
	if before != "" {
		before_t, err = time.Parse(time.RFC3339, before)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `before` query parameter"))
			return
		}
	}
	if after != "" {
		after_t, err = time.Parse(time.RFC3339, after)
		if err != nil {
			c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `after` query parameter"))
			return
		}
	}
	pinsRepo := h.PinsRepo
	pin_results, err := pinsRepo.Find(c, getUserIdFromContext(c), cids, statuses, name, before_t, after_t, match, limit_int)
	c.JSON(http.StatusOK, pin_results)
}

// ReplacePinByRequestId - Replace pin object
func (h *Handlers) ReplacePinByRequestId(c *gin.Context) {
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `requestid` query parameter"))
		return
	}
	log := h.Log
	var pin models.Pin

	if err := c.ShouldBindJSON(&pin); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, err.Error()))
		return
	}
	loggerContext := log.WithFields(logger.Fields{
		"topic":          "Handler-ReplacePin",
		"user_id":        getUserIdFromContext(c),
		"old_request_id": req_id,
		"cid":            pin.Cid,
	})
	pinsRepo := h.PinsRepo
	pin_status, err := pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}

	cid := pin_status.Pin.Cid
	if cid == pin.Cid {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "The cid is the same as the one you want to replace!"))
		return
	}
	loggerContext.Debug("Trying to acquire the lock for pin operation")
	pinsRepo.LockByCID(pin.Cid)
	loggerContext.Debug("Lock acquired")

	cl, err := ipfsController.GetClusterController(h.Config.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		pinsRepo.UnlockByCID(pin.Cid)
		return
	}

	request_id := uuid.New().String()
	pinStatus := models.PinStatus{
		Pin:       pin,
		Requestid: request_id,
		Status:    models.QUEUED,
		Created:   time.Now(),
	}
	pinStatus, err = pinsRepo.InsertOrGet(c, getUserIdFromContext(c), pinStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		pinsRepo.UnlockByCID(pin.Cid)
		return
	}
	delegates, err := cl.Delegates(c)
	if err != nil {
		loggerContext.WithFields(logger.Fields{
			"from_error": err.Error(),
		}).Error("Can't get delegates. set to empty!")
	}
	pinStatus.Delegates = delegates
	isNew := pinStatus.Requestid == request_id
	loggerContext = loggerContext.WithFields(logger.Fields{
		"request_id": pinStatus.Requestid,
	})
	err = cl.Add(c, pin)
	if err != nil {
		ce, ok := err.(*ipfsController.ControllerError)
		if ok {
			defer pinsRepo.UnlockByCID(pin.Cid)
			switch ce.Type {
			case ipfsController.INVALID_CID, ipfsController.INVALID_ORIGINS:
				c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, ce.Error()))
				return
			default:
				c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
				return
			}
		}
	}
	isPinned, err := cl.IsPinned(c, pin.Cid)
	if err != nil {
		loggerContext.WithFields(logger.Fields{
			"from_error": err.Error(),
		}).Error("Can't get pin status")
	} else if !isPinned && isNew {
		go func(ctx *gin.Context) {
			err := cl.WaitForPinned(ctx, pin.Cid) // TODO: this will timeout if no progress could be made, but the cluster will retry later. need to sync db in that case.
			if err != nil {
				loggerContext.WithFields(logger.Fields{
					"from_error": err.Error(),
				}).Error("First attempt for pin failed, cluster will keep retry")
				return
			}
			pinsRepo.Patch(ctx, getUserIdFromContext(ctx), pinStatus.Requestid, map[string]interface{}{"status": db.PINNED})

		}(c.Copy())
	}
	if isPinned {
		pinStatus.Status = models.PINNED
		pinsRepo.Patch(c, getUserIdFromContext(c), pinStatus.Requestid, map[string]interface{}{"status": db.PINNED})
	}
	pinsRepo.UnlockByCID(pin.Cid)
	loggerContext.Debug("Lock released")

	loggerContext.Debug("Trying to acquire the lock for unpin operation")
	pinsRepo.LockByCID(cid)
	defer func() {
		pinsRepo.UnlockByCID(cid)
		loggerContext.Debug("Lock released")
	}()
	// workaround
	pin_status, err = pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}
	// delete
	count, err := pinsRepo.CIDRefrenceCount(c, cid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}

	err = pinsRepo.Delete(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}

	if count == 1 {
		loggerContext.Debug("This cid no longer referenced by service records, and will be unpinned.")
		cl, err := ipfsController.GetClusterController(h.Config.Cluster)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
			return
		}
		err = cl.Remove(c, cid)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
			return
		}
	}
	c.JSON(http.StatusAccepted, gin.H{})
}

func (h *Handlers) GetPeers(c *gin.Context) {
	cl, err := ipfsController.GetClusterController(h.Config.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	peersInfo, err := cl.Peers(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, peersInfo)

}

func (h *Handlers) GetAlerts(c *gin.Context) {
	cl, err := ipfsController.GetClusterController(h.Config.Cluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	alerts, err := cl.Alerts(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, alerts)
}
