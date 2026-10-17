package payment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"pay-langgan/internal/models"
	paymentrepo "pay-langgan/internal/repositories/payment"
	paymentservice "pay-langgan/internal/services/payment"
	"pay-langgan/internal/utils"
	provider "pay-langgan/pkg/payment"
)

type coreCardStub struct {
	charges int
	result  *provider.MidtransCardTransaction
	err     error
}

func (p *coreCardStub) ChargeCard(_ context.Context, request provider.MidtransCardRequest) (*provider.MidtransCardTransaction, error) {
	p.charges++
	if request.TokenID != "token-only" || request.Amount != 150000 {
		return nil, errors.New("unexpected charge")
	}
	p.result.OrderID = request.OrderID
	return p.result, p.err
}
func (p *coreCardStub) GetCardTransaction(context.Context, string) (*provider.MidtransCardTransaction, error) {
	return p.result, nil
}

func TestMidtransCoreCheckoutPostgres(t *testing.T) {
	for _, scenario := range []string{"pending-then-paid", "response-lost", "wrong-amount"} {
		t.Run(scenario, func(t *testing.T) {
			db := paymentTestDB(t)
			xendit, err := provider.NewXenditProvider("", "unused", "https://api.xendit.co", true)
			if err != nil {
				t.Fatal(err)
			}
			service, err := paymentservice.NewCheckoutService(db, paymentrepo.NewCheckoutRepository(db), xendit, "http://localhost:5173", "https://paylanggan.example", time.Hour)
			if err != nil {
				t.Fatal(err)
			}
			gateway := &paymentservice.BusinessGatewayRuntime{BusinessID: "test", Config: paymentservice.BusinessGatewayInput{Provider: "midtrans", Environment: "sandbox", MidtransMerchantID: "merchant-test", MidtransClientKey: "public-key"}}
			service.SetBusinessGatewayResolver(midtransTestResolver{gateway})
			stub := &coreCardStub{result: &provider.MidtransCardTransaction{StatusCode: "201", TransactionID: "card-transaction", GrossAmount: "150000.00", Currency: "IDR", PaymentType: "credit_card", MerchantID: "merchant-test", TransactionStatus: "pending", RedirectURL: "https://api.sandbox.midtrans.com/v2/token/authentication"}}
			service.SetMidtransCardProviderFactory(func(*paymentservice.BusinessGatewayRuntime) paymentservice.MidtransCardProvider { return stub })
			ctx := context.Background()
			link, err := service.CreateLink(ctx, 1, "test")
			if err != nil {
				t.Fatal(err)
			}
			contact := models.StartCheckoutRequest{Name: "Alya", Email: "alya@example.com", Mobile: "+628123456789"}
			session, err := service.Start(ctx, link.Token, contact)
			if err != nil || session.Mode != "CORE" || session.ClientKey != "public-key" || stub.charges != 0 {
				t.Fatalf("incorrect preparation: %+v %v", session, err)
			}
			if scenario == "response-lost" {
				stub.err = errors.New("connection lost")
			}
			if scenario == "wrong-amount" {
				stub.result.GrossAmount = "1.00"
			}
			request := models.MidtransCardCheckoutRequest{TokenID: "token-only", Name: contact.Name, Email: contact.Email, Mobile: contact.Mobile}
			response, err := service.ChargeMidtransCard(ctx, link.Token, request)
			if scenario == "pending-then-paid" {
				if err != nil || response.AuthenticationURL == "" {
					t.Fatalf("charge: %+v %v", response, err)
				}
				resumed, err := service.Start(ctx, link.Token, models.StartCheckoutRequest{})
				if err != nil || resumed.AuthenticationURL != response.AuthenticationURL {
					t.Fatalf("cannot resume authentication: %+v %v", resumed, err)
				}
			} else if err == nil {
				t.Fatal("uncertain or mismatched response accepted")
			}
			if _, err := service.ChargeMidtransCard(ctx, link.Token, request); !errors.Is(err, utils.ErrConflict) {
				t.Fatalf("duplicate charge allowed: %v", err)
			}
			if stub.charges != 1 {
				t.Fatalf("charged %d times", stub.charges)
			}
			var status string
			if err := db.Get(&status, `SELECT status FROM invoices WHERE id=1`); err != nil || status != "pending" {
				t.Fatal("unverified payment marked paid", status, err)
			}
			if scenario == "wrong-amount" {
				if _, err := service.VerifyMidtransCard(ctx, link.Token); err == nil {
					t.Fatal("verification accepted a mismatched amount")
				}
			}
			stub.result.GrossAmount = "150000.00"
			stub.result.StatusCode = "200"
			stub.result.TransactionStatus = "capture"
			stub.result.FraudStatus = "accept"
			stub.result.RedirectURL = ""
			verified, err := service.VerifyMidtransCard(ctx, link.Token)
			if err != nil || verified.InvoiceStatus != "paid" {
				t.Fatal("immediate verification failed", err)
			}
			if _, err := service.VerifyMidtransCard(ctx, link.Token); err != nil {
				t.Fatal("repeat verification failed", err)
			}
			if stub.charges != 1 {
				t.Fatal("verification submitted another charge")
			}
			if err := db.Get(&status, `SELECT status FROM invoices WHERE id=1`); err != nil || status != "paid" {
				t.Fatal("verified payment not applied", status, err)
			}
			var count int
			if err := db.Get(&count, `SELECT count(*) FROM payments WHERE status='success'`); err != nil || count != 1 {
				t.Fatalf("payments: %d %v", count, err)
			}
		})
	}
}
