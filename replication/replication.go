package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func main() {
	doWork := func(
		done <-chan any,
		id int,
		wg *sync.WaitGroup,
		result chan<- int,
	) {
		started := time.Now()
		defer wg.Done()
		simulatedTime := time.Duration(rand.Intn(5)) * time.Second
		select {
		case <-done:
		case <-time.After(simulatedTime):
		}
		select {
		case <-done:
		case result <- id:
		}
		took := time.Since(started)
		if took < simulatedTime {
			took = simulatedTime
		}
		fmt.Printf("%v took %v\n", id, took)
	}
	done := make(chan any)
	result := make(chan int)
	var wg sync.WaitGroup
	wg.Add(10)
	for i := range 10 {
		go doWork(done, i, &wg, result)
	}
	firstReturned := <-result
	close(done)
	wg.Wait()

	fmt.Printf("received answer from %v", firstReturned)
}
