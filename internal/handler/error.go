package handler

import (
	"fmt"
	"go-map-proxy/internal/model"
	"net/http"

	"github.com/labstack/echo/v4"
)

func CustomGlobalHTTPErrorHandler(err error, c echo.Context) {
	// Check if the response has already been committed
	// If it has, we will skip the error handling
	if c.Response().Committed {
		return
	}

	code := http.StatusInternalServerError
	// type assertion
	// get echo.HTTPError status code
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
	}
	c.Logger().Error(fmt.Sprintf("CustomGlobalHTTPErrorHandler error: %v code: %d", err, code))

	c.JSON(code, model.BaseAPIResponse[any]{
		Code:    code,
		Message: fmt.Sprintf("Global error handler: %v", err),
		Data:    nil,
	})

}
