package main

import "time"

func DoWork(
	done <-chan any,
	nums ...int,
) (<-chan any, <-chan int) {
	heartbeat := make(chan any, 1)
	res := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(res)

		time.Sleep(2 * time.Second)

		for _, n := range nums {
			select {
			case heartbeat <- struct{}{}:
			default:
			}

			select {
			case <-done:
				return
			case res <- n:
			}
		}
	}()
	return heartbeat, res
}
