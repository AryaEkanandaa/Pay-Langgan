package payment

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

type MidtransProvider struct {
	ServerKey string
	ClientKey string
	BaseURL   string
	Mock      bool
}

func NewMidtransProvider(serverKey, clientKey, baseURL string, mock bool) *MidtransProvider {
	return &MidtransProvider{
		ServerKey: serverKey,
		ClientKey: clientKey,
		BaseURL:   baseURL,
		Mock:      mock,
	}
}

func (p *MidtransProvider) Charge(req ChargeRequest) (*ChargeResponse, error) {
	if p.Mock {
		return p.mockCharge(req)
	}
	return nil, fmt.Errorf("real midtrans not implemented")
}

func (p *MidtransProvider) mockCharge(req ChargeRequest) (*ChargeResponse, error) {
	if req.GrossAmount <= 0 {
		return &ChargeResponse{
			Status:  "failed",
			Message: "invalid amount",
		}, nil
	}

	txID := "mock_midtrans_" + randomHex(16)

	return &ChargeResponse{
		ProviderTransactionID: txID,
		Status:                "success",
		Message:               "Mock charge success",
	}, nil
}

func randomHex(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return hex.EncodeToString(b)
}
