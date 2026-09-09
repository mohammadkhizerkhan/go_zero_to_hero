package interface_example

import "fmt"

type MockPaymentProcessor struct{}

func (m MockPaymentProcessor) ProcessPayment(amount float64) error {
	// Mock payment processing logic for testing purposes
	fmt.Println("processing payment through mock processor for amount:", amount)
	return nil
}

func NewMockPaymentProcessor() PaymentProcessor {
	return MockPaymentProcessor{}
}