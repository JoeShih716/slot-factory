package mock

import (
	"github.com/joe_shih/slot-factory/internal/application/wallet"
	"github.com/shopspring/decimal"
)

type MockPayment struct {
	fakeUserBalanceList map[string]decimal.Decimal
}

var _ wallet.Payment = (*MockPayment)(nil)

func NewPayment() *MockPayment {
	return &MockPayment{
		fakeUserBalanceList: make(map[string]decimal.Decimal),
	}
}

func (p *MockPayment) GetBalance(playerID string) (decimal.Decimal, *wallet.PaymentError) {
	balance, ok := p.fakeUserBalanceList[playerID]
	if !ok {
		balance = decimal.NewFromInt(100000)
		p.fakeUserBalanceList[playerID] = balance
	}
	return balance, nil
}

func (p *MockPayment) Debit(playerID string, amount decimal.Decimal) (decimal.Decimal, *wallet.PaymentError) {
	balance, _ := p.GetBalance(playerID)
	if balance.LessThan(amount) {
		return balance, &wallet.PaymentError{
			Code:    400,
			Message: "balance is not enough",
		}
	}
	balance = balance.Sub(amount)
	return balance, nil
}

func (p *MockPayment) Credit(playerID string, amount decimal.Decimal) (decimal.Decimal, *wallet.PaymentError) {
	balance, _ := p.GetBalance(playerID)
	balance = balance.Add(amount)
	p.fakeUserBalanceList[playerID] = balance
	return balance, nil
}

func (p *MockPayment) DebitAndCredit(playerID string, debitAmount decimal.Decimal, creditAmount decimal.Decimal) (decimal.Decimal, *wallet.PaymentError) {
	balance, _ := p.GetBalance(playerID)
	if balance.LessThan(debitAmount) {
		return balance, &wallet.PaymentError{
			Code:    400,
			Message: "balance is not enough",
		}
	}
	balance = balance.Sub(debitAmount).Add(creditAmount)
	p.fakeUserBalanceList[playerID] = balance
	return balance, nil
}
