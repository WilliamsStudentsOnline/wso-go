package words

import (
	"math/rand"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
}

//go:generate go run words_gen.go

// NewController constructs a new words controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	return &Controller{
		BaseController: services.BaseController{Log: log},
	}
}

// GetWords godoc
// @Summary Get words
// @Description gets three WSO words
// @ID getWords
// @Tags words
// @Accept  json
// @Produce  json
// @Success 200 {string} string
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /words [get]
func (t *Controller) GetWords(c *gin.Context) {
	w := wWords[rand.Intn(len(wWords))]
	s := sWords[rand.Intn(len(sWords))]
	o := oWords[rand.Intn(len(oWords))]
	t.RespondOK(c, w+" "+s+" "+o)
}
