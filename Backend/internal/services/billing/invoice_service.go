package billing

import (
	"fmt"

	"pay-langgan/internal/database"
	"pay-langgan/internal/models"
	"pay-langgan/internal/repositories/billing"
	"pay-langgan/internal/utils"
)

type InvoiceService struct {
	db   *database.DB
	repo *billing.InvoiceRepository
}

func NewInvoiceService(db *database.DB, repo *billing.InvoiceRepository) *InvoiceService {
	return &InvoiceService{db: db, repo: repo}
}

func (s *InvoiceService) List(businessID string, page, limit int, status, search string) ([]models.InvoiceListItem, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	return s.repo.FindAllByBusinessID(businessID, page, limit, status, search)
}

func (s *InvoiceService) GetDetail(id int, businessID string) (*models.InvoiceDetailResponse, error) {
	invoice, err := s.repo.FindByIDAndBusinessID(id, businessID)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, utils.ErrNotFound
	}

	items, err := s.repo.FindItems(id)
	if err != nil {
		return nil, err
	}
	if items == nil {
		items = []models.InvoiceItemResponse{}
	}
	payment, err := s.repo.FindLatestPayment(id)
	if err != nil {
		return nil, err
	}

	return &models.InvoiceDetailResponse{InvoiceListItem: *invoice, Items: items, Payment: payment}, nil
}

func (s *InvoiceService) MarkPaid(id int, businessID string) (*models.InvoiceDetailResponse, error) {
	tx, err := s.db.Beginx()
	if err != nil {
		return nil, fmt.Errorf("begin mark invoice paid transaction: %w", err)
	}
	defer tx.Rollback()

	_, updated, err := s.repo.MarkPaid(tx, id, businessID)
	if err != nil {
		return nil, err
	}
	if !updated {
		invoice, findErr := s.repo.FindByIDAndBusinessID(id, businessID)
		if findErr != nil {
			return nil, findErr
		}
		if invoice == nil {
			return nil, utils.ErrNotFound
		}
		return nil, fmt.Errorf("invoice is already paid or cancelled")
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit mark invoice paid transaction: %w", err)
	}

	return s.GetDetail(id, businessID)
}
