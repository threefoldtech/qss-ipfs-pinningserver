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
	"github.com/threefoldtech/tf-pinning-service/pinning-api/models"
)

func getUserIdFromContext(ctx *gin.Context) uint {
	user_id := ctx.GetUint("userID")
	return user_id
}

// AddPin - Add pin object
func AddPin(c *gin.Context) {
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
	if !isPinned {
		go func(ctx *gin.Context) {
			err := cl.WaitForPinned(ctx, pin.Cid) // TODO: this will timeout if no progress could be made, but the cluster will retry later. need to sync db in that case.
			if err != nil {
				pinsRepo.Patch(ctx, getUserIdFromContext(c), request_id, map[string]interface{}{"status": db.FAILED})
				return
			}
			pinsRepo.Patch(c, getUserIdFromContext(c), request_id, map[string]interface{}{"status": db.PINNED})
			// last_status := models.QUEUED
			// for end := time.Now().Add(time.Minute * 10); ; {
			// 	status, _ := cl.Status(c, pin.Cid)
			// 	if last_status != status {
			// 		err = pinsRepo.Patch(ctx, request_id, map[string]interface{}{"status": db.Status(status)})
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
	} else {
		pinStatus.Status = models.PINNED
		pinsRepo.Patch(c, getUserIdFromContext(c), request_id, map[string]interface{}{"status": db.PINNED})
	}
	// fmt.Println(cl.DagSize(c, pin.Cid))
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
	// pin_results, err := pinsRepo.Find(c, []string{cid}, []string{}, "", time.Time{}, time.Time{}, "", 2)

	if pinsRepo.CIDRefrenceCount(c, cid) == 1 { // TODO: handle possible race condition
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
	//before := c.Param("before")
	//after := c.Param("after")
	match := c.Param("match")
	var cids, statuses []string
	if status != "" {
		statuses = strings.Split(status, ",")
	}
	if cid != "" {
		cids = strings.Split(cid, ",")
	}
	limit_int, err := strconv.Atoi(limit)
	if err != nil {
		limit_int = 10
	}
	pinsRepo := database.NewPinsRepository()
	pin_results, err := pinsRepo.Find(c, getUserIdFromContext(c), cids, statuses, name, time.Time{}, time.Time{}, match, limit_int)
	c.JSON(http.StatusOK, pin_results)
}

// ReplacePinByRequestId - Replace pin object
func ReplacePinByRequestId(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.NewAPIError(http.StatusNotImplemented, "This functionality not implemented yet"))
}
