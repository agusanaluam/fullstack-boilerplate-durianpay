package handler

import (
	"encoding/json"
	"net/http"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	paymentUsecase "github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
	"github.com/durianpay/fullstack-boilerplate/internal/openapigen"
	"github.com/durianpay/fullstack-boilerplate/internal/transport"
)

type PaymentHandler struct {
	paymentUC paymentUsecase.PaymentUsecase
}

func NewPaymentHandler(paymentUC paymentUsecase.PaymentUsecase) *PaymentHandler {
	return &PaymentHandler{paymentUC: paymentUC}
}

func (h *PaymentHandler) GetDashboardV1Payments(w http.ResponseWriter, r *http.Request, params openapigen.GetDashboardV1PaymentsParams) {
	var status, sort, id string
	if params.Status != nil {
		status = *params.Status
	}
	if params.Sort != nil {
		sort = *params.Sort
	}
	if params.Id != nil {
		id = *params.Id
	}

	payments, err := h.paymentUC.ListPayments(status, sort, id)
	if err != nil {
		transport.WriteError(w, err)
		return
	}

	type response struct {
		Payments any `json:"payments"`
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response{Payments: payments}); err != nil {
		transport.WriteAppError(w, entity.ErrorInternal("encode error"))
	}
}
