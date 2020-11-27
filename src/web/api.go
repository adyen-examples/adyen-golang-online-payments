package web

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/adyen/adyen-go-api-library/v3/src/checkout"
	"github.com/adyen/adyen-go-api-library/v3/src/common"
	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
)

// A temporary store to keep payment data to be sent in additional payment details and redirects.
// This is more secure than a cookie. In a real application this should be in a database.
var paymentDataStore = map[string]string{}

// PaymentMethodsHandler retrieves a list of available payment methods from Adyen API
func PaymentMethodsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.PaymentMethodsRequest

	if err := c.BindJSON(&req); err != nil {
		handleError("PaymentMethodsHandler", c, err, nil)
		return
	}
	req.MerchantAccount = merchantAccount
	req.Channel = "Web"
	log.Printf("Request for %s API::\n%+v\n", "PaymentMethods", req)
	res, httpRes, err := client.Checkout.PaymentMethods(&req)
	if err != nil {
		handleError("PaymentMethodsHandler", c, err, httpRes)
		return
	}
	c.JSON(http.StatusOK, res)
	return
}

// PaymentsHandler makes payment using Adyen API
func PaymentsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.PaymentRequest

	if err := c.BindJSON(&req); err != nil {
		handleError("PaymentsHandler", c, err, nil)
		return
	}
	req.MerchantAccount = merchantAccount // required
	pmType := req.PaymentMethod["type"].(string)
	req.Amount = checkout.Amount{
		Currency: findCurrency(pmType),
		Value:    1000, // value is 10€ in minor units
	}
	orderRef := uuid.Must(uuid.NewRandom())
	req.Reference = orderRef.String() // required
	req.Channel = "Web"               // required
	req.AdditionalData = map[string]interface{}{
		// required for 3ds2 native flow
		"allow3DS2": true,
	}
	req.Origin = "http://localhost:3000" // required for 3ds2 native flow
	req.ShopperIP = c.ClientIP()         // required by some issuers for 3ds2

	// we pass the orderRef in return URL to get paymentData during redirects
	req.ReturnUrl = fmt.Sprintf("http://localhost:3000/api/handleShopperRedirect?orderRef=%s", orderRef) // required for 3ds2 redirect flow
	// Required for Klarna:
	if strings.Contains(pmType, "klarna") {
		req.CountryCode = "DE"
		req.ShopperReference = "12345"
		req.ShopperEmail = "youremail@email.com"
		req.ShopperLocale = "en_US"
		req.LineItems = &[]checkout.LineItem{
			{
				Quantity:           1,
				AmountExcludingTax: 331,
				TaxPercentage:      2100,
				Description:        "Sunglasses",
				Id:                 "Item 1",
				TaxAmount:          69,
				AmountIncludingTax: 400,
			},
			{
				Quantity:           1,
				AmountExcludingTax: 248,
				TaxPercentage:      2100,
				Description:        "Headphones",
				Id:                 "Item 2",
				TaxAmount:          52,
				AmountIncludingTax: 300,
			},
		}
	}

	log.Printf("Request for %s API::\n%+v\n", "Payments", req)
	res, httpRes, err := client.Checkout.Payments(&req)
	log.Printf("Response for %s API::\n%+v\n", "Payments", res)
	log.Printf("HTTP Response for %s API::\n%+v\n", "Payments", httpRes)
	if err != nil {
		handleError("PaymentsHandler", c, err, httpRes)
		return
	}
	if res.Action != nil && res.Action.PaymentData != "" {
		log.Printf("Setting payment data cache for %s\n", orderRef)
		paymentDataStore[orderRef.String()] = res.Action.PaymentData
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusOK, map[string]string{
			"pspReference":  res.PspReference,
			"resultCode":    res.ResultCode.String(),
			"refusalReason": res.RefusalReason,
		})
	}
	return
}

// PaymentDetailsHandler gets payment details using Adyen API
func PaymentDetailsHandler(c *gin.Context) {
	c.Header("Content-Type", "application/json")
	var req checkout.DetailsRequest

	if err := c.BindJSON(&req); err != nil {
		handleError("PaymentDetailsHandler", c, err, nil)
		return
	}
	log.Printf("Request for %s API::\n%+v\n", "PaymentDetails", req)
	res, httpRes, err := client.Checkout.PaymentsDetails(&req)
	log.Printf("HTTP Response for %s API::\n%+v\n", "PaymentDetails", httpRes)
	if err != nil {
		handleError("PaymentDetailsHandler", c, err, httpRes)
		return
	}
	if res.Action != nil {
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusOK, map[string]string{
			"pspReference":  res.PspReference,
			"resultCode":    res.ResultCode.String(),
			"refusalReason": res.RefusalReason,
		})
	}
	return
}

type Redirect struct {
	MD             string
	PaRes          string
	Payload        string `form:"payload"`
	RedirectResult string `form:"redirectResult"`
}

// RedirectHandler handles POST and GET redirects from Adyen API
func RedirectHandler(c *gin.Context) {
	var redirect Redirect
	log.Println("Redirect received")

	orderRef := c.Query("orderRef")
	paymentData := paymentDataStore[orderRef]
	log.Printf("cached paymentData for %s: %s", orderRef, paymentData)

	if err := c.ShouldBind(&redirect); err != nil {
		handleError("RedirectHandler", c, err, nil)
		return
	}

	var details map[string]interface{}
	if redirect.Payload != "" {
		details = map[string]interface{}{
			"payload": redirect.Payload,
		}
	} else if redirect.RedirectResult != "" {
		details = map[string]interface{}{
			"redirectResult": redirect.RedirectResult,
		}
	} else {
		details = map[string]interface{}{
			"MD":    redirect.MD,
			"PaRes": redirect.PaRes,
		}
	}

	req := checkout.DetailsRequest{Details: details, PaymentData: paymentData}

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
		switch res.ResultCode {
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

func handleError(method string, c *gin.Context, err error, httpRes *http.Response) {
	log.Printf("Error in %s: %s\n", method, err.Error())
	if httpRes != nil && httpRes.StatusCode >= 300 {
		c.JSON(httpRes.StatusCode, err.Error())
		return
	}
	c.JSON(http.StatusBadRequest, err.Error())
}
