package web

import (
	"context"
	"fmt"
	"github.com/adyen/adyen-go-api-library/v11/src/webhook"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/adyen/adyen-go-api-library/v11/src/checkout"
	"github.com/adyen/adyen-go-api-library/v11/src/common"
	"github.com/adyen/adyen-go-api-library/v11/src/hmacvalidator"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionsHandler r
func SessionsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")

	service := client.Checkout()

	orderRef := uuid.Must(uuid.NewRandom())

	// ReturnUrl required for 3ds2 redirect flow
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	body := checkout.CreateCheckoutSessionRequest{
		Reference: orderRef.String(),
		Amount: checkout.Amount{
			Value:    10000, // value is 100€ in minor units
			Currency: "EUR",
		},
		CountryCode:     common.PtrString("NL"),
		MerchantAccount: merchantAccount,
		Channel:         common.PtrString("Web"),
		ReturnUrl:       fmt.Sprintf(scheme+"://"+c.Request.Host+"/api/handleShopperRedirect?orderRef=%s", orderRef),
		ShopperIP:       common.PtrString(c.ClientIP()), // optional but recommended (see https://docs.adyen.com/api-explorer/#/CheckoutService/v69/post/sessions__reqParam_shopperIP)
		// set lineItems required for some payment methods (ie Klarna)
		LineItems: []checkout.LineItem{
			{Quantity: common.PtrInt64(1), AmountIncludingTax: common.PtrInt64(5000), Description: common.PtrString("Sunglasses")},
			{Quantity: common.PtrInt64(1), AmountIncludingTax: common.PtrInt64(5000), Description: common.PtrString("Headphones")},
		},
	}

	req := service.PaymentsApi.SessionsInput().CreateCheckoutSessionRequest(body)
	log.Printf("Request for %s API::\n%+v\n", "SessionsHandler", req)
	res, httpRes, err := service.PaymentsApi.Sessions(context.Background(), req)
	log.Printf("Response for %s API::\n%+v\n", "SessionsHandler", res.SessionData)
	log.Printf("Response for %s API::\n%+v\n", "SessionsHandler", res.Id)
	if err != nil {
		handleError("SessionHandler", c, err, httpRes)
		return
	}
	c.JSON(http.StatusOK, res)
	return

}

// WebhookHandler process incoming webhook notifications (https://docs.adyen.com/development-resources/webhooks)
func WebhookHandler(c *gin.Context) {
	log.Printf("Webhook received")

	// get webhook request body
	body, _ := ioutil.ReadAll(c.Request.Body)

	notificationRequest, err := webhook.HandleRequest(string(body))

	if err != nil {
		handleError("WebhookHandler", c, err, nil)
		return
	}

	var ret bool

	// fetch first (and only) NotificationRequestItem
	notification := notificationRequest.GetNotificationItems()[0]

	if hmacvalidator.ValidateHmac(*notification, hmacKey) {
		log.Println("Received webhook PspReference: " + notification.PspReference +
			" EventCode: " + notification.EventCode)

		// consume event asynchronously
		consumeEvent(*notification)

		ret = true
	} else {
		// HMAC signature is invalid
		log.Println("HMAC signature is invalid")
		ret = false
	}

	if ret {
		c.Status(http.StatusAccepted)
	} else {
		c.String(401, "Invalid hmac signature")
	}

}

// process payload asynchronously
func consumeEvent(item webhook.NotificationRequestItem) {

	log.Println("Processing eventCode " + item.EventCode)

	// add item to DB, queue or run in a different thread

}

// RedirectHandler handles POST and GET redirects from Adyen API
func RedirectHandler(c *gin.Context) {
	log.Println("Redirect received")

	service := client.Checkout()

	req := service.PaymentsApi.PaymentsDetailsInput()
	req = req.PaymentDetailsRequest(checkout.PaymentDetailsRequest{
		PaymentData: common.PtrString("1234"),
		Details: checkout.PaymentCompletionDetails{
			RedirectResult: common.PtrString(c.Query("redirectResult")),
			Payload:        common.PtrString(c.Query("payload")),
		},
	})
	log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
	res, httpRes, err := service.PaymentsApi.PaymentsDetails(context.Background(), req)
	log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
	if err != nil {
		handleError("RedirectHandler", c, err, httpRes)
		return
	}
	log.Printf("Response for %s API::\n%+v\n", "PaymentDetails", res)

	if !common.IsNil(*res.PspReference) {
		var redirectURL string
		// Conditionally handle different result codes for the shopper
		switch *res.ResultCode {
		case "Authorised":
			redirectURL = "/result/success"
			break
		case "Pending":
		case "Received":
			redirectURL = "/result/pending"
			break
		case "Refused":
			redirectURL = "/result/failed"
			break
		default:
			{
				reason := *res.RefusalReason
				log.Printf(reason)
				if !common.IsNil(reason) {
					reason = *res.ResultCode
				}
				log.Printf(reason)
				log.Printf("res1" + *res.ResultCode)
				redirectURL = fmt.Sprintf("/result/error?reason=%s", url.QueryEscape(reason))
				break
			}
		}
		c.Redirect(
			http.StatusFound,
			redirectURL,
		)
		return
	}
	c.JSON(httpRes.StatusCode, httpRes.Status)
	return
}

/* Utils */

func handleError(method string, c *gin.Context, err error, httpRes *http.Response) {
	log.Printf("Error in %s: %s\n", method, err.Error())
	if httpRes != nil && httpRes.StatusCode >= 300 {
		c.JSON(httpRes.StatusCode, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, err.Error())
}
