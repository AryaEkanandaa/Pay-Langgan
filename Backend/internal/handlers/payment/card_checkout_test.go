package payment

import (
	"github.com/labstack/echo/v4"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCardCheckoutRejectsRawCardDetails(t *testing.T) {
	for _, body := range []string{`{"token_id":"token","card_number":"4811111111111114"}`, `{"token_id":"token","cvv":"123"}`, `{"token_id":"token"} {"token_id":"other"}`} {
		e := echo.New()
		request := httptest.NewRequest(http.MethodPost, "/checkout/token/card-payment", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()
		if err := NewCheckoutHandler(nil).ChargeMidtransCard(e.NewContext(request, response)); err != nil {
			t.Fatal(err)
		}
		if response.Code != http.StatusBadRequest {
			t.Fatalf("raw details accepted: %d", response.Code)
		}
	}
}
