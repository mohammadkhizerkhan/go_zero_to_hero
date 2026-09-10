package error

import (
	"errors"
	"fmt"
)


type BusinessValidationError struct {
	Message string
	Code   int
}

func (e *BusinessValidationError) Error() string {
	return fmt.Sprintf("BusinessValidationError: %s (Code: %d)", e.Message, e.Code)
}



func DoSomething(age int) (bool, error) {
	if age < 0 {
		// Simulating an error condition
		return false, &BusinessValidationError{Message: "Age cannot be negative", Code: 1002}
	}
	// Simulating an error condition
	return true, nil
}

func CheckIfBusinessValidationError(err error) bool {
	var target *BusinessValidationError
	if errors.As(err,&target){
		fmt.Println("the error belongs to businessValidation error")
		return true;
	}
	return false;
}
