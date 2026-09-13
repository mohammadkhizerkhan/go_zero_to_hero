package concurrency

import (
	"fmt"
	"sync"
	"time"
)

func worker(
    jobs <-chan int,
    res chan<- int,
    id int,
    wg *sync.WaitGroup,
) {
    defer wg.Done()

    for job := range jobs {
        fmt.Printf("Worker %d: started job %d\n", id, job)
        time.Sleep(time.Second)
        res <- job * 2
    }
}

func WorkerPoolExample() []int {
    jobs := make(chan int, 10)
    res := make(chan int, 100)

    var wg sync.WaitGroup
    t:=time.Now();

    // change this to 5, 10, 20, 50, 100 and see the time taken for each
    for i := 0; i < 5; i++ {
        wg.Add(1)
        go worker(jobs, res, i, &wg)
    }

    for i := 0; i < 100; i++ {
        jobs <- i
    }
    close(jobs)

    go func() {
        wg.Wait()
        close(res)
    }()

    results := make([]int, 0, 100)

    for value := range res {
        results = append(results, value)
    }
    fmt .Println("All jobs completed. Results: ", results," time taken: ",time.Since(t))
    return results
}