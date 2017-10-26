// +build gingonic,!gorillamux

package api2go

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/manyminds/api2go/routing"
)

func newTestRouter() routing.Routeable {
	gin.SetMode(gin.ReleaseMode)
	gg := gin.Default()
	return routing.Gin(gg)
}

func init() {
	log.Println("Testing with gin router")
}
