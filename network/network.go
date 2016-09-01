// Package network models a setup with a single server
// and multiple clients searching for bitcoin
package main

import (
	"fmt"
	"math/rand"
	"time"
)

// dice
func toss() int {
	return rand.Intn(6)
}

// client has n id and  owrks on a load of work
// needs to communicate with the server
func client(id int, load string, toServer chan int) {
	fmt.Printf("howdy, I am %d and have work: %s\n", id, load)
	foundit := false
	// spin wheels
	tick := time.Tick(1 * time.Second)
	// decide to win or lose
	for c := 0; c < 40; c++ {
		a, b := toss(), toss()
		if a == b && a == 5 {
			fmt.Printf("<%d> --> found it\n", id)
			foundit = true
			break
		}
		<-tick
		fmt.Println(id, " ", c)
		// check for a stop order
		if cancelled() {
			fmt.Printf("(%d) got QUIT signal!\n", id)
			break
		}
	}
	if foundit {
		toServer <- id // declare we are done
	}
	toServer <- id
	fmt.Printf("**[%d] .. BYE\n", id)
}

var done = make(chan struct{})

// this will return false until channel done is closed
func cancelled() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

// server doles out work
func server(work []string) {
	listen := make(chan int) // where server waits for word of success
	for i, w := range work {
		go func(i int, w string) {
			client(i, w, listen)
		}(i, w)
	}
	// wait for a response,
	// TODO or an external timeout
	theOne := <-listen

	fmt.Printf("server: found by %d\n ", theOne)
	// stop the rest
	close(done)
	// drain the listen channel
	for i := 0; i < len(work); i++ {
		fmt.Printf(">>from %d\n", <-listen)
	}
	fmt.Println("server done, bye!")
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	server([]string{"al", "ph", "bet", "ic"})
}
