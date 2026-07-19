package models

type CreateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Meta        JSONMap `json:"meta"`
}

type UpdateServiceRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Meta        JSONMap `json:"meta"`
}

type CreateProductRequest struct {
	ServiceID   int     `json:"service_id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Meta        JSONMap `json:"meta"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Meta        JSONMap `json:"meta"`
}

type CreatePlanRequest struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	TrialDays    int     `json:"trial_days"`
	Meta         JSONMap `json:"meta"`
}

type UpdatePlanRequest struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	TrialDays    *int    `json:"trial_days"`
	Meta         JSONMap `json:"meta"`
}

type CreateAddOnRequest struct {
	ProductID    int     `json:"product_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	Meta         JSONMap `json:"meta"`
}

type UpdateAddOnRequest struct {
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	BillingCycle string  `json:"billing_cycle"`
	Meta         JSONMap `json:"meta"`
}
