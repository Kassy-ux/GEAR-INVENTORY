package handlers

import (
	"context"
	"net/http"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	Cld *cloudinary.Cloudinary
}

func NewUploadHandler(cld *cloudinary.Cloudinary) *UploadHandler {
	return &UploadHandler{Cld: cld}
}

// POST /upload  (multipart/form-data, field name "image")
func (h *UploadHandler) UploadImage(c echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "image file is required (field name: image)"})
	}

	src, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not open uploaded file"})
	}
	defer src.Close()

	uploadResult, err := h.Cld.Upload.Upload(context.Background(), src, uploader.UploadParams{
		Folder: "inventory-items",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"image_url": uploadResult.SecureURL,
	})
}
