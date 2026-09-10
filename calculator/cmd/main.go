package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/calc"
	errorExample "github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/error"
	interfaceExample "github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/interface_example"
	structscomposition1 "github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/structs_composition_1"
	mylogger "github.com/mohammadkhizerkhan/my-logger"
)

func main() {
	mylogger.Info("starting calculator")
	fmt.Println("2 + 3 =", calc.Add(2, 3))
	fmt.Println(uuid.NewV7())
	structscomposition1.TestNestedStructs()
	interfaceExample.TestInterfaceExamples()
	errorExample.WrapErrorExample()
	userName, err := errorExample.GetUserNameExample(1)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("User Name:", userName)
	}
	errorExample.CheckIfBusinessValidationError(&errorExample.BusinessValidationError{Message: "Test error", Code: 1002})
	errorExample.TestPanic()
}
