package interface_example

type PaymentProcessor interface {
	ProcessPayment(amount float64) error
}
