// +build !gingonic,gorillamux

package api2go

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/manyminds/api2go/routing"
)

func newTestRouter() routing.Routeable {
	return routing.Gorilla(mux.NewRouter())
}

func init() {
	log.Println("Testing with gorilla router")
}
