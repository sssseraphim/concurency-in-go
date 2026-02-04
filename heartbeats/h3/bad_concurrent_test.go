package main

import (
	"testing"
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
