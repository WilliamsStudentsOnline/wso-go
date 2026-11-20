package pic

import (
	"errors"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/lib/pictures"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

var userPhotoPath = regexp.MustCompile(`^user/(?:thumb|large)/([^/]+)\.jpg$`)

type Controller struct {
	services.BaseController
	userModel      *models.UserModel
	pictureBackend pictures.PictureBackend
}

func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	pb, err := pictures.NewPictureBackend(cfg, log)
	if err != nil {
		log.Error(err)
		log.Warn("Using picture backend none")
		pb = pictures.NewPictureBackendDummy()
	}

	return &Controller{
		BaseController: services.BaseController{Log: log},
		userModel:      models.NewUserModel(db, log),
		pictureBackend: pb,
	}
}

// GetPicture serves a file from the picture backend. Paths match what nginx
// previously proxied to Telos: /pic/thumb/x and /pic/large/x are shorthand for
// /pic/user/thumb/x and /pic/user/large/x.
func (t *Controller) GetPicture(c *gin.Context) {
	relPath := strings.TrimPrefix(c.Param("path"), "/")
	if strings.HasPrefix(relPath, "thumb/") || strings.HasPrefix(relPath, "large/") {
		relPath = "user/" + relPath
	}

	if m := userPhotoPath.FindStringSubmatch(relPath); m != nil {
		var user models.User
		err := t.userModel.GetUserByUnixID(m[1], &user)
		if err != nil && !gorm.IsRecordNotFoundError(err) {
			t.RespondError(c, err)
			return
		}
		if err == nil && user.PhotoHidden != nil && *user.PhotoHidden {
			t.respondNotFound(c)
			return
		}
	}

	f, err := t.pictureBackend.OpenPicture(relPath)
	if err != nil {
		if os.IsNotExist(err) {
			t.respondNotFound(c)
			return
		}
		t.RespondError(c, err)
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		t.RespondError(c, err)
		return
	}

	c.Header("Cache-Control", "private, max-age=300")
	http.ServeContent(c.Writer, c.Request, info.Name(), info.ModTime(), f)
}

func (t *Controller) respondNotFound(c *gin.Context) {
	t.RespondErrorCode(c, http.StatusNotFound, errors.New("picture not found"))
}
