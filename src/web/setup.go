package web

import (
	"log"
	"os"

	"github.com/adyen/adyen-go-api-library/v5/src/adyen"
	"github.com/adyen/adyen-go-api-library/v5/src/common"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var (
	client          *adyen.APIClient
	port            string
	merchantAccount string
	clientKey       string
)

func Init() {
	godotenv.Load("./.env")

	client = adyen.NewClient(&common.Config{
		ApiKey:      os.Getenv("ADYEN_API_KEY"),
		Environment: common.TestEnv,
	})

	port = os.Getenv("PORT")

	if port == "" {
		port = "8080" // default when missing
	}

	merchantAccount = os.Getenv("ADYEN_MERCHANT_ACCOUNT")
	clientKey = os.Getenv("ADYEN_CLIENT_KEY")

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
	log.Printf("Running on http://localhost:" + port)
	router.Run(":" + port)
}
