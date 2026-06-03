# TODO / Roadmap

## Phase 5 - Billing & Payments (Next)

- [ ] Invoice generation (with invoice items from subscription items)
- [ ] Billing cron job (bill due subscriptions)
- [ ] Billing queue worker (process billing queue)
- [ ] Payment mock Midtrans charge
- [ ] Payment attempts tracking
- [ ] Midtrans webhook handler
- [ ] Payment status update from webhook

## Phase 6 - Testing & Polish

- [ ] Unit tests for all repositories
- [ ] Unit tests for all services
- [ ] Unit tests for all handlers
- [ ] Integration tests
- [ ] API load testing

## Improvements

- [ ] Coupon tenant-specific: tambahkan `business_id` ke tabel coupons agar coupon bisa dibuat per-business
- [ ] Soft delete untuk services, products, plans, add-ons
- [ ] Audit log untuk semua operasi CRUD
- [ ] Rate limiting pada API endpoints
- [ ] Request validation menggunakan library validator
- [ ] Caching untuk lookup data
