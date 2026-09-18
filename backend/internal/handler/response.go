package handler

import (
	"github.com/blueship581/mindgarden/backend/internal/constants"
	"github.com/blueship581/mindgarden/backend/internal/dto"
	"github.com/blueship581/mindgarden/backend/internal/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
)

var validate = validator.New()

func ok(c *gin.Context, data any) {
	c.JSON(http.StatusOK, dto.Response{Code: constants.CodeOK, Message: constants.MessageOK, Data: data})
}
func created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, dto.Response{Code: constants.CodeOK, Message: constants.MessageCreated, Data: data})
}
func bind(c *gin.Context, dst any) bool {
	if e := c.ShouldBindJSON(dst); e != nil {
		c.Error(util.NewAppError(constants.CodeValidation, "Request[body] validation failed: malformed JSON", e))
		return false
	}
	if e := validate.Struct(dst); e != nil {
		c.Error(util.NewAppError(constants.CodeValidation, "Request[body] validation failed: required fields invalid", e))
		return false
	}
	return true
}
