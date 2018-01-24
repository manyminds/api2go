// +build !gingonic,!echo,gorillamux

package api2go

import (
	"log"

	"github.com/gorilla/mux"
	"github.com/manyminds/api2go/routing"
)

func newTestRouter() routing.Routeable {
	router := mux.NewRouter()
	router.MethodNotAllowedHandler = notAllowedHandler{}
	return routing.Gorilla(router)
}

func init() {
	log.Println("Testing with gorilla router")
}
