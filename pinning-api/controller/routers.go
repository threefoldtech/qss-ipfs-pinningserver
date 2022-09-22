package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Route is the information for every URI.
type Route struct {
	// Name is the name of this Route.
	Name string
	// Method is the string for the HTTP method. ex) GET, POST etc..
	Method string
	// Pattern is the pattern of the URI.
	Pattern string
	// HandlerFunc is the handler function of this route.
	HandlerFunc gin.HandlerFunc
}

// Routes is the list of the generated Route.
type Routes []Route

// NewRouter returns a new router.
func NewRouter() *gin.Engine {
	router := gin.Default()
	v1 := router.Group("/api/v1")

	for _, route := range routes {
		switch route.Method {
		case http.MethodGet:
			v1.GET(route.Pattern, route.HandlerFunc)
		case http.MethodPost:
			v1.POST(route.Pattern, route.HandlerFunc)
		case http.MethodPut:
			v1.PUT(route.Pattern, route.HandlerFunc)
		case http.MethodPatch:
			v1.PATCH(route.Pattern, route.HandlerFunc)
		case http.MethodDelete:
			v1.DELETE(route.Pattern, route.HandlerFunc)
		}
	}

	return router
}

// Index is the index handler.
func Index(c *gin.Context) {
	c.String(http.StatusOK, "Hello World!")
}

var routes = Routes{
	{
		"Index",
		http.MethodGet,
		"/",
		Index,
	},

	{
		"AddPin",
		http.MethodPost,
		"/pins",
		AddPin,
	},

	{
		"DeletePinByRequestId",
		http.MethodDelete,
		"/pins/:requestid",
		DeletePinByRequestId,
	},

	{
		"GetPinByRequestId",
		http.MethodGet,
		"/pins/:requestid",
		GetPinByRequestId,
	},

	{
		"GetPins",
		http.MethodGet,
		"/pins",
		GetPins,
	},

	{
		"ReplacePinByRequestId",
		http.MethodPost,
		"/pins/:requestid",
		ReplacePinByRequestId,
	},
}
