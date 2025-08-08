package api

import (
	"github.com/gqvz/mvc/pkg/config"
	"github.com/gqvz/mvc/pkg/models"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gqvz/mvc/pkg/controllers"
	"github.com/gqvz/mvc/pkg/middlewares"
)

func CreateRouter(appConfig *config.AppConfig) *mux.Router {
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

	return router
}

func RegisterRoutes(router *mux.Router) {
	RegisterUserRoutes(router)
	RegisterTokenRoutes(router)
	RegisterTagRoutes(router)
	RegisterRequestRoutes(router)
	RegisterItemRoutes(router)
	RegisterOrderRoutes(router)
	RegisterOrderItemRoutes(router)
	RegisterPaymentRoutes(router)
}

func RegisterPaymentRoutes(router *mux.Router) {
	c := controllers.CreatePaymentController()
	createPaymentHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.CreatePaymentHandler))
	router.Handle("/payments", createPaymentHandler).Methods("POST")

	getPaymentHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetPaymentHandler))
	router.Handle("/payments/{id:[0-9]+}", getPaymentHandler).Methods("GET")

	getPaymentsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetPaymentsHandler))
	router.Handle("/payments", getPaymentsHandler).Methods("GET")

	editPaymentStatusHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.EditPaymentStatusHandler))
	router.Handle("/payments/{id:[0-9]+}", editPaymentStatusHandler).Methods("PATCH")
}

func RegisterOrderItemRoutes(router *mux.Router) {
	c := controllers.CreateOrderItemController()
	createOrderItemHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.CreateOrderItem))
	router.Handle("/orders/{id:[0-9]+}/items", createOrderItemHandler).Methods("POST")

	editOrderItemStatusHandler := middlewares.Authorize(models.Chef)(http.HandlerFunc(c.EditOrderItemStatus))
	router.Handle("/orders/items/{id:[0-9]+}/", editOrderItemStatusHandler).Methods("PATCH")

	getOrderItemsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetOrderItems))
	router.Handle("/orders/{id:[0-9]+}/items", getOrderItemsHandler).Methods("GET")

	getOrderItemsByStatusHandler := middlewares.Authorize(models.Chef)(http.HandlerFunc(c.GetOrderItemsByStatus))
	router.Handle("/orders/items", getOrderItemsByStatusHandler).Methods("GET")
}

func RegisterOrderRoutes(router *mux.Router) {
	c := controllers.CreateOrderController()
	createOrderHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.CreateOrder))
	router.Handle("/orders", createOrderHandler).Methods("POST")

	closeOrderHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.CloseOrder))
	router.Handle("/orders/{id:[0-9]+}/close", closeOrderHandler).Methods("POST")

	getOrderHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetOrder))
	router.Handle("/orders/{id:[0-9]+}", getOrderHandler).Methods("GET")

	getOrdersHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetOrders))
	router.Handle("/orders", getOrdersHandler).Methods("GET")
}

func RegisterItemRoutes(router *mux.Router) {
	c := controllers.CreateItemController()
	createItemHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.CreateItemHandler))
	router.Handle("/items", createItemHandler).Methods("POST")

	getItemHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetItemHandler))
	router.Handle("/items/{id:[0-9]+}", getItemHandler).Methods("GET")

	getItemsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetItemsHandler))
	router.Handle("/items", getItemsHandler).Methods("GET")

	editTagHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.EditItemHandler))
	router.Handle("/items/{id:[0-9]+}", editTagHandler).Methods("PUT")
}

func RegisterRequestRoutes(router *mux.Router) {
	c := controllers.CreateRequestController()
	router.HandleFunc("/requests", c.CreateRequestHandler).Methods("POST")

	router.HandleFunc("/requests", c.GetRequestsHandler).Methods("GET")

	grantRequestHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.GrantRequestHandler))
	router.Handle("/requests/{id:[0-9]+}/grant", grantRequestHandler).Methods("POST")

	rejectRequestHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.RejectRequestHandler))
	router.Handle("/requests/{id:[0-9]+}/reject", rejectRequestHandler).Methods("POST")

	router.HandleFunc("/requests/{id:[0-9]+}/seen", c.MarkRequestSeenHandler).Methods("POST")
}

func RegisterTagRoutes(router *mux.Router) {
	c := controllers.CreateTagController()
	createTagHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.CreateTagHandler))
	router.Handle("/tags", createTagHandler).Methods("POST")

	getTagsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetTagsHandler))
	router.Handle("/tags", getTagsHandler).Methods("GET")

	getTagHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetTagHandler))
	router.Handle("/tags/{id:[0-9]+}", getTagHandler).Methods("GET")

	editTagHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.EditTagHandler))
	router.Handle("/tags/{id:[0-9]+}", editTagHandler).Methods("PUT")
}

func RegisterTokenRoutes(router *mux.Router) {
	c := controllers.CreateTokenController()
	router.HandleFunc("/token", c.CreateTokenHandler).Methods("POST")
}

func RegisterUserRoutes(router *mux.Router) {
	uc := controllers.CreateUserController()
	router.HandleFunc("/users/", uc.CreateUserHandler).Methods("POST")

	editUserHandler := middlewares.Authorize(models.Any)(http.HandlerFunc(uc.EditUserHandler))
	router.Handle("/users/{id}", editUserHandler).Methods("PATCH")

	getUserHandler := middlewares.Authorize(models.Any)(http.HandlerFunc(uc.GetUserHandler))
	router.Handle("/users/{id}", getUserHandler).Methods("GET")

	getUsersHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(uc.GetUsersHandler))
	router.Handle("/users", getUsersHandler).Methods("GET")
}
