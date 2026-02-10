package main

import (
	"fmt"
	"math/rand"
)

func main() {
	repeat := func(
		done <-chan any,
		values ...any,
	) <-chan any {
		valueStream := make(chan any)
		go func() {
			defer close(valueStream)
			for {
				for _, v := range values {
					select {
					case <-done:
						return
					case valueStream <- v:
					}
				}
			}
		}()
		return valueStream
	}

	take := func(
		done <-chan any,
		valueStream <-chan any,
		num int,
	) <-chan any {
		resultStream := make(chan any)
		go func() {
			defer close(resultStream)
			for range num {
				select {
				case <-done:
					return
				case resultStream <- <-valueStream:
				}
			}
		}()
		return resultStream
	}

	repeatFn := func(
		done <-chan any,
		fn func() any,
	) <-chan any {
		resStream := make(chan any)
		go func() {
			defer close(resStream)
			for {
				select {
				case <-done:
					return
				case resStream <- fn():
				}
			}
		}()
		return resStream
	}

	done := make(chan any)
	defer close(done)
	for num := range take(done, repeat(done, 2, 2), 3) {
		fmt.Printf("%d ", num)
	}

	rand := func() any { return rand.Int() }
	for num := range take(done, repeatFn(done, rand), 5) {
		fmt.Println(num)
	}
}
