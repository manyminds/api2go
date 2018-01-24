// +build gingonic,!gorillamux,!echo

package api2go

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/manyminds/api2go/routing"
)

func newTestRouter() routing.Routeable {
	gin.SetMode(gin.ReleaseMode)
	gg := gin.Default()
	notFound := func(c *gin.Context) {
		notAllowedHandler{}.ServeHTTP(c.Writer, c.Request)
	}

	gg.NoRoute(notFound)

	return routing.Gin(gg)
}

func init() {
	log.Println("Testing with gin router")
}
