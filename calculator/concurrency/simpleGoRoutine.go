package concurrency

import (
	"fmt"
	"time"
)

func sayHello() {
	fmt.Println("this is hello from a goroutine")
}
func SimpleGoRoutine() {
	go sayHello();
	fmt.Println("this is hello after spawning goroutine")
}

func SimpleGoRoutineWithTimer() {
	go sayHello();
	fmt.Println("this is hello after spawning goroutine")
	// Wait for a while to let the goroutine finish
	time.Sleep(1 * time.Second)
}