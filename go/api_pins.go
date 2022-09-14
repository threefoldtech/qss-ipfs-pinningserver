package tfpin

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AddPin - Add pin object
func AddPin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// DeletePinByRequestId - Remove pin object
func DeletePinByRequestId(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// GetPinByRequestId - Get pin object
func GetPinByRequestId(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// GetPins - List pin objects
func GetPins(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}

// ReplacePinByRequestId - Replace pin object
func ReplacePinByRequestId(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{})
}
