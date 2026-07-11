package router

import (
	"mindex-api/core/handler"
	"mindex-api/core/middleware"
	"mindex-api/core/repository"
	"mindex-api/core/service"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Dependencies struct {
	Pool          *pgxpool.Pool
	AdminPassword string
	CORSOrigin    string
}

func SetupRouter(deps Dependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	r.Use(middleware.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.CORS(deps.CORSOrigin))

	entryRepo := repository.NewPgxEntryRepository(deps.Pool)
	entryService := service.NewEntryService(entryRepo)
	loginService := service.NewLoginService(deps.AdminPassword)

	entryHandler := handler.NewEntryHandler(entryService)
	loginHandler := handler.NewLoginHandler(loginService)

	r.GET("/health", handler.Health)

	api := r.Group("/api")
	{
		api.GET("/entries", entryHandler.List)
		api.GET("/categories", entryHandler.ListByCategories)
		api.POST("/entries", middleware.Auth(deps.AdminPassword), entryHandler.Create)
		api.PUT("/entries", middleware.Auth(deps.AdminPassword), entryHandler.Update)
		api.DELETE("/entries", middleware.Auth(deps.AdminPassword), entryHandler.Delete)
		api.POST("/entries/archive", middleware.Auth(deps.AdminPassword), entryHandler.Archive)
		api.POST("/entries/unarchive", middleware.Auth(deps.AdminPassword), entryHandler.Unarchive)
		api.POST("/login", loginHandler.Login)
		api.POST("/logout", loginHandler.Logout)
	}

	return r
}
