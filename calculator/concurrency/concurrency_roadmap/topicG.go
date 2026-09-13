package concurrency_roadmap

import (
	"fmt"
	"sync"
	"time"
)

func player(name string, in <-chan int, out chan<- int, done <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		select {
		case <-done:
			return
		case val, ok := <-in:
			if !ok {
				return
			}

			// Pacing: Wait 500ms before processing and passing
			time.Sleep(500 * time.Millisecond)
			val++
			fmt.Printf("[%s] count: %d\n", name, val)

			select {
			case out <- val:
			case <-done:
				return
			}
		}
	}
}

func TestPingPong() {
	pingCh := make(chan int)
	pongCh := make(chan int)
	done := make(chan struct{})
	var wg sync.WaitGroup

	wg.Add(2)
	// Ping reads from pingCh, writes to pongCh
	go player("ping", pingCh, pongCh, done, &wg)
	// Pong reads from pongCh, writes to pingCh
	go player("pong", pongCh, pingCh, done, &wg)

	// Serve initial ball into pingCh
	pingCh <- 0

	time.Sleep(10 * time.Second)

	// Clean shutdown
	close(done)
	wg.Wait()
}