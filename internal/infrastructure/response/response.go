package response

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ErrorResponse Error Response
func ErrorResponse(ctx *gin.Context, message string, reason error, isDevelopment bool) {
	response := gin.H{
		"status":  "error",
		"message": message,
	}

	if reason != nil && isDevelopment {
		response["reason"] = reason.Error()
	}

	ctx.JSON(http.StatusBadRequest, response)
}

func SuccessWithStatusResponse(ctx *gin.Context, code int, message string, data interface{}) {
	response := gin.H{
		"status":  strconv.Itoa(code),
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}

	ctx.JSON(code, response)
}

// SuccessResponse Success Response
func SuccessResponse(ctx *gin.Context, message string, data interface{}) {
	response := gin.H{
		"status":  "success",
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}

	ctx.JSON(http.StatusOK, response)
}

func Success220(ctx *gin.Context, message string, data interface{}) {
	response := gin.H{
		"status":  "success",
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}

	ctx.JSON(220, response)
}

func SuccessResponseData(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusOK, data)
}

// GenericResponse Generic Response with status code
func GenericResponse(ctx *gin.Context, code int, message string, data any) {
	response := gin.H{
		"status":  strconv.Itoa(code),
		"message": message,
	}
	if data != nil {
		response["data"] = data
	}
	ctx.JSON(code, response)
}

func CreatedResponseData(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusCreated, data)
}

func TeapotResponse(ctx *gin.Context, data any) {
	ctx.JSON(http.StatusTeapot, data)
}

// FailedResponse Failed Response
func FailedResponse(ctx *gin.Context, message string, reason error, isDevelopment bool) {
	response := gin.H{
		"status":  "failed",
		"message": message,
	}

	if reason != nil && isDevelopment {
		response["reason"] = reason.Error()
	}

	ctx.JSON(http.StatusBadRequest, response)
}

func FailedResponseGeneric(ctx *gin.Context) {
	response := gin.H{
		"status":  "failed",
		"message": "Something went wrong. Try again",
	}
	ctx.JSON(http.StatusBadRequest, response)
}

// NotFoundResponse NotFound Response
func NotFoundResponse(ctx *gin.Context, message string, reason error, isDevelopment bool) {
	response := gin.H{
		"status":  "not found",
		"message": message,
	}

	if reason != nil && isDevelopment {
		response["reason"] = reason.Error()
	}
	ctx.JSON(http.StatusNotFound, response)
}

// UnauthorizedResponse Unauthorized Response
func UnauthorizedResponse(ctx *gin.Context, message string, reason error, isDevelopment bool) {

	response := gin.H{
		"status":  "unauthorized",
		"message": message,
	}

	if reason != nil && isDevelopment {
		response["reason"] = reason.Error()
	}

	ctx.JSON(http.StatusUnauthorized, response)
}

// ForbiddenResponse Forbidden Response
func ForbiddenResponse(ctx *gin.Context, message string, reason error, isDevelopment bool) {
	response := gin.H{
		"status":  "forbidden",
		"message": message,
	}

	if reason != nil && isDevelopment {
		response["reason"] = reason.Error()
	}

	ctx.JSON(http.StatusForbidden, response)
}

func LockedResponse(ctx *gin.Context, message string, reason error, isDevelopment bool) {
	response := gin.H{
		"status":  "locked",
		"message": message,
	}
	if reason != nil && isDevelopment {
		response["reason"] = reason.Error()
	}
	ctx.JSON(http.StatusLocked, response)
}
