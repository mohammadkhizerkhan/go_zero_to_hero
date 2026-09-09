package interface_example

import (
	"fmt"

	payment "github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/interface_example/payment"
)

func TestInterfaceExamples() {
	circle := Circle{Radius: 5}
	rectangle := Rectangle{Width: 4, Height: 6}
	PrintArea(circle)
	PrintArea(rectangle)

	// writer exmaple
	writer := ConsoleWriter{}
	fmt.Fprintf(writer,"this text is comming with custom writer \n")

	// payment processor example
	mockProcessor := payment.NewMockPaymentProcessor()
	paypalProcessor := payment.NewPaypalPaymentProcessor()

	order := Order{}
	order.ProcessPayment(100.0, mockProcessor)
	order.ProcessPayment(100.0, paypalProcessor)

}