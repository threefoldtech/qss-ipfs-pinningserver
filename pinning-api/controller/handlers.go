package controller

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/google/uuid"

	"github.com/threefoldtech/tf-pinning-service/database"
	db "github.com/threefoldtech/tf-pinning-service/database"
	ipfsController "github.com/threefoldtech/tf-pinning-service/ipfs-controller"
	"github.com/threefoldtech/tf-pinning-service/logger"
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

func getUserIdFromContext(ctx *gin.Context) uint {
	user_id := ctx.GetUint("userID")
	return user_id
}

// AddPin - Add pin object
func AddPin(c *gin.Context) {
	log := logger.GetDefaultLogger()

	var pin models.Pin
	// c.Get("userID")

	if err := c.ShouldBindJSON(&pin); err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, err.Error()))
		return
	}

	pinsRepo := database.NewPinsRepository()
	pinsRepo.Lock()
	defer pinsRepo.Unlock()
	cl, err := ipfsController.NewClusterController()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}

	request_id := uuid.New().String()
	delegates, err := cl.Delegates(c)
	// log delegates error
	pinStatus := models.PinStatus{
		Pin:       pin,
		Requestid: request_id,
		Status:    models.QUEUED,
		Created:   time.Now(),
		Delegates: delegates,
	}
	pinStatus, err = pinsRepo.InsertOrGet(c, getUserIdFromContext(c), pinStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	isNew := pinStatus.Requestid == request_id
	loggerContext := log.WithFields(logger.Fields{
		"topic":      "Handler-AddPin",
		"user_id":    getUserIdFromContext(c),
		"request_id": pinStatus.Requestid,
		"cid":        pin.Cid,
	})
	err = cl.Add(c, pin)
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
				pinsRepo.Patch(ctx, getUserIdFromContext(c), pinStatus.Requestid, map[string]interface{}{"status": db.FAILED})
				return
			}
			pinsRepo.Patch(c, getUserIdFromContext(c), pinStatus.Requestid, map[string]interface{}{"status": db.PINNED})
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
	}).Info("Request Received and handled")
	c.JSON(http.StatusAccepted, pinStatus)
}

// DeletePinByRequestId - Remove pin object
func DeletePinByRequestId(c *gin.Context) {
	// check if the cid shared between multiple requests
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `requestid` query parameter"))
		return
	}
	pinsRepo := database.NewPinsRepository()
	pinsRepo.Lock()
	defer pinsRepo.Unlock()
	pin_status, err := pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}
	cid := pin_status.Pin.Cid

	if pinsRepo.CIDRefrenceCount(c, cid) == 1 {
		cl, err := ipfsController.NewClusterController()
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
	err = pinsRepo.Delete(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.NewAPIError(http.StatusInternalServerError, err.Error()))
		return
	}
	c.JSON(http.StatusAccepted, gin.H{})
}

// GetPinByRequestId - Get pin object
func GetPinByRequestId(c *gin.Context) {
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.NewAPIError(http.StatusBadRequest, "Invalid `requestid` query parameter"))
		return
	}

	pinsRepo := database.NewPinsRepository()
	pin_status, err := pinsRepo.FindByID(c, getUserIdFromContext(c), req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.NewAPIError(http.StatusNotFound, "The specified resource was not found"))
		return
	}
	c.JSON(http.StatusOK, pin_status)
}

// GetPins - List pin objects
func GetPins(c *gin.Context) {
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
	pinsRepo := database.NewPinsRepository()
	pin_results, err := pinsRepo.Find(c, getUserIdFromContext(c), cids, statuses, name, before_t, after_t, match, limit_int)
	c.JSON(http.StatusOK, pin_results)
}

// ReplacePinByRequestId - Replace pin object
func ReplacePinByRequestId(c *gin.Context) {
	// TODO
	c.JSON(http.StatusNotImplemented, models.NewAPIError(http.StatusNotImplemented, "This functionality not implemented yet"))
}
