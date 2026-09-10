package error

import "fmt"

func TestPanic(){
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
		}
	}()
	panic("this is a test panic")
}