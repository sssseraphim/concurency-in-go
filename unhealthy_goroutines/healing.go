package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func or(chans ...<-chan any) <-chan any {
	switch len(chans) {
	case 0:
		return nil
	case 1:
		return chans[0]
	}
	orDone := make(chan any)
	go func() {
		defer close(orDone)
		switch len(chans) {
		case 2:
			select {
			case <-chans[0]:
			case <-chans[1]:
			}
		default:
			select {
			case <-chans[0]:
			case <-chans[1]:
			case <-chans[2]:
			case <-or(append(chans[3:], orDone)...):
			}
		}
	}()
	return orDone
}

func take(
	done <-chan any,
	valueStream <-chan any,
	num int,
) <-chan any {
	resStream := make(chan any)
	go func() {
		defer close(resStream)
		for range num {
			select {
			case <-done:
				return
			case resStream <- <-valueStream:
			}
		}
	}()
	return resStream
}

func orDone(
	done <-chan any,
	c <-chan any,
) <-chan any {
	valStream := make(chan any)
	go func() {
		defer close(valStream)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case <-done:
				case valStream <- v:
				}
			}
		}
	}()
	return valStream
}
func bridge(
	done <-chan any,
	chanStream <-chan <-chan any,
) <-chan any {
	res := make(chan any)
	go func() {
		defer close(res)
		for {
			var stream <-chan any
			select {
			case maybeSream, ok := <-chanStream:
				if !ok {
					return
				}
				stream = maybeSream
			case <-done:
				return
			}
			for val := range orDone(done, stream) {
				select {
				case <-done:
				case res <- val:
				}
			}
		}
	}()
	return res
}

type StartGoroutineFn func(
	done <-chan any,
	pulseInterval time.Duration,
) (heartbeat <-chan any)

func main() {
	newSteward := func(
		timeout time.Duration,
		startGouroutine StartGoroutineFn,
	) StartGoroutineFn {
		return func(
			done <-chan any,
			pulseInterval time.Duration,
		) <-chan any {
			heartbeat := make(chan any)
			go func() {
				defer close(heartbeat)

				var wardDone chan any
				var wardHeartbeat <-chan any
				startWard := func() {
					wardDone = make(chan any)
					wardHeartbeat = startGouroutine(or(wardDone, done), timeout/2)
				}
				startWard()
				pulse := time.Tick(pulseInterval)

			monitorLoop:
				for {
					timeoutSignal := time.After(timeout)

					for {
						select {
						case <-pulse:
							select {
							case heartbeat <- struct{}{}:
							default:
							}
						case <-wardHeartbeat:

							continue monitorLoop
						case <-timeoutSignal:
							log.Println("steward: ward unhealthy; restarting")
							close(wardDone)
							startWard()
							continue monitorLoop
						case <-done:
							return
						}
					}
				}
			}()
			return heartbeat
		}
	}

	doWorkFn := func(
		done <-chan any,
		intList ...int,
	) (StartGoroutineFn, <-chan any) {
		intChanStream := make(chan (<-chan any))
		intStream := bridge(done, intChanStream)
		doWork := func(
			done <-chan any,
			pulseInterval time.Duration,
		) <-chan any {
			intStream := make(chan any)
			heartbeat := make(chan any)
			go func() {
				defer close(intStream)
				select {
				case intChanStream <- intStream:
				case <-done:
					return
				}

				pulse := time.Tick(pulseInterval)

				for {
				valueLoop:
					for _, intVal := range intList {
						if intVal < 0 {
							return
						}

						for {
							select {
							case <-pulse:
								select {
								case heartbeat <- struct{}{}:
								default:
								}
							case intStream <- intVal:
								continue valueLoop
							case <-done:
								return
							}
						}
					}
				}
			}()
			return heartbeat
		}
		return doWork, intStream
	}

	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	done := make(chan any)
	defer close(done)
	doWork, intStream := doWorkFn(done, 1, 2, -1, 3, 4, 5)
	doWorkWithSteward := newSteward(1*time.Millisecond, doWork)
	doWorkWithSteward(done, 1*time.Hour)

	for intVal := range take(done, intStream, 6) {
		fmt.Printf("Received: %v\n", intVal)
	}

}
