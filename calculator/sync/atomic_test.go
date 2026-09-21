    package sync

    import (
        "sync"
        "sync/atomic"
        "testing"
    )

	//  go test -v -count=1 -run TestAtomicCounter ./sync

    func TestAtomicCounter(t *testing.T) {
        var count int64
        var wg sync.WaitGroup

        for i := 0; i < 1000; i++ {
            wg.Add(1)                                                                                                                           
            go func() {                                                                                                                         
                defer wg.Done()                                                                                                                 
                // Hardware-level atomic addition                                                                                               
                atomic.AddInt64(&count, 1)                                                                                                      
            }()                                                                                                                                 
        }

        wg.Wait()

        finalCount := atomic.LoadInt64(&count)
        t.Logf("Final atomic counter value: %d", finalCount)

        if finalCount != 1000 {
            t.Fatalf("Expected 1000, got %d", finalCount)                                                                                       
        }
    }