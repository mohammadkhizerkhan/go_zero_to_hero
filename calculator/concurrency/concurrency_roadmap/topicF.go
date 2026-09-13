package concurrency_roadmap

import (
	"fmt"
	"time"
)

// The Spawner: Launch 10,000 goroutines that print “Done”. Measure execution time.

func Spawner() {
	t := time.Now()
	for i := 0; i < 10000; i++ {
		go func() {
			println("Done")
		}()
	}
	fmt.Println("Time taken to spawn 10,000 goroutines: ", time.Since(t))
}

func SpawnerWithoutGoroutine() {
	t := time.Now()
	for i := 0; i < 10000; i++ {
		println("Done")
	}
	fmt.Println("Time taken to execute 10,000 operations without goroutines: ", time.Since(t))
}


// The Heartbeat: Background goroutine printing “Pulse” every 500ms. Stop it when main exits.

func Heartbeat() {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
				println("Pulse")
				time.Sleep(500 * time.Millisecond)
			}
		}
	}();
	time.Sleep(5*time.Second)
	close(done)
	fmt.Println("Heartbeat stopped.")
}