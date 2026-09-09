package interface_example

import "fmt"

type ShapeInterface interface {
	Area() float64
}

func PrintArea(s ShapeInterface) {
	fmt.Println("Area:", s.Area())
}