package main

import (
	"testing"
	"time"
)

func TestDoWork_GeneratesAllNumbers(t *testing.T) {
	done := make(chan any)
	defer close(done)

	intSlice := []int{1, 2, 3, 4, 5}
	heartbeat, res := DoWork(done, intSlice...)

	<-heartbeat

	i := 0
	for r := range res {
		if intSlice[i] != r {
			t.Errorf("index %v, expected %v, received %v", i, intSlice[i], r)
		}
		i++
	}
}

func TestDoWorkWithPulse_GeneratesAllNumbers(t *testing.T) {
	done := make(chan any)
	defer close(done)

	intSlice := []int{1, 2, 3, 4, 5}
	heartbeat, res := DoWorkWithPulse(done, time.Second, intSlice...)
	<-heartbeat

	i := 0
	for {
		select {
		case r, ok := <-res:
			if !ok {
				return
			} else if r != intSlice[i] {
				t.Errorf("index %v: expected %v, but received %v,",
					i,
					intSlice[i],
					r,
				)
			}
			i++
		case <-heartbeat:
		case <-time.After(time.Second * 2):
			t.Fatal("test timed out")

		}
	}
}
