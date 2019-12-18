package pictures

import (
	"errors"
	"image"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures/local"
	"go.uber.org/zap"
)

const (
	PictureBackendLocal = "local"
	PictureBackendNone  = "none"
)

type PictureBackend interface {
	SaveThumb(img image.Image, unixID string) error
	SaveLarge(img image.Image, unixID string) error
	Save(img image.Image, unixID string, category string) error
}

func NewPictureBackend(cfg *config.Config, log *zap.SugaredLogger) (PictureBackend, error) {
	switch cfg.PictureBackend {
	case PictureBackendLocal:
		return local.NewBackend(cfg.PictureLocalPath, log)
	case PictureBackendNone:
		return &pictureBackendDummy{}, nil
	default:
		return nil, errors.New("unknown picture backend specified")
	}
}
