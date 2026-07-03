package handlers

import (
	"context"
	"log"
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
		log.Printf("cloudinary upload error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	log.Printf("cloudinary upload result: %+v", uploadResult)

	if uploadResult.SecureURL == "" {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "upload succeeded but no URL was returned - check Cloudinary credentials"})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"image_url": uploadResult.SecureURL,
	})
}
