package gin

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AppError represents a structured API error.
type AppError struct {
	Status  int     `json:"-"`
	Code    string  `json:"code"`
	Message string  `json:"message"`
	Details []error `json:"details,omitempty"`
}

func (e *AppError) Error() string {
	return e.Message
}

func ErrorHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		// after all the processing if there is a error then process it
		if len(ctx.Errors) > 0 {
			// It means we have error
			var errList []interface{}
			for _, ginErr := range ctx.Errors {
				appErr, ok := ginErr.Err.(*AppError)
				if ok {
					errList = append(errList, appErr)
				} else {
					errList = append(errList, &AppError{
						Status:  500,
						Message: "Internal Server error",
						Details: []error{ginErr},
					})
				}
			}
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"errors": errList})
		}
	}
}

func MapJSONData() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodPost {
			body := make(map[string]interface{})
			err := ctx.ShouldBindBodyWithJSON(&body)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": "Invalid JSON Format"})
				return
			}

			ctx.Next()
		}
	}
}

func AddCorsPolicy() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Writer.Header().Add("Access-Control-Allow-Origin", "*")
		ctx.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, "+string(UsidKey)+", "+string(AccessToken))
		if ctx.Request.Method == "OPTIONS" {
			ctx.AbortWithStatus(204)
			return
		}
		ctx.Set("header", ctx.Request.Header)

		ctx.Next()
	}
}

// GlobalExceptionHandler is your existing middleware
func GlobalExceptionHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				fmt.Printf("Recovered from panic: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
				// Optionally log the stack trace
			}
		}()

		c.Next()
	}
}
