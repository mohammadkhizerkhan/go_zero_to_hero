package concurrency

import (
	"fmt"
	"sync"
)

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func startFanOut(jobs <-chan int, results chan<- int, wg *sync.WaitGroup, workerID int) {
	defer wg.Done()
	for job := range jobs {
		if isPrime(job) {
			results <- job
		}
	}
}

func collectResults(results <-chan int) int {
    sum := 0
    for prime := range results {
        sum += prime
    }
    return sum
}

func FanOutFanInExample() {
	wg := sync.WaitGroup{}
	jobs := make(chan int, 100)
	results := make(chan int, 100)

	// Start fan-out workers
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go startFanOut(jobs, results, &wg, i)
	}

	go func() {
        for i := 0; i <= 10; i++ {
            jobs <- i
        }
        close(jobs)
    }()

	go func() {
        wg.Wait()
        close(results)
    }()

	sum := collectResults(results)
	fmt.Println("the total sum of all prime numbers from 0 to 1000: ",sum)
	
}