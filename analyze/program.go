package main

import (
	"log"
	"os"
	"runtime/pprof"
	"time"
)

func main() {
	log.SetFlags(log.Ltime | log.LUTC)
	log.SetOutput(os.Stdout)

	go func() {
		goroutines := pprof.Lookup("goroutine")
		for range time.Tick(time.Second) {
			log.Printf("goroutine count: %v", goroutines.Count())
		}
	}()
	var blockForever chan struct{}
	for range 10 {
		go func() { <-blockForever }()
		time.Sleep(500 * time.Millisecond)
	}

}

func newProfIfNotDef(name string) *pprof.Profile {
	prof := pprof.Lookup(name)
	if prof == nil {
		prof = pprof.NewProfile(name)
	}
	return prof
}
