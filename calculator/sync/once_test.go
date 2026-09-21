package sync

import (
	"sync"
	"testing"
)

//  go test -v -run TestSingleton ./sync
func TestSingleton(t *testing.T) {
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			db := GetDBInstance()
			if db == nil {
				t.Error("Failed to retrieve DB instance")
			}
		}()
	}

	wg.Wait()
}
