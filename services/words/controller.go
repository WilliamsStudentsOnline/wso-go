package words

import (
	"math/rand"

	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type Controller struct {
	services.BaseController
}

//go:generate go run words_gen.go

// NewController constructs a new words controller
func NewController(db *gorm.DB) *Controller {
	return &Controller{}
}

// GetWords godoc
// @Summary Get words
// @Description gets three WSO words
// @ID words-get-words
// @Tags words
// @Accept  json
// @Produce  json
// @Success 200 {object} string
// @Failure 500 {object} lib.APIError
// @Security Bearer
// @Router /words [get]
func (t *Controller) GetWords(c *gin.Context) {
	w := wWords[rand.Intn(len(wWords))]
	s := sWords[rand.Intn(len(sWords))]
	o := oWords[rand.Intn(len(oWords))]
	t.RespondOK(c, w+" "+s+" "+o)
}
