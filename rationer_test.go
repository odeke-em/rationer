package schedSlicer

import (
	"fmt"
	"testing"
	"time"
)

func ranger(n int) Job {
	return func() interface{} {
		tick := time.Tick(1e5)
		start := time.Now()
		for i := 0; i < n; i++ {
			fmt.Println("# i", start, i)
			<-tick
		}
		fmt.Println("\033[47mDone\033[00m", n)
		if n&1 == 0 {
			return true
		}
		return n
	}
}

func TestRationer(t *testing.T) {
	capacity := uint64(10)

	rationer := NewRationer(capacity)

	if rationer == nil {
		t.Errorf("expected non-nil Rationer")
	}

	loader := rationer.Run()

	for i := 0; i < 5; i += 1 {
		fmt.Println("\033[92m# i ", i, "\033[00m")
		loader <- ranger(10)
	}

	loader <- Sentinel
	rationer.Wait()
}
