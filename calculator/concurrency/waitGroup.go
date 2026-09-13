package concurrency

import (
	"fmt"
	"sync"
	"time"
)

var wg sync.WaitGroup


// with wait group example
func workerSpawanedByGoroutine(id int, wg *sync.WaitGroup) {
	defer wg.Done()
	time.Sleep(1 * time.Second)
	fmt.Printf("completing the worker %d\n",id)
}

func WaitGroupExample() {
	t := time.Now();
	for i := 1; i <= 5; i++ {
		wg.Add(1)
		go workerSpawanedByGoroutine(i, &wg)
	}
	wg.Wait()
	fmt.Printf("time took: %v\n", time.Since(t).Milliseconds())
}


// without wait group example
func workerSpawanedWithoutByGoroutine(id int) {
	time.Sleep(1 * time.Second)
	fmt.Printf("completing the worker %d\n",id)
}

func WithoutWaitGroupExample() {
	t := time.Now()
	for i := 1; i <= 5; i++ {
		workerSpawanedWithoutByGoroutine(i)
	}
	fmt.Printf("time took: %v\n", time.Since(t).Milliseconds())
}