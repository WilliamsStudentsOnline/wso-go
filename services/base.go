package services

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

type BaseController struct{}

type BaseResponse struct {
	Status int         `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  *RespError  `json:"error,omitempty"`
}

type RespError struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

// Respond to a request with an OK and some data
func (BaseController) RespondOK(data interface{}, c *gin.Context) {
	c.JSON(http.StatusOK, BaseResponse{
		Status: http.StatusOK,
		Data:   data,
		Error:  nil,
	})
}

// Base controller object for outside packages to call to access methods
var Base = BaseController{}

// Respond to request with an error and abort
func (BaseController) RespondError(code int, err error, c *gin.Context) {
	// If the error is that the DB could not find a record, return a not found error
	if gorm.IsRecordNotFoundError(err) {
		code = http.StatusNotFound
	}

	// We don't want clients knowing what's happening here, so just log the error and return something inconspicuous
	if code >= http.StatusInternalServerError {
		_ = c.Error(err)
		err = errors.New("internal server error")
	}

	c.AbortWithStatusJSON(code, BaseResponse{
		Status: code,
		Error: &RespError{
			ErrorCode: code,
			Message:   err.Error(),
		},
	})
}

// Get a parameter that is a uint
func GetUIntParam(key string, ctx *gin.Context) (uint, error) {
	param := ctx.Param(key)
	paramInt, err := strconv.Atoi(param)
	if err != nil {
		return 0, err
	}

	return uint(paramInt), nil
}

// Get the User ID from context store
func GetUserID(ctx *gin.Context) uint {
	return (ctx.MustGet("user_id")).(uint)
}
