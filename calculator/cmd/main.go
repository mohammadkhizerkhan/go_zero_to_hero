package main

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/calc"
	structscomposition1 "github.com/mohammadkhizerkhan/go_zero_to_hero/calculator/structs_composition_1"
	mylogger "github.com/mohammadkhizerkhan/my-logger"
)

func main() {
	mylogger.Info("starting calculator")
	fmt.Println("2 + 3 = ", calc.Add(2, 3))
	fmt.Println(uuid.NewV7())
	structscomposition1.TestNestedStructs()
}
