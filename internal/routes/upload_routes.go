package routes

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/labstack/echo/v4"
	"inventory-system/internal/handlers"
)

func RegisterUploadRoutes(e *echo.Echo, cld *cloudinary.Cloudinary) {
	h := handlers.NewUploadHandler(cld)
	e.POST("/upload", h.UploadImage)
}
