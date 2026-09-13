package concurrency

import (
	"fmt"
	"sync"
	"time"
)

var counter = 0

func incrementCounter() {
	for i := 0; i < 1000; i++ {
		counter = counter + 1
	}
}

func incrementCounterWithMutex(mutex *sync.Mutex) {
	for i := 0; i < 1000; i++ {
		mutex.Lock()
		counter = counter + 1
		mutex.Unlock()
	}
}

func SimulateRaceCondition() {
	for i := 0; i < 5; i++ {
		go incrementCounter()
	}
	time.Sleep(2 * time.Second)
	fmt.Println("the total value of counter is : ",counter)
}

func SimulateRaceConditionWithMutex() {
	var mutex sync.Mutex
	for i := 0; i < 5; i++ {
		go incrementCounterWithMutex(&mutex)
	}
	time.Sleep(2 * time.Second)
	fmt.Println(" the value of counter with mutex is: ",counter);
}