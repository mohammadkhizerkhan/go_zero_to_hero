package interface_example

import (
	payment "github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/interface_example/payment"
)

type Order struct{}

func (o Order) ProcessPayment(amount float64, processor payment.PaymentProcessor) error {
	return processor.ProcessPayment(amount)
}