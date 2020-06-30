package web

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/adyen/adyen-go-api-library/v2/src/checkout"
	"github.com/adyen/adyen-go-api-library/v2/src/common"

	"github.com/gin-gonic/gin"
)

const PaymentDataCookie = "paymentData"

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
	pmType := req.PaymentMethod["type"].(string)
	req.Amount = checkout.Amount{
		Currency: findCurrency(pmType),
		Value:    1000, // value is 10€ in minor units
	}
	req.Reference = fmt.Sprintf("%v", time.Now())
	req.Channel = "Web"
	req.AdditionalData = map[string]interface{}{
		"allow3DS2": true,
	}

	req.ReturnUrl = "http://localhost:3000/api/handleShopperRedirect"
	// Required for Klarna:
	if pmType == "klarna" {
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
	req.MerchantAccount = merchantAccount

	log.Printf("Request for %s API::\n%+v\n", "Payments", req)
	res, httpRes, err := client.Checkout.Payments(&req)
	log.Printf("Response for %s API::\n%+v\n", "Payments", res)
	log.Printf("HTTP Response for %s API::\n%+v\n", "Payments", httpRes)
	if err != nil {
		handleError("PaymentsHandler", c, err, httpRes)
		return
	}
	if res.Action != nil && res.Action.PaymentData != "" {
		log.Printf("Setting payment data cookie %s\n", res.Action.PaymentData)
		c.SetCookie(PaymentDataCookie, res.Action.PaymentData, 3600, "", "localhost", false, true)
		c.JSON(http.StatusOK, res)
	} else {
		c.JSON(http.StatusOK, map[string]string{
			"pspReference": res.PspReference,
			"resultCode":   res.ResultCode.String(),
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
			"pspReference": res.PspReference,
			"resultCode":   res.ResultCode.String(),
		})
	}
	return
}

type Redirect struct {
	MD      string
	PaRes   string
	Payload string `form:"payload"`
}

// RedirectHandler handles POST and GET redirects from Adyen API
func RedirectHandler(c *gin.Context) {
	var redirect Redirect
	log.Println("Redirect received")

	if err := c.ShouldBind(&redirect); err != nil {
		handleError("RedirectHandler", c, err, nil)
		return
	}
	paymentData, err := c.Cookie(PaymentDataCookie)
	log.Printf(">>>>>>>>>>>>>>>>>>>>>>>>>>>>Cookie paymentData %s", paymentData)

	if err != nil {
		handleError("RedirectHandler", c, err, nil)
		return
	}
	var details map[string]interface{}
	if redirect.Payload != "" {
		details = map[string]interface{}{
			"payload": redirect.Payload,
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
	c.SetCookie(PaymentDataCookie, "", 3600, "", "localhost", false, true)
	if err != nil {
		handleError("RedirectHandler", c, err, httpRes)
		return
	}
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
