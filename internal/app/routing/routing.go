package routing

import (
	"context"
	http2 "net/http"

	api "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/docs"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/container"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/app/middleware"
	authDomain "github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/auth/domain"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/logger"
	"github.com/POS-FIAP-15SOAT-TEAM-DARK-MODE/auto-repair-shop-customer-service/internal/pkg/web"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

func SetupRouter(c *container.HandlersWrapper, m *container.Middlewares) *gin.Engine {
	router := gin.New()

	for name, mid := range *m {
		logger.Global().Info("middleware.registered", zap.String("name", name))
		router.Use(mid)
	}

	mountSwaggerUI(router)

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http2.StatusOK, gin.H{"message": "pong"})
	})

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	v1 := router.Group("/v1")

	v1.POST("/auth/register", middleware.Auth(authDomain.ADMIN), GinHandler(c.UserHandler.Register, http2.StatusCreated))
	v1.POST("/auth/login", GinHandler(c.UserHandler.Login))
	v1.PATCH("/users/:id/role", middleware.Auth(authDomain.ADMIN), GinOnlyErrorHandler(c.UserHandler.ChangeRole))

	v1.GET("/customers", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.GetByDocument))
	v1.GET("/customers/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.GetByID))
	v1.POST("/customers", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.Create, http2.StatusCreated))
	v1.PUT("/customers/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.CustomerHandler.Update))
	v1.DELETE("/customers/:id", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.CustomerHandler.Delete))

	v1.GET("/vehicles", middleware.Auth(authDomain.AttendantAndMechanicRoles...), GinHandler(c.VehicleHandler.List))
	v1.POST("/vehicles", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.VehicleHandler.Create, http2.StatusCreated))
	v1.PUT("/vehicles/:id", middleware.Auth(authDomain.AttendantRoles...), GinHandler(c.VehicleHandler.Edit))
	v1.DELETE("/vehicles/:id", middleware.Auth(authDomain.AttendantRoles...), GinOnlyErrorHandler(c.VehicleHandler.Delete))

	// Service-to-service routes. Never routed by the gateway; protected by X-Internal-Token.
	internal := router.Group("/internal", middleware.InternalAuth())
	internal.GET("/customers/:id", GinHandler(c.CustomerInternalHandler.GetInternalByID))
	internal.GET("/vehicles/:id", GinHandler(c.VehicleInternalHandler.GetInternalByID))

	return router
}

func mountSwaggerUI(router *gin.Engine) {
	router.GET("/swagger.yaml", func(c *gin.Context) {
		c.Data(http2.StatusOK, "application/yaml; charset=utf-8", api.SwaggerBytes)
	})

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler, ginSwagger.URL("/swagger.yaml")))
}

// respondError writes the standard error body. The full error (with its root cause) is logged
// for 5XX and for infrastructure conflicts/violations, with the request_id carried by ctx.
func respondError(ctx context.Context, c *gin.Context, err error) {
	status, errResp := web.Error(err)
	if status >= http2.StatusInternalServerError || web.IsInfrastructure(err) {
		logger.Of(ctx).Error(err, zap.Int("status", status))
	}
	c.JSON(status, errResp)
}

func GinOnlyErrorHandler(cb func(context.Context, *http2.Request) error, customStatus ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// context keys are gin param names (strings) intentionally;
		// collision risk is acceptable here since params are route-scoped.
		for _, v := range c.Params {
			ctx = context.WithValue(ctx, v.Key, v.Value) //nolint:staticcheck
		}

		if err := cb(ctx, c.Request); err != nil {
			respondError(ctx, c, err)
			return
		}

		if len(customStatus) == 0 {
			c.Status(http2.StatusNoContent)
			c.Writer.WriteHeaderNow()
			return
		}

		c.Status(customStatus[0])
		c.Writer.WriteHeaderNow()
	}
}

func GinHandler[T any](cb func(context.Context, *http2.Request) (T, error), customStatus ...int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// context keys are gin param names (strings) intentionally;
		// collision risk is acceptable here since params are route-scoped.
		for _, v := range c.Params {
			ctx = context.WithValue(ctx, v.Key, v.Value) //nolint:staticcheck
		}

		resp, err := cb(ctx, c.Request)
		if err != nil {
			respondError(ctx, c, err)
			return
		}

		if len(customStatus) == 0 {
			c.JSON(http2.StatusOK, resp)
			return
		}

		c.JSON(customStatus[0], resp)
	}
}
