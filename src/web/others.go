package web

import (
	"context"
	"github.com/adyen/adyen-go-api-library/v7/src/checkout"
	"github.com/adyen/adyen-go-api-library/v7/src/common"
	"github.com/google/uuid"
	"log"
)

func PayWithGooglePay() {

	paymentsService := client.Checkout().PaymentsApi

	orderRef := uuid.Must(uuid.NewRandom())

	paymentMethod := checkout.GooglePayDetails{
		GooglePayToken: "gpay-token",
		Type:           common.PtrString("googlepay"),
	}

	body := checkout.PaymentRequest{
		Reference: orderRef.String(),
		Amount: checkout.Amount{
			Value:    10000, // value is 100€ in minor units
			Currency: "EUR",
		},
		MerchantAccount: merchantAccount,
		ReturnUrl:       "http://.....",
		PaymentMethod:   checkout.CheckoutPaymentMethod{GooglePayDetails: &paymentMethod},
	}

	req := paymentsService.PaymentsConfig(context.Background()).PaymentRequest(body)
	// but ideally:
	// req := paymentsService.makePaymentRequest(context.Background()).PaymentRequest(body)

	log.Printf("Request for %s API::\n%+v\n", "Payments", req)
	res, _, err := paymentsService.Payments(req)
	log.Printf("Response for %s API::\n%+v\n", "Payments", res)

	if err != nil {
		log.Printf("Error in %s: %s\n", "Payments", err.Error())
	}

}
