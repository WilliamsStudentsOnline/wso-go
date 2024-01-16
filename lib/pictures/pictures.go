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
	DoesUserPhotoExists(unixID string) (bool, error)
	// SaveUserPhotoBoth saves both `large` and `thumb` user profile, by resizing
	// img. Large photos are resized to 300x300. Thumb photos are resized to 50x50.
	SaveUserPhotoBoth(unixID string, img image.Image) error
	SaveUserPhotoLarge(unixID string, img image.Image) error
	SaveUserPhotoThumb(unixID string, img image.Image) error
	SaveEphmatchPhoto(unixID string, img image.Image) error
	DeleteEphmatchPhoto(unixID string) error
	SaveDormRoom(dormRoomID uint, reviewID uint, img image.Image) error
	ListDormRoom(dormRoomID uint) ([]string, error)
}

func NewPictureBackend(cfg *config.Config, log *zap.SugaredLogger) (PictureBackend, error) {
	switch cfg.PictureBackend {
	case PictureBackendLocal:
		return local.NewBackend(cfg.PictureLocalPath, log.Named("pictures_backend"))
	case PictureBackendNone:
		return &pictureBackendDummy{}, nil
	default:
		return nil, errors.New("unknown picture backend specified")
	}
}

func IsErrorMaxDormtrakPhotos(err error) bool {
	return err.Error() == local.ErrorMaxDormtrakPhotos.Error()
}
