package web

import (
	"context"
	"github.com/adyen/adyen-go-api-library/v7/src/balanceplatform"
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

	req := issuing.PaymentInstrumentsApi.CreatePaymentInstrumentConfig(context.Background()).PaymentInstrumentInfo(body)

	log.Printf("Request for %s API::\n%+v\n", "CreatePaymentInstrument", req)
	res, httpRes, err := issuing.PaymentInstrumentsApi.CreatePaymentInstrument(req)
	log.Printf("Response for %s API::\n%+v\n", "CreatePaymentInstrument", res)

	if err != nil {
		log.Printf("Error in %s: %s\n", "CreatePaymentInstrument", err)
		log.Printf("Error in %s: %s\n", "CreatePaymentInstrument", httpRes.Body)
	}

	// calling
	// func (a *PaymentInstrumentsApi) CreatePaymentInstrument(r CreatePaymentInstrumentConfig) (PaymentInstrument, *_nethttp.Response, error) {
	// should be
	// func (a *PaymentInstrumentsApi) CreatePaymentInstrument(r CreatePaymentInstrumentRequest) (PaymentInstrumentResponse, *_nethttp.Response, error) {

}
