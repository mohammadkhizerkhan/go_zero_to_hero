package concurrency

import (
	"fmt"
	"sync"
)

func SimulateChannelDataPassing(){
	wg := sync.WaitGroup{}
	ch := make(chan int)
	wg.Add(2);
	var data int;
	go func(){
		defer wg.Done()
		ch <- 69;
	}();

	go func(){
		defer wg.Done()
		data = <- ch
	}();

	wg.Wait();
	fmt.Println("the value of data is: ",data)
}


// same simulation using different workers

func producer(wg *sync.WaitGroup, ch chan<- int) {
   defer wg.Done()
   
   ch <- 25
}

func consumer(wg *sync.WaitGroup, ch <-chan int) {
   defer wg.Done()

   // Receive data from channel
   data := <- ch
   fmt.Println("Received:", data)
}



func SimulateChannelDataPassingUsingWorkers() {
   ch := make(chan int)
   var wg sync.WaitGroup
   wg.Add(2)

   go producer(&wg, ch)
   go consumer(&wg, ch)

   wg.Wait()
}



// the below example fail, explaination is below
func unbufferedWorker(wg *sync.WaitGroup, ch chan int) {
   defer wg.Done()
   
   ch <- 25
   
   data := <- ch
   fmt.Println("Received:", data)
}


func SimulateChannelDataPassingInSameGoroutine() {
   ch := make(chan int)
   var wg sync.WaitGroup
   wg.Add(1)
   go unbufferedWorker(&wg, ch)
   wg.Wait()
}

// `SimulateChannelDataPassingInSameGoroutine` deadlocks because `unbufferedWorker` uses the same goroutine to both send and receive from an unbuffered channel:

// ```go
// ch <- 25
// data := <-ch
// ```

// An unbuffered channel has no storage. The send can complete only when another goroutine is ready to receive the value.

// Execution:

// 1. `main` starts `worker` and calls `wg.Wait()`.
// 2. `worker` reaches `ch <- 25`.
// 3. Because no other goroutine is receiving, the send blocks.
// 4. `worker` never reaches `data := <-ch`.
// 5. `main` remains blocked in `wg.Wait()`.
// 6. `worker` cannot call `wg.Done()`, so the program deadlocks.

// The receive cannot happen because the goroutine is stuck at the previous send. `WaitGroup` only waits for completion; it does not allow the blocked send to proceed.

// Your first function works because two separate goroutines perform the send and receive:

// ```go
// go func() {
// 	ch <- 69
// }()

// go func() {
// 	data = <-ch
// }()
// ```

// For the `worker` version, use a buffered channel:

// ```go
// ch := make(chan int, 1)
// ```

// The buffer stores `25`, allowing the same goroutine to continue to the receive:

// ```go
// func worker(wg *sync.WaitGroup, ch chan int) {
// 	defer wg.Done()

// 	ch <- 25
// 	data := <-ch

// 	fmt.Println("Received:", data)
// }
// ```

// Alternatively, keep the channel unbuffered but use separate goroutines for sending and receiving.