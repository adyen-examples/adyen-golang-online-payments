package web

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"github.com/adyen/adyen-go-api-library/v6/src/checkout"
	"github.com/adyen/adyen-go-api-library/v6/src/common"
	"github.com/adyen/adyen-go-api-library/v6/src/hmacvalidator"
	"github.com/adyen/adyen-go-api-library/v6/src/notification"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionsHandler r
func SessionsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.CreateCheckoutSessionRequest

	orderRef := uuid.Must(uuid.NewRandom())
	req.Reference = orderRef.String() // required
	req.Amount = checkout.Amount{
		Currency: "EUR",
		Value:    1000, // value is 10€ in minor units
	}
	req.CountryCode = "NL"
	req.MerchantAccount = merchantAccount // required
	req.ShopperIP = c.ClientIP()          // optional but recommended (see https://docs.adyen.com/api-explorer/#/CheckoutService/v69/post/sessions__reqParam_shopperIP)

	// ReturnUrl required for 3ds2 redirect flow
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	req.ReturnUrl = fmt.Sprintf(scheme+"://"+c.Request.Host+"/api/handleShopperRedirect?orderRef=%s", orderRef)

	log.Printf("Request for %s API::\n%+v\n", "SessionsHandler", req)
	res, httpRes, err := client.Checkout.Sessions(&req)
	if err != nil {
		handleError("SessionHandler", c, err, httpRes)
		return
	}
	c.JSON(http.StatusOK, res)
	return
}

// WebhookHandler: process incoming webhook notifications (https://docs.adyen.com/development-resources/webhooks)
func WebhookHandler(c *gin.Context) {
	log.Println("Webhook received")

	// get webhook request body
	body, _ := ioutil.ReadAll(c.Request.Body)

	var notificationService notification.NotificationService
	notificationRequest, err := notificationService.HandleNotificationRequest(string(body))

	if err != nil {
		handleError("WebhookHandler", c, err, nil)
		return
	}

	// process notificationRequestItems
	ret := true
	for _, notification := range notificationRequest.GetNotificationItems() {
		if hmacvalidator.ValidateHmac(*notification, hmacKey) {
			// HMAC signature is valid: process notification
			log.Println("Received webhook PspReference: " + notification.PspReference +
				" EventCode: " + notification.EventCode)
		} else {
			// HMAC signature is invalid: reject notificaiton
			log.Println("HMAC signature is invalid")
			ret = false
			break
		}
	}

	if ret {
		c.String(200, "[accepted]")
	} else {
		c.String(401, "Invalid hmac signature")
	}

}

// RedirectHandler handles POST and GET redirects from Adyen API
func RedirectHandler(c *gin.Context) {
	log.Println("Redirect received")
	var details checkout.PaymentCompletionDetails

	if err := c.ShouldBind(&details); err != nil {
		handleError("RedirectHandler", c, err, nil)
		return
	}

	details.RedirectResult = c.Query("redirectResult")
	details.Payload = c.Query("payload")

	req := checkout.DetailsRequest{Details: details}

	log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
	res, httpRes, err := client.Checkout.PaymentsDetails(&req)
	log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
	if err != nil {
		handleError("RedirectHandler", c, err, httpRes)
		return
	}
	log.Printf("Response for %s API::\n%+v\n", "PaymentDetails", res)

	if res.PspReference != "" {
		var redirectURL string
		// Conditionally handle different result codes for the shopper
		switch *res.ResultCode {
		case common.Authorised:
			redirectURL = "/result/success"
			break
		case common.Pending:
		case common.Received:
			redirectURL = "/result/pending"
			break
		case common.Refused:
			redirectURL = "/result/failed"
			break
		default:
			{
				reason := res.RefusalReason
				if reason == "" {
					reason = res.ResultCode.String()
				}
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

func findCurrency(typ string) string {
	switch typ {
	case "ach":
		return "USD"
	case "wechatpayqr":
	case "alipay":
		return "CNY"
	case "dotpay":
		return "PLN"
	case "boletobancario":
	case "boletobancario_santander":
		return "BRL"
	default:
		return "EUR"
	}
	return ""
}

func getPaymentType(pm interface{}) string {
	switch v := pm.(type) {
	case *checkout.CardDetails:
		return v.Type
	case *checkout.IdealDetails:
		return v.Type
	case *checkout.DotpayDetails:
		return v.Type
	case *checkout.GiropayDetails:
		return v.Type
	case *checkout.AchDetails:
		return v.Type
	case *checkout.KlarnaDetails:
		return v.Type
	case map[string]interface{}:
		return v["type"].(string)
	}
	return ""
}

func handleError(method string, c *gin.Context, err error, httpRes *http.Response) {
	log.Printf("Error in %s: %s\n", method, err.Error())
	if httpRes != nil && httpRes.StatusCode >= 300 {
		c.JSON(httpRes.StatusCode, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, err.Error())
}
