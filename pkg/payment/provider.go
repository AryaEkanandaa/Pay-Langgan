package payment

type ChargeRequest struct {
	OrderID     string
	GrossAmount float64
	CustomerName string
	CustomerEmail string
}

type ChargeResponse struct {
	ProviderTransactionID string
	Status                string
	Message               string
}

type Provider interface {
	Charge(req ChargeRequest) (*ChargeResponse, error)
}
