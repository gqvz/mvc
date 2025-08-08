package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/config"
	"github.com/gqvz/mvc/pkg/controllers"
	"github.com/gqvz/mvc/pkg/middlewares"
	httpSwagger "github.com/swaggo/http-swagger"
)

func CreateHTTPServer(appConfig *config.AppConfig) *http.Server {
	router := mux.NewRouter()

	router.PathPrefix("/swagger/").Handler(httpSwagger.Handler(
		httpSwagger.URL("/swagger/doc.json"),
		httpSwagger.DeepLinking(true),
		httpSwagger.DocExpansion("none"),
		httpSwagger.DomID("swagger-ui"),
	))

	apiRouter := router.PathPrefix("/api").Subrouter()

	authMiddleware := middlewares.CreateAuthenticationMiddleware(appConfig.JwtSecret)
	apiRouter.Use(authMiddleware)

	RegisterRoutes(apiRouter)

	return &http.Server{
		Addr:    appConfig.ServerAddress,
		Handler: router,
	}
}

func RegisterRoutes(router *mux.Router) {
	userController := controllers.CreateUserController()
	userController.RegisterRoutes(router)

	authController := controllers.CreateTokenController()
	authController.RegisterRoutes(router)

	tagController := controllers.CreateTagController()
	tagController.RegisterRoutes(router)

	itemController := controllers.CreateItemController()
	itemController.RegisterRoutes(router)

	requestController := controllers.CreateRequestController()
	requestController.RegisterRoutes(router)

	orderController := controllers.CreateOrderController()
	orderController.RegisterRoutes(router)

	orderItemController := controllers.CreateOrderItemController()
	orderItemController.RegisterRoutes(router)

	paymentController := controllers.CreatePaymentController()
	paymentController.RegisterRoutes(router)
}
