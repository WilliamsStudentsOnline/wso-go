package chat

import (
	"io/ioutil"
	"time"

	"github.com/WilliamsStudentsOnline/wso-go/config"
	"github.com/WilliamsStudentsOnline/wso-go/models"
	"github.com/WilliamsStudentsOnline/wso-go/services"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

type Controller struct {
	services.BaseController
	cfg       *config.Config
	userModel *models.UserModel
	jwtKey    interface{}
}

// Construct a new user controller
func NewController(db *gorm.DB, cfg *config.Config, log *zap.SugaredLogger) *Controller {
	var key interface{}
	if usingPublicKeyAlgo(cfg) {
		priv, _, err := readKeys(cfg)
		if err != nil {
			log.Error("Key parsing failed", err)
			return nil
		}
		key = priv
	} else {
		key = []byte(cfg.Secrets.JWTSecretKey)
	}

	return &Controller{
		BaseController: services.BaseController{Log: log},
		cfg:            cfg,
		userModel:      models.NewUserModel(db, log),
		jwtKey:         key,
	}
}

// Checks if the passed API token is valid and if user should be authenticated.
// Currently, it ensures that the user is a student, just because of the current use exclusively as an Ephmatch feature
// GetChatAuthToken godoc
// @Summary Get Chat Auth JWT token
// @Description gets auth JWT token of chat feature
// @ID chat-get-auth-token
// @Tags chat
// @Accept  json
// @Produce  json
// @Success 200 {string} string token
// @Failure 400 {object} services.BaseErrorResponse
// @Failure 404 {object} services.BaseErrorResponse
// @Failure 500 {object} services.BaseErrorResponse
// @Security Bearer
// @Router /chat/auth/token [get]
func (t *Controller) GetChatAuthToken(c *gin.Context) {
	userID := services.GetUserID(c)

	// Do database query
	var user models.User
	err := t.userModel.GetUserByID(userID, &user)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jid": user.UnixID + "@" + t.cfg.ChatEjabberdName,
		"exp": time.Now().Add(45 * 24 * time.Hour).Unix(),
	})

	tokenString, err := token.SignedString(t.jwtKey)
	if err != nil {
		t.RespondError(c, err)
		return
	}

	t.RespondOK(c, tokenString)
}

func readKeys(cfg *config.Config) (privKey interface{}, pubKey interface{}, err error) {
	privKey, err = privateKey(cfg)
	if err != nil {
		return
	}
	pubKey, err = publicKey(cfg)
	if err != nil {
		return
	}
	return
}

func privateKey(cfg *config.Config) (interface{}, error) {
	keyData, err := ioutil.ReadFile(cfg.JWTPrivateKeyFile)
	if err != nil {
		return nil, err
	}

	var key interface{}
	switch cfg.JWTSigningAlgo {
	case "RS256", "RS512", "RS384":
		key, err = jwt.ParseRSAPrivateKeyFromPEM(keyData)
	case "ES256", "ES384", "ES512":
		key, err = jwt.ParseECPrivateKeyFromPEM(keyData)
	}
	if err != nil || key == nil {
		return nil, err
	}

	return key, nil
}

func publicKey(cfg *config.Config) (interface{}, error) {
	keyData, err := ioutil.ReadFile(cfg.JWTPublicKeyFile)
	if err != nil {
		return nil, err
	}

	var key interface{}
	switch cfg.JWTSigningAlgo {
	case "RS256", "RS512", "RS384":
		key, err = jwt.ParseRSAPublicKeyFromPEM(keyData)
	case "ES256", "ES384", "ES512":
		key, err = jwt.ParseECPublicKeyFromPEM(keyData)
	}
	if err != nil || key == nil {
		return nil, err
	}

	return key, nil
}

func usingPublicKeyAlgo(cfg *config.Config) bool {
	switch cfg.JWTSigningAlgo {
	case "RS256", "RS512", "RS384", "ES256", "ES384", "ES512":
		return true
	}
	return false
}
