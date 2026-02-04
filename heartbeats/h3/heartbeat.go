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

func DoWorkWithPulse(
	done <-chan any,
	pulseInterval time.Duration,
	nums ...int,
) (<-chan any, <-chan int) {
	heartbeat := make(chan any, 1)
	res := make(chan int)
	go func() {
		defer close(heartbeat)
		defer close(res)

		time.Sleep(2 * time.Second)

		pulse := time.Tick(pulseInterval)
	numLoop:
		for _, n := range nums {
			for {
				select {
				case <-done:
					return
				case <-pulse:
					select {
					case heartbeat <- struct{}{}:
					default:
					}
				case res <- n:
					continue numLoop
				}
			}
		}
	}()
	return heartbeat, res
}
