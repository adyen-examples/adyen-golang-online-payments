package web

import (
	"context"
	"github.com/adyen/adyen-go-api-library/v7/src/balanceplatform"
	"github.com/adyen/adyen-go-api-library/v7/src/common"
	"log"
)

func CreateCard() {
	issuing := bankClient.BalancePlatform()

	body := balanceplatform.PaymentInstrumentInfo{
		BalanceAccountId: "myPlatformCode",
		Card: &balanceplatform.CardInfo{
			FormFactor:     "physical",
			CardholderName: "Beppe",
			Brand:          "mc",
			BrandVariant:   "mcdebit",
			Configuration: &balanceplatform.CardConfiguration{
				ConfigurationProfileId: "ABC",
			},
		},
		Type: "card",
	}

	req := issuing.PaymentInstrumentsApi.CreatePaymentInstrumentInput().PaymentInstrumentInfo(body)
	log.Printf("Request for %s API::\n%+v\n", "CreatePaymentInstrument", req)
	res, httpRes, err := issuing.PaymentInstrumentsApi.CreatePaymentInstrument(context.Background(), req)
	log.Printf("Response for %s API::\n%+v\n", "CreatePaymentInstrument", res)

	if err != nil {
		log.Printf("Error in %s: %s\n", "CreatePaymentInstrument", err)
		log.Printf("Error in %s: %s\n", "CreatePaymentInstrument", httpRes.Body)
		errorMessage := err.(common.APIError).Message
		errorCode := err.(common.APIError).Code
		errorType := err.(common.APIError).Type
		log.Printf(errorMessage + " " + errorCode + " " + errorType)
	}

	// calling
	// func (a *PaymentInstrumentsApi) CreatePaymentInstrument(r CreatePaymentInstrumentConfig) (PaymentInstrument, *_nethttp.Response, error) {
	// should be
	// func (a *PaymentInstrumentsApi) CreatePaymentInstrument(r CreatePaymentInstrumentRequest) (PaymentInstrumentResponse, *_nethttp.Response, error) {

}
