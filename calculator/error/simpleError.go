package error

import "fmt"

var ErrDivisionByZero = fmt.Errorf("division by zero is not allowed")

func divide(numerator, denominator float64) (float64, error) {
	if denominator == 0 {
		return 0, ErrDivisionByZero
	}
	return numerator / denominator, nil
}
