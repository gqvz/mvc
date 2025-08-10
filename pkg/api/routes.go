package api

import (
	"github.com/gqvz/mvc/pkg/config"
	"github.com/gqvz/mvc/pkg/models"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"regexp"

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

	router.Use(corsMiddleware)
	apiRouter := router.PathPrefix("/api").Subrouter()

	authMiddleware := middlewares.CreateAuthenticationMiddleware(appConfig.JwtSecret)
	apiRouter.Use(authMiddleware)

	RegisterRoutes(apiRouter)

	return router
}

func corsMiddleware(next http.Handler) http.Handler {
	var localhostRegex = regexp.MustCompile(`^https?://localhost(:[0-9]+)?$`)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if localhostRegex.MatchString(origin) {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, PATCH")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		}
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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
	router.Handle("/payments", createPaymentHandler).Methods("POST", "OPTIONS")

	getPaymentHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetPaymentHandler))
	router.Handle("/payments/{id:[0-9]+}", getPaymentHandler).Methods("GET", "OPTIONS")

	getPaymentsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetPaymentsHandler))
	router.Handle("/payments", getPaymentsHandler).Methods("GET", "OPTIONS")

	editPaymentStatusHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.EditPaymentStatusHandler))
	router.Handle("/payments/{id:[0-9]+}", editPaymentStatusHandler).Methods("PATCH", "OPTIONS")
}

func RegisterOrderItemRoutes(router *mux.Router) {
	c := controllers.CreateOrderItemController()
	createOrderItemHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.CreateOrderItem))
	router.Handle("/orders/{id:[0-9]+}/items", createOrderItemHandler).Methods("POST", "OPTIONS")

	editOrderItemStatusHandler := middlewares.Authorize(models.Chef)(http.HandlerFunc(c.EditOrderItemStatus))
	router.Handle("/orders/items/{id:[0-9]+}/", editOrderItemStatusHandler).Methods("PATCH", "OPTIONS")

	getOrderItemsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetOrderItems))
	router.Handle("/orders/{id:[0-9]+}/items", getOrderItemsHandler).Methods("GET", "OPTIONS")

	getOrderItemsByStatusHandler := middlewares.Authorize(models.Chef)(http.HandlerFunc(c.GetOrderItemsByStatus))
	router.Handle("/orders/items", getOrderItemsByStatusHandler).Methods("GET", "OPTIONS")
}

func RegisterOrderRoutes(router *mux.Router) {
	c := controllers.CreateOrderController()
	createOrderHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.CreateOrder))
	router.Handle("/orders", createOrderHandler).Methods("POST", "OPTIONS")

	closeOrderHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.CloseOrder))
	router.Handle("/orders/{id:[0-9]+}/close", closeOrderHandler).Methods("POST", "OPTIONS")

	getOrderHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetOrder))
	router.Handle("/orders/{id:[0-9]+}", getOrderHandler).Methods("GET", "OPTIONS")

	getOrdersHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetOrders))
	router.Handle("/orders", getOrdersHandler).Methods("GET", "OPTIONS")
}

func RegisterItemRoutes(router *mux.Router) {
	c := controllers.CreateItemController()
	createItemHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.CreateItemHandler))
	router.Handle("/items", createItemHandler).Methods("POST", "OPTIONS")

	getItemHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetItemHandler))
	router.Handle("/items/{id:[0-9]+}", getItemHandler).Methods("GET", "OPTIONS")

	getItemsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetItemsHandler))
	router.Handle("/items", getItemsHandler).Methods("GET", "OPTIONS")

	editItemHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.EditItemHandler))
	router.Handle("/items/{id:[0-9]+}", editItemHandler).Methods("PUT", "OPTIONS")
}

func RegisterRequestRoutes(router *mux.Router) {
	c := controllers.CreateRequestController()
	router.HandleFunc("/requests", c.CreateRequestHandler).Methods("POST", "OPTIONS")

	router.HandleFunc("/requests", c.GetRequestsHandler).Methods("GET", "OPTIONS")

	grantRequestHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.GrantRequestHandler))
	router.Handle("/requests/{id:[0-9]+}/grant", grantRequestHandler).Methods("POST", "OPTIONS")

	rejectRequestHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.RejectRequestHandler))
	router.Handle("/requests/{id:[0-9]+}/reject", rejectRequestHandler).Methods("POST", "OPTIONS")

	router.HandleFunc("/requests/{id:[0-9]+}/seen", c.MarkRequestSeenHandler).Methods("POST", "OPTIONS")
}

func RegisterTagRoutes(router *mux.Router) {
	c := controllers.CreateTagController()
	createTagHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.CreateTagHandler))
	router.Handle("/tags", createTagHandler).Methods("POST", "OPTIONS")

	getTagsHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetTagsHandler))
	router.Handle("/tags", getTagsHandler).Methods("GET", "OPTIONS")

	getTagHandler := middlewares.Authorize(models.Customer)(http.HandlerFunc(c.GetTagHandler))
	router.Handle("/tags/{id:[0-9]+}", getTagHandler).Methods("GET", "OPTIONS")

	editTagHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(c.EditTagHandler))
	router.Handle("/tags/{id:[0-9]+}", editTagHandler).Methods("PUT", "OPTIONS")
}

func RegisterTokenRoutes(router *mux.Router) {
	c := controllers.CreateTokenController()
	router.HandleFunc("/token", c.CreateTokenHandler).Methods("POST", "OPTIONS")
}

func RegisterUserRoutes(router *mux.Router) {
	uc := controllers.CreateUserController()
	router.HandleFunc("/users/", uc.CreateUserHandler).Methods("POST", "OPTIONS")

	editUserHandler := middlewares.Authorize(models.Any)(http.HandlerFunc(uc.EditUserHandler))
	router.Handle("/users/{id}", editUserHandler).Methods("PATCH", "OPTIONS")

	getUserHandler := middlewares.Authorize(models.Any)(http.HandlerFunc(uc.GetUserHandler))
	router.Handle("/users/{id}", getUserHandler).Methods("GET", "OPTIONS")

	getUsersHandler := middlewares.Authorize(models.Admin)(http.HandlerFunc(uc.GetUsersHandler))
	router.Handle("/users", getUsersHandler).Methods("GET", "OPTIONS")
}
