package interface_example

import (
	"fmt"
	"os"
)

type ConsoleWriter struct{}

func (cw ConsoleWriter) Write(p []byte) (int, error) {
	fmt.Println("Writing to console:", string(p))
	return os.Stdout.Write(p)
}