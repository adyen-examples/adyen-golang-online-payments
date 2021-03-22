package web

import (
	"os"

	"github.com/adyen/adyen-go-api-library/v5/src/adyen"
	"github.com/adyen/adyen-go-api-library/v5/src/common"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	client          *adyen.APIClient
	merchantAccount string
)

func Init() {
	godotenv.Load("./.env")

	client = adyen.NewClient(&common.Config{
		ApiKey:      os.Getenv("API_KEY"),
		Environment: common.TestEnv,
	})

	merchantAccount = os.Getenv("MERCHANT_ACCOUNT")

	// Set the router as the default one shipped with Gin
	router := gin.Default()
	// Serve HTML templates
	router.LoadHTMLGlob("./templates/*")
	// Serve frontend static files
	router.Use(static.Serve("/static", static.LocalFile("./static", true)))

	// setup client side templates
	router.GET("/", IndexHandler)
	router.GET("/preview/:type", PreviewHandler)
	router.GET("/checkout/:type", CheckoutHandler)
	router.GET("/result/:status", ResultHandler)

	// Setup route group and routes for the API
	api := router.Group("/api")

	api.POST("/getPaymentMethods", PaymentMethodsHandler)
	api.POST("/initiatePayment", PaymentsHandler)
	api.POST("/submitAdditionalDetails", PaymentDetailsHandler)
	// handle redirects
	api.GET("/handleShopperRedirect", RedirectHandler)
	api.POST("/handleShopperRedirect", RedirectHandler)

	// Start and run the server
	router.Run(":3000")
}
