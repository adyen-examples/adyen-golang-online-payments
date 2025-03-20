package web

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func setPageAndData(c *gin.Context, data gin.H) {
	// Call the HTML method of the Context to render a template
	c.HTML(
		// Set the HTTP status to 200 (OK)
		http.StatusOK,
		// Use the index.html template
		"index.html",
		// Pass the data that the page uses
		data,
	)
}

// IndexHandler serves the index.html page
func IndexHandler(c *gin.Context) {
	log.Println("Loading main page")
	setPageAndData(c, gin.H{
		"page": "main",
	})
}

// PreviewHandler serves the preview.html page
func PreviewHandler(c *gin.Context) {
	log.Println("Loading preview page")
	setPageAndData(c, gin.H{
		"page": "preview",
		"type": c.Param("type"),
	})
}

// CheckoutHandler serves the selected page (dropin, card, etc..)
func CheckoutHandler(c *gin.Context) {
	log.Println("Loading page: " + c.Param("type"))
	setPageAndData(c, gin.H{
		"page":      c.Param("type"),
		"clientKey": clientKey,
	})
}

// ResultHandler serves the result.html page
func ResultHandler(c *gin.Context) {
	log.Println("Loading result page")

	status := c.Param("status")
	refusalReason := c.Query("reason")
	var msg, img string
	switch status {
	case "pending":
		msg = "Your order has been received! Payment completion pending."
		img = "success"
		break
	case "failed":
		msg = "The payment was refused. Please try a different payment method or card."
		img = "failed"
		break
	case "error":
		msg = fmt.Sprintf("Error! Reason: %s", refusalReason)
		img = "failed"
		break
	default:
		msg = "Your order has been successfully placed."
		img = "success"
	}
	setPageAndData(c, gin.H{
		"page":   "result",
		"status": status,
		"msg":    msg,
		"img":    img,
	})
}
