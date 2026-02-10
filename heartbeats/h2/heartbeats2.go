package main

import (
	"fmt"
	"math/rand"
)

func main() {
	doWork := func(done <-chan any) (<-chan any, <-chan int) {
		heartbeatStream := make(chan any, 1)
		resultStream := make(chan int)
		go func() {
			defer close(heartbeatStream)
			defer close(resultStream)

			for range 10 {
				select {
				case heartbeatStream <- struct{}{}:
				default:
				}

				select {
				case resultStream <- rand.Intn(10):
				case <-done:
					return
				}
			}
		}()
		return heartbeatStream, resultStream
	}
	done := make(chan any)
	defer close(done)
	heartbeat, results := doWork(done)
	for {
		select {
		case _, ok := <-heartbeat:
			if !ok {
				return
			}
			fmt.Println("pulse")
		case r, ok := <-results:
			if !ok {
				return
			}
			fmt.Println(r)
		}
	}

}
