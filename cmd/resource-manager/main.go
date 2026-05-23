package main

import (
	"fmt"
	"math/rand/v2"
	"sync"
)

func main() {
	wg := sync.WaitGroup{}

	wg.Go(func() {
		g := rand.IntN(10)
		fmt.Printf("g: %d ", g)
	})
	wg.Wait()
}
