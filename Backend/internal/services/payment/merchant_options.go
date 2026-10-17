package payment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"pay-langgan/internal/models"
	billing "pay-langgan/internal/services/billing"
	"pay-langgan/internal/utils"
)

type CheckoutAddOn struct {
	AddOnID  int `json:"add_on_id"`
	Quantity int `json:"quantity"`
}

type CheckoutAddOnOption struct {
	ID    int     `db:"id" json:"id"`
	Name  string  `db:"name" json:"name"`
	Price float64 `db:"price" json:"price"`
}

type checkoutCatalogPlan struct {
	ProductID    int     `db:"product_id"`
	Name         string  `db:"name"`
	Price        float64 `db:"price"`
	BillingCycle string  `db:"billing_cycle"`
	TrialDays    int     `db:"trial_days"`
}

func validateCheckoutOptions(addOns []CheckoutAddOn, coupon string) error {
	if len(addOns) > 20 || len(coupon) > 50 {
		return fmt.Errorf("%w: terlalu banyak add-on atau kode kupon terlalu panjang", utils.ErrBadRequest)
	}
	seen := map[int]bool{}
	for _, item := range addOns {
		if item.AddOnID < 1 || item.Quantity < 1 || item.Quantity > 100 || seen[item.AddOnID] {
			return fmt.Errorf("%w: pilihan atau jumlah add-on tidak valid", utils.ErrBadRequest)
		}
		seen[item.AddOnID] = true
	}
	return nil
}

func moneyCents(value float64) (int64, error) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 || value > 200000000 {
		return 0, fmt.Errorf("%w: harga tidak didukung", utils.ErrBadRequest)
	}
	return int64(math.Round(value * 100)), nil
}

// quoteCheckout reads prices from the tenant's catalog, never from browser amounts.
// A transaction is used during creation so coupon quota and invoice are committed together.
func quoteCheckout(ctx context.Context, db sqlx.ExtContext, business string, plan checkoutCatalogPlan, options []CheckoutAddOn, coupon string, lock bool) ([]models.SubscriptionPreviewItem, float64, int, error) {
	if err := validateCheckoutOptions(options, coupon); err != nil {
		return nil, 0, 0, err
	}
	cents, err := moneyCents(plan.Price)
	if err != nil {
		return nil, 0, 0, err
	}
	charges := []billing.QuoteCharge{{Kind: "plan", Name: plan.Name, Quantity: 1, Cents: cents, Cycle: plan.BillingCycle}}
	for _, item := range options {
		var addon CheckoutAddOnOption
		query := `SELECT a.id,a.name,a.price FROM add_ons a JOIN products pr ON pr.id=a.product_id JOIN services sv ON sv.id=pr.service_id WHERE a.id=$1 AND a.product_id=$2 AND a.billing_cycle=$3 AND sv.business_id=$4 AND pr.status='active' AND a.visible_in_checkout=TRUE`
		if lock {
			query += " FOR SHARE OF a"
		}
		err := sqlx.GetContext(ctx, db, &addon, query, item.AddOnID, plan.ProductID, plan.BillingCycle, business)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, 0, fmt.Errorf("%w: add-on tidak tersedia untuk paket ini", utils.ErrBadRequest)
		}
		if err != nil {
			return nil, 0, 0, err
		}
		cents, err = moneyCents(addon.Price)
		if err != nil {
			return nil, 0, 0, err
		}
		charges = append(charges, billing.QuoteCharge{Kind: "add_on", Name: addon.Name, Quantity: int64(item.Quantity), Cents: cents, Cycle: plan.BillingCycle})
	}
	couponID := 0
	if strings.TrimSpace(coupon) != "" {
		var c struct {
			ID      int        `db:"id"`
			Code    string     `db:"code"`
			Kind    string     `db:"discount_type"`
			Value   float64    `db:"discount_value"`
			Max     *int       `db:"max_usage"`
			Used    int        `db:"used_count"`
			Expires *time.Time `db:"expires_at"`
		}
		query := `SELECT id,code,discount_type,discount_value,max_usage,used_count,expires_at FROM coupons WHERE business_id=$1 AND code=$2`
		if lock {
			query += " FOR UPDATE"
		}
		err := sqlx.GetContext(ctx, db, &c, query, business, coupon)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, 0, 0, fmt.Errorf("%w: kode kupon tidak ditemukan", utils.ErrBadRequest)
		}
		if err != nil {
			return nil, 0, 0, err
		}
		if c.Expires != nil && !c.Expires.After(time.Now().UTC()) || c.Max != nil && c.Used >= *c.Max {
			return nil, 0, 0, fmt.Errorf("%w: kupon kedaluwarsa atau batas penggunaan telah tercapai", utils.ErrBadRequest)
		}
		cents, err = moneyCents(c.Value)
		if err != nil {
			return nil, 0, 0, err
		}
		charges = append(charges, billing.QuoteCharge{Kind: c.Kind, Name: "Coupon " + c.Code, Quantity: 1, Cents: cents})
		couponID = c.ID
	}
	subtotals, total, err := billing.PriceQuote(charges, plan.BillingCycle)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("%w: rincian harga tidak valid", utils.ErrBadRequest)
	}
	if total != 0 && (total < 500000 || total > 20000000000 || total%100 != 0) {
		return nil, 0, 0, fmt.Errorf("%w: total pembayaran kartu harus Rp 5.000–Rp 200.000.000 dalam rupiah bulat", utils.ErrBadRequest)
	}
	items := make([]models.SubscriptionPreviewItem, 0, len(charges))
	for i, c := range charges {
		kind := c.Kind
		if kind == "percentage" || kind == "fixed" {
			kind = "discount"
		}
		items = append(items, models.SubscriptionPreviewItem{Type: kind, Name: c.Name, Quantity: int(c.Quantity), UnitPrice: float64(c.Cents) / 100, Subtotal: float64(subtotals[i]) / 100})
	}
	return items, float64(total) / 100, couponID, nil
}
