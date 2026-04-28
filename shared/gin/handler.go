package gin

import (
	"bytes"
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/gin-gonic/gin"
)

type HandlerInitilization struct {
	Svc *handler.Server
}

var (
	handlerInitilizationInstance *HandlerInitilization
	handlerInitilizationOnce     sync.Once
)

func NewHandlerInitializer(handlerServer *handler.Server) *HandlerInitilization {
	handlerInitilizationOnce.Do(func() {
		handlerInitilizationInstance = &HandlerInitilization{
			Svc: handlerServer,
		}
	})
	return handlerInitilizationInstance
}

func (hi *HandlerInitilization) RegisterHandler(router *gin.Engine) {
	router.Use(GlobalExceptionHandler())
	router.Use(AddCorsPolicy())

	router.POST("/query", func(c *gin.Context) {
		timeout := time.Second * 30
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()
		c.Request = c.Request.WithContext(ctx)

		originalWriter := c.Writer
		buffer := &bytes.Buffer{}

		rcw := &responseCaptureWriter{
			ResponseWriter: originalWriter,
			body:           buffer,
			statusCode:     http.StatusOK,
		}

		c.Writer = rcw

		defer func() {
			err := recover()
			if err != nil {
				c.Writer = originalWriter

				c.Writer.Header().Set("X-Cache-Cdn", "no-cache")
				c.Writer.Header().Set("Content-Type", "application/json")

				// Write 500 response
				c.Writer.WriteHeader(http.StatusInternalServerError)
				c.Writer.Write([]byte(`{"error":"internal server error"}`))

				return
			}

			// Restore original writer and send response
			c.Writer = originalWriter
			c.Writer.WriteHeader(rcw.statusCode)
			c.Writer.Write(buffer.Bytes())
		}()
		// Run GraphQL
		hi.Svc.ServeHTTP(rcw, c.Request)
	})
	// GraphQL playground endpoint
	router.GET("/playground", func(c *gin.Context) {
		playground.Handler("GraphQL Playground", "/query").ServeHTTP(c.Writer, c.Request)
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})

	router.Use(UUIDMiddleware())
	router.Use(GeoIpMiddleware())
	router.Use(CustomerIdMiddleware())
	router.Use(AccessTokenMiddleware())
	router.Use(AppSessionIDMiddleware())
	router.Use(BasketIdMiddleware())
	router.Use(AuthorizationMiddleware())
	router.Use(GinContextToContextMiddleware())
}

type responseCaptureWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseCaptureWriter) Write(b []byte) (int, error) {
	return w.body.Write(b) // Capture response body
}

func (w *responseCaptureWriter) WriteHeader(code int) {
	w.statusCode = code
	// Do not write headers yet
}
