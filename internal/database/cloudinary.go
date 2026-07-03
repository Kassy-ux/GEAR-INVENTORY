package database

import (
	"github.com/cloudinary/cloudinary-go/v2"
	"inventory-system/internal/config"
)

func NewCloudinary(cfg *config.Config) (*cloudinary.Cloudinary, error) {
	cld, err := cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
	if err != nil {
		return nil, err
	}
	return cld, nil
}
