package main

import (
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
	log.SetOutput(os.Stdout)
	log.SetFlags(log.Ltime | log.LUTC)

	doWork := func(done <-chan any, _ time.Duration) <-chan any {
		log.Println("ward: hello im irresponsible!")
		go func() {
			<-done
			log.Println("ward: im halting")
		}()
		return nil
	}

	doWorkWithSteward := newSteward(4*time.Second, doWork)

	done := make(chan any)
	time.AfterFunc(9*time.Second, func() {
		log.Println("main: halting steward and ward")
		close(done)
	})

	for range doWorkWithSteward(done, 4*time.Second) {
	}
	log.Println("Done")

}
