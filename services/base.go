package services

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/WilliamsStudentsOnline/wso-go/lib"
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"gopkg.in/go-playground/validator.v8"
)

type BaseController struct{}

type BaseResponse struct {
	Status          int         `json:"status"`
	Data            interface{} `json:"data,omitempty"`
	Error           *RespError  `json:"error,omitempty"`
	UpdateToken     bool        `json:"updateToken,omitempty"`
	PaginationTotal int         `json:"paginationTotal,omitempty"`
}

type RespError struct {
	ErrorCode int      `json:"errorCode"`
	Message   string   `json:"message"`
	Errors    []string `json:"errors,omitempty"`
}

const (
	UpdateTokenKey     = "updateToken"
	PaginationTotalKey = "paginationTotal"
	ErrorCodeKey       = "errorCode"
)

// Respond to a request with an OK and some data
func (BaseController) RespondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, BaseResponse{
		Status: http.StatusOK,
		Data:   data,
		Error:  nil,
		// We set this in the context at any point if we need to update the token
		UpdateToken:     c.GetBool(UpdateTokenKey),
		PaginationTotal: c.GetInt(PaginationTotalKey),
	})
}

// Respond to a request with a no content
func (BaseController) RespondCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, BaseResponse{
		Status: http.StatusCreated,
		Data:   data,
		Error:  nil,
		// We set this in the context at any point if we need to update the token
		UpdateToken:     c.GetBool(UpdateTokenKey),
		PaginationTotal: c.GetInt(PaginationTotalKey),
	})
}

// Base controller object for outside packages to call to access methods
var Base = BaseController{}

func (BaseController) RespondAPIError(c *gin.Context, err *lib.APIError) {
	var errs []string
	if len(err.Errors) > 0 {
		errs = make([]string, len(err.Errors))
		for i, value := range err.Errors {
			errs[i] = value.Error()
		}
	}

	respondError(c, err.HTTPCode, &RespError{
		ErrorCode: err.Code,
		Message:   err.Error(),
		Errors:    errs,
	})
}

// Respond to request with an error and abort
func (b BaseController) RespondError(c *gin.Context, err error) {
	b.RespondErrorCode(c, http.StatusInternalServerError, err)
}

// Respond to request with an error and abort
func (b BaseController) RespondErrorCode(c *gin.Context, code int, err error) {
	// If it is an API error, return like that
	if apiErr, ok := err.(*lib.APIError); ok {
		b.RespondAPIError(c, apiErr)
		return
	}

	// If the error is that the db could not find a record, return a not found error
	if gorm.IsRecordNotFoundError(err) {
		b.RespondAPIError(c, lib.ErrorRecordNotFound)
		return
	}

	// If it is a validation error, format it and send it to respond API error
	if validateErrs, ok := err.(validator.ValidationErrors); ok {
		errs := make([]error, len(validateErrs))
		i := 0
		for field, fieldErr := range validateErrs {
			errs[i] = fmt.Errorf("validation for field %s failed on the '%s' requirement", field, fieldErr.Tag)
			i++
		}
		b.RespondAPIError(c, lib.NewErrorRequestDataValidationFailed(errs))
		return
	}

	// We don't want clients knowing what's happening here, so just log the error and return something inconspicuous
	if code >= http.StatusInternalServerError {
		// Only put the error in the context if it is important
		c.Error(err)
		err = lib.ErrorInternalServerError
	}

	respondError(c, code, &RespError{
		ErrorCode: code,
		Message:   err.Error(),
	})
}

// Respond to request with an error and abort
func respondError(c *gin.Context, httpCode int, err *RespError) {
	c.Set(ErrorCodeKey, err.ErrorCode)
	c.AbortWithStatusJSON(httpCode, BaseResponse{
		Status: err.ErrorCode,
		Error:  err,
		// We set this in the context at any point if we need to update the token
		UpdateToken: c.GetBool(UpdateTokenKey),
	})
}

// Get a parameter that is a uint
func GetUIntParam(c *gin.Context, key string) (uint, error) {
	param := c.Param(key)
	paramInt, err := strconv.Atoi(param)
	if err != nil {
		return 0, err
	}

	return uint(paramInt), nil
}

// Get the User ID from context store
func GetUserID(ctx *gin.Context) uint {
	val, ok := ctx.Get("id")
	if !ok {
		return 0
	}
	return val.(uint)
}

func GetUIntQuery(c *gin.Context, key string) (uint, error) {
	query := c.Query(key)
	queryInt, err := strconv.Atoi(query)
	if err != nil {
		return 0, err
	}

	return uint(queryInt), nil
}

func GetPaginationParams(ctx *gin.Context) (offset, limit int, err error) {
	offsetStr := ctx.Query("offset")
	limitStr := ctx.Query("limit")

	if limitStr == "" {
		return
	}

	if offsetStr != "" {
		offset, err = strconv.Atoi(offsetStr)
		if err != nil {
			return
		}
	}

	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		return
	}

	return
}

func (BaseController) SetUpdateToken(c *gin.Context) {
	c.Set(UpdateTokenKey, true)
}

func (BaseController) SetPaginationTotal(c *gin.Context, total int) {
	c.Set(PaginationTotalKey, total)
}
