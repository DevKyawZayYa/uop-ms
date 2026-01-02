package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		last := c.Errors.Last().Err

		if appErr, ok := last.(*AppError); ok {
			c.JSON(appErr.Status, APIResponse{
				Success: false,
				Error: &APIError{
					Code:    appErr.Code,
					Message: appErr.Message},
			})
			return
		}

		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error: &APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Unexpected error"},
		})
	}
}
