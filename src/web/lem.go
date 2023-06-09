package web

import (
	"context"
	"github.com/adyen/adyen-go-api-library/v7/src/legalentity"
	"log"
)

func CreateTransferInstrument() {
	lem := client.LegalEntity()

	body := legalentity.TransferInstrumentInfo{
		Type:          "bankAccount",
		LegalEntityId: "LE000000001",
		BankAccount: legalentity.BankAccountInfo{
			AccountIdentification: &legalentity.BankAccountInfoAccountIdentification{
				IbanAccountIdentification: &legalentity.IbanAccountIdentification{
					Iban: "NL1234567890",
					Type: "iban",
				},
			},
		},
	}

	req := lem.TransferInstrumentsApi.CreateTransferInstrumentInput().TransferInstrumentInfo(body)

	log.Printf("Request for %s API::\n%+v\n", "CreateTransferInstrument", req)
	res, _, err := lem.TransferInstrumentsApi.CreateTransferInstrument(context.Background(), req)
	log.Printf("Response for %s API::\n%+v\n", "CreateTransferInstrument", res)

	if err != nil {
		log.Printf("Error in %s: %s\n", "CreateTransferInstrument", err.Error())
	}

	// calling
	// func (a *PaymentInstrumentsApi) CreatePaymentInstrument(r CreatePaymentInstrumentConfig) (PaymentInstrument, *_nethttp.Response, error) {
	// should be
	// func (a *PaymentInstrumentsApi) CreatePaymentInstrument(r CreatePaymentInstrumentRequest) (PaymentInstrumentResponse, *_nethttp.Response, error) {

}
