package common

import (
	"fmt"
	"go-map-proxy/internal/model"
	"go-map-proxy/pkg/system"

	"github.com/labstack/echo/v4"
)

func SystemInfo(c echo.Context) error {
	systemInfo, err := system.GetSystemInfo()
	if err != nil {
		return c.JSON(500, model.BaseAPIResponse[any]{
			Code:    500,
			Message: fmt.Sprintf("Get system info error: %v", err),
			Data:    nil,
		})
	}
	return c.JSON(200, model.BaseAPIResponse[*system.SystemInfo]{
		Code:    200,
		Message: "Get system info success",
		Data:    systemInfo,
	})

}
