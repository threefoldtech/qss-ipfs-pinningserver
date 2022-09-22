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

// AddPin - Add pin object
func AddPin(c *gin.Context) {
	var pin models.Pin
	if err := c.ShouldBindJSON(&pin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	pinsRepo := database.New()
	cl, err := ipfsController.NewClusterController()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.FailureError{
			Reason:  "INTERNAL_SERVER_ERROR",
			Details: err.Error(),
		})
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

	err = pinsRepo.Insert(c, pinStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.FailureError{
			Reason:  "INTERNAL_SERVER_ERROR",
			Details: err.Error(),
		})
		return
	}

	err = cl.Add(c, pin)
	if err != nil {
		ce, ok := err.(*ipfsController.ControllerError)
		if ok {
			switch ce.Type {
			case ipfsController.INVALID_CID, ipfsController.INVALID_ORIGINS:
				c.JSON(http.StatusBadRequest, models.FailureError{
					Reason:  "BAD_REQUEST",
					Details: ce.Error(),
				})
				return
			default:
				c.JSON(http.StatusInternalServerError, models.FailureError{
					Reason:  "INTERNAL_SERVER_ERROR",
					Details: err.Error(),
				})
				return
			}
		}
	}
	isPinned, err := cl.IsPinned(c, pin.Cid)
	if !isPinned {
		go func(ctx *gin.Context) {
			// err := cl.WaitForPinned(ctx, pin.Cid)
			// if err != nil {
			// 	pinsRepo.Patch(ctx, request_id, map[string]interface{}{"status": db.FAILED})
			// 	return
			// }
			// pinsRepo.Patch(c, request_id, map[string]interface{}{"status": db.PINNED})
			for end := time.Now().Add(time.Minute * 10); ; {
				status, _ := cl.Status(c, pin.Cid)
				err = pinsRepo.Patch(ctx, request_id, map[string]interface{}{"status": db.Status(status)})
				if status == models.PINNED || status == models.FAILED {
					return
				}
				if time.Now().After(end) {
					break
				}
				time.Sleep(10 * time.Second)
			}

		}(c.Copy())
	} else {
		pinStatus.Status = models.PINNED
		pinsRepo.Patch(c, request_id, map[string]interface{}{"status": db.PINNED})
	}
	c.JSON(http.StatusAccepted, pinStatus)
}

// DeletePinByRequestId - Remove pin object
func DeletePinByRequestId(c *gin.Context) {
	// check if the cid shared between multiple requests
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.FailureError{
			Reason:  "BAD_REQUEST",
			Details: "Invalid `requestid` query parameter",
		})
		return
	}
	pinsRepo := database.New()
	pin_status, err := pinsRepo.FindByID(c, req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.FailureError{
			Reason:  "NOT_FOUND",
			Details: "The specified resource was not found",
		})
		return
	}
	cid := pin_status.Pin.Cid
	pin_results, err := pinsRepo.Find(c, []string{cid}, []string{}, "", time.Time{}, time.Time{}, "", 2)
	err = pinsRepo.Delete(c, req_id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.FailureError{
			Reason:  "INTERNAL_SERVER_ERROR",
			Details: err.Error(),
		})
		return
	}
	if pin_results.Count == 1 {
		cl, err := ipfsController.NewClusterController()
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.FailureError{
				Reason:  "INTERNAL_SERVER_ERROR",
				Details: err.Error(),
			})
			return
		}
		err = cl.Remove(c, cid)
	}

	c.JSON(http.StatusAccepted, gin.H{})
}

// GetPinByRequestId - Get pin object
func GetPinByRequestId(c *gin.Context) {
	req_id := c.Params.ByName("requestid")
	_, err := uuid.Parse(req_id)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.FailureError{
			Reason:  "BAD_REQUEST",
			Details: "Invalid `requestid` query parameter",
		})
		return
	}

	pinsRepo := database.New()
	pin_status, err := pinsRepo.FindByID(c, req_id)
	if err != nil {
		c.JSON(http.StatusNotFound, models.FailureError{
			Reason:  "NOT_FOUND",
			Details: "The specified resource was not found",
		})
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
	pinsRepo := database.New()
	pin_results, err := pinsRepo.Find(c, cids, statuses, name, time.Time{}, time.Time{}, match, limit_int)
	c.JSON(http.StatusOK, pin_results)
}

// ReplacePinByRequestId - Replace pin object
func ReplacePinByRequestId(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
