package checkoutstate

// InvoiceSQL accepts only a trusted SQL expression from repository code.
// Provider sessions in progress remain in progress until reconciliation confirms
// their outcome; a local timer must never claim a card payment has expired.
func InvoiceSQL(invoiceID string) string {
	return `COALESCE((SELECT CASE
 WHEN cs.status IN ('created','failed') AND cs.expires_at <= (NOW() AT TIME ZONE 'UTC') THEN 'expired'
 ELSE cs.status END FROM checkout_sessions cs WHERE cs.invoice_id=` + invoiceID + `
 ORDER BY CASE WHEN cs.status IN ('creating','active') THEN 0 ELSE 1 END, cs.id DESC LIMIT 1), '')`
}
