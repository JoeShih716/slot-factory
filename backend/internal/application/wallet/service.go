package wallet

import (
	"log/slog"

	"github.com/shopspring/decimal"
)

type PaymentError struct {
	Code    int
	Message string
}

// Payment 定義了錢包相關的業務邏輯介面。
type Payment interface {
	// GetBalance 取得玩家餘額。
	GetBalance(playerID string) (balance decimal.Decimal, err *PaymentError)

	// Debit 扣款。
	// 如果餘額不足，返回錯誤。
	Debit(playerID string, amount decimal.Decimal) (newBalance decimal.Decimal, err *PaymentError)

	// Credit 加款。
	Credit(playerID string, amount decimal.Decimal) (newBalance decimal.Decimal, err *PaymentError)

	// DebitAndCredit 扣款和加款。
	DebitAndCredit(playerID string, debitAmount decimal.Decimal, creditAmount decimal.Decimal) (newBalance decimal.Decimal, err *PaymentError)
}

type Service struct {
	payment Payment
	logger  *slog.Logger
}

// NewService 建立一個新的錢包服務實例。
func NewService(logger *slog.Logger, payment Payment) *Service {
	return &Service{
		payment: payment,
		logger:  logger.With("component", "wallet_service"),
	}
}

func (s *Service) GetBalance(playerID string) (decimal.Decimal, *PaymentError) {
	balance, err := s.payment.GetBalance(playerID)
	if err != nil {
		s.logger.Error("get balance failed", "playerID", playerID, "error", err)
		return decimal.Zero, err
	}
	return balance, nil
}

func (s *Service) Debit(playerID string, amount decimal.Decimal) (decimal.Decimal, *PaymentError) {
	newBalance, err := s.payment.Debit(playerID, amount)
	if err != nil {
		s.logger.Error("debit failed", "playerID", playerID, "amount", amount, "error", err)
		return newBalance, err
	}
	return newBalance, nil
}

func (s *Service) Credit(playerID string, amount decimal.Decimal) (decimal.Decimal, *PaymentError) {
	newBalance, err := s.payment.Credit(playerID, amount)
	if err != nil {
		s.logger.Error("credit failed", "playerID", playerID, "amount", amount, "error", err)
		return decimal.Zero, err
	}
	return newBalance, nil
}

func (s *Service) DebitAndCredit(playerID string, debitAmount decimal.Decimal, creditAmount decimal.Decimal) (decimal.Decimal, *PaymentError) {
	newBalance, err := s.payment.DebitAndCredit(playerID, debitAmount, creditAmount)
	if err != nil {
		s.logger.Error("debit and credit failed", "playerID", playerID, "debitAmount", debitAmount, "creditAmount", creditAmount, "error", err)
		return newBalance, err
	}
	return newBalance, nil
}
