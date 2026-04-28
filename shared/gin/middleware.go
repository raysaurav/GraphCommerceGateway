package gin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raysaurav/GraphCommerceGateway/shared/constant"
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

type (
	// Define a custom type for the context key
	correlationIDKeyType string
	// Define a custom type for the context key
	contextKey      string
	usidKey         string
	accessToken     string
	customerId      string
	basketId        string
	authorization   string
	environment     string
	appSessionIDKey string
)

const (
	correlationIDKey  correlationIDKeyType = constant.CorrelationIDHeader
	UsidKey           usidKey              = constant.UsidHeader
	GinContextKey     contextKey           = constant.GinContextKey
	XForwardedFor_Key contextKey           = constant.XForwardedForKey
	AccessToken       accessToken          = constant.ACCESS_TOKEN
	CustomerID        customerId           = constant.CUSTOMER_ID
	BasketID          basketId             = constant.BASKET_ID
	XLivePreview_Key  contextKey           = constant.XLivePreviewKey
	XChannel_Key      contextKey           = constant.XChannelKey
	XStoreId_Key      contextKey           = constant.XStoreIdKey
	XUserAgent_Key    contextKey           = constant.XUserAgentKey
	XAppSessionID_Key appSessionIDKey      = constant.XAppSessionIDKey
	Authorization     authorization        = constant.AUTHORIZATION
	Environment       environment          = constant.ENVIRONMENT
	XCacheCDNKey      correlationIDKeyType = constant.XCacheCDN
)

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
			var body map[string]interface{}
			err := ctx.ShouldBindBodyWithJSON(&body)
			if err != nil {
				fmt.Println("Error while JSON Changes", err)
				ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"errors": "Invalid JSON Format"})
				return
			}

			// Convert sanitized map back to JSON
			jsonBytes, err := json.Marshal(body)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
				return
			}

			// CRITICAL: Restore the body for downstream handlers
			ctx.Request.Body = io.NopCloser(bytes.NewBuffer(jsonBytes))
		}

		ctx.Next()
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

func UUIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.Request.Header.Get(string(UsidKey))
		ctx := context.WithValue(c.Request.Context(), UsidKey, accessToken)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GeoIpMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		xForwardedFor := c.Request.Header.Get("x-forwarded-for")
		ctx := context.WithValue(c.Request.Context(), XForwardedFor_Key, xForwardedFor)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func CustomerIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		customerId := c.Request.Header.Get(string(CustomerID))
		ctx := context.WithValue(c.Request.Context(), CustomerID, customerId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func AccessTokenMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		accessToken := c.Request.Header.Get(string(AccessToken))
		ctx := context.WithValue(c.Request.Context(), AccessToken, accessToken)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AuthorizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authorization := c.Request.Header.Get(string(Authorization))
		ctx := context.WithValue(c.Request.Context(), Authorization, authorization)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func BasketIdMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		basketId := c.Request.Header.Get(string(BasketID))
		ctx := context.WithValue(c.Request.Context(), BasketID, basketId)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func AppSessionIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		appSessionID := c.Request.Header.Get(string(XAppSessionID_Key))
		ctx := context.WithValue(c.Request.Context(), XAppSessionID_Key, appSessionID)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
func GinContextToContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := context.WithValue(c.Request.Context(), GinContextKey, c)
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func GinContextFromContext(ctx context.Context) (*gin.Context, error) {
	ginContext := ctx.Value(GinContextKey)
	if ginContext == nil {
		err := fmt.Errorf("could not retrieve gin.Context")
		return nil, err
	}
	gc, ok := ginContext.(*gin.Context)
	if !ok {
		err := fmt.Errorf("gin.Context has wrong type")
		return nil, err
	}
	return gc, nil
}
