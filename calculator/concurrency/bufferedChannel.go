package concurrency

import (
	"fmt"
	"sync"
	"time"
)

func SimpleBufferedChannelExample() {
	var wg sync.WaitGroup
	wg.Add(1)
	ch := make(chan int, 3)

	ch <- 1
	ch <- 2
	ch <- 3

	data := <-ch
	fmt.Println("Received:", data)
	data = <-ch
	fmt.Println("Received:", data)
	data = <-ch
	fmt.Println("Received:", data)

	wg.Done()
}

// read from multiple channels using select statement


func worker1( ch chan int) {
	for{
		time.Sleep(1 * time.Second);
		ch <- 1;
	}
}

func worker2(ch chan int) {
	for{
		time.Sleep(2 * time.Second)
		ch <- 2
	}
}

func ReadFromMultipleChannelsUsingSelect() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go worker1( ch1)
	go worker2( ch2)

	select {
	case data := <-ch1:
		fmt.Println("Received from channel 1:", data)
	case data := <-ch2:
		fmt.Println("Received from channel 2:", data)
	}
}

func ReadFromMultipleChannelsUsingForever() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	go worker1( ch1)
	go worker2( ch2)

	for{
	select {
	case data := <-ch1:
		fmt.Println("Received from channel 1:", data)
	case data := <-ch2:
		fmt.Println("Received from channel 2:", data)
	}
	}
}