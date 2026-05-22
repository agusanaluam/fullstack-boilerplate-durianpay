package usecase_test

import (
	"errors"
	"testing"

	"github.com/durianpay/fullstack-boilerplate/internal/entity"
	"github.com/durianpay/fullstack-boilerplate/internal/module/payment/usecase"
)

type mockPaymentRepo struct {
	payments []*entity.Payment
	err      error
	gotStatus, gotSort, gotID string
}

func (m *mockPaymentRepo) ListPayments(status, sort, id string) ([]*entity.Payment, error) {
	m.gotStatus = status
	m.gotSort = sort
	m.gotID = id
	return m.payments, m.err
}

func TestListPayments_All(t *testing.T) {
	repo := &mockPaymentRepo{
		payments: []*entity.Payment{
			{ID: "PAY-001", Merchant: "A", Amount: 100, Status: "completed"},
		},
	}
	uc := usecase.NewPaymentUsecase(repo)
	result, err := uc.ListPayments("", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 payment, got %d", len(result))
	}
}

func TestListPayments_FilterStatus(t *testing.T) {
	repo := &mockPaymentRepo{payments: []*entity.Payment{}}
	uc := usecase.NewPaymentUsecase(repo)
	_, _ = uc.ListPayments("completed", "", "")
	if repo.gotStatus != "completed" {
		t.Errorf("expected status=completed, got %q", repo.gotStatus)
	}
}

func TestListPayments_SortForwarded(t *testing.T) {
	repo := &mockPaymentRepo{payments: []*entity.Payment{}}
	uc := usecase.NewPaymentUsecase(repo)
	_, _ = uc.ListPayments("", "-amount", "")
	if repo.gotSort != "-amount" {
		t.Errorf("expected sort=-amount, got %q", repo.gotSort)
	}
}

func TestListPayments_RepoError(t *testing.T) {
	repo := &mockPaymentRepo{err: errors.New("db down")}
	uc := usecase.NewPaymentUsecase(repo)
	_, err := uc.ListPayments("", "", "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
