// +build echo,!gingonic,!gorillamux

package api2go

import (
	"log"
	"net/http"

	"github.com/labstack/echo"
	"github.com/manyminds/api2go/routing"
)

func customHTTPErrorHandler(err error, c echo.Context) {
	if he, ok := err.(*echo.HTTPError); ok {
		if he == echo.ErrMethodNotAllowed {
			handleError(NewHTTPError(he, "Method Not Allowed", http.StatusMethodNotAllowed), c.Response(), c.Request(), defaultContentTypHeader)
		}
	}
}

func newTestRouter() routing.Routeable {
	e := echo.New()
	// not found handler, this needs to be fixed as well: see: https://github.com/manyminds/api2go/issues/301
	e.HTTPErrorHandler = customHTTPErrorHandler
	return routing.Echo(e)
}

func init() {
	log.Println("Testing with echo router")
}
