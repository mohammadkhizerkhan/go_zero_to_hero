package interface_example

import "fmt"

type PaypalPaymentProcessor struct{}

func (p PaypalPaymentProcessor) ProcessPayment(amount float64) error {
	// Implement PayPal payment processing logic here
	fmt.Printf("processing payment through paypal: %.2f\n", amount)
	return nil
}

func NewPaypalPaymentProcessor() PaymentProcessor {
	return PaypalPaymentProcessor{}
}