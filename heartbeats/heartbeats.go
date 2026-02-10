package main

import (
	"fmt"
	"time"
)

func main() {
	doWork := func(
		done <-chan interface{},
		pulseInterval time.Duration,
	) (<-chan interface{}, <-chan time.Time) {
		heartbeat := make(chan any)
		results := make(chan time.Time)
		go func() {
			defer close(heartbeat)
			defer close(results)

			pulse := time.Tick(pulseInterval)
			workGen := time.Tick(pulseInterval * 2)

			sendPulse := func() {
				select {
				case heartbeat <- struct{}{}:
				default:
				}
			}
			sendResult := func(r time.Time) {
				for {
					select {
					case <-done:
						return
					case <-pulse:
						sendPulse()
					case results <- r:
						return
					}
				}
			}
			for {
				select {
				case <-done:
					return
				case <-pulse:
					sendPulse()
				case r := <-workGen:
					sendResult(r)
				}
			}
		}()
		return heartbeat, results
	}
	done := make(chan any)
	heartbeat, results := doWork(done, time.Second)
	const timeout = 2 * time.Second
	time.AfterFunc(10*time.Second, func() { close(done) })
	for {
		select {
		case _, ok := <-heartbeat:
			if ok == false {
				return
			}
			fmt.Println("Heartbeat!")
		case res, ok := <-results:
			if ok == false {
				return
			}
			fmt.Println(res)
		case <-time.After(timeout):
			fmt.Println("worker isnt healthy")
		}
	}

}
