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
			fmt.Printf("** %d *** FOUND it\n", id)
			foundit = true
			break
		}
		<-tick
		fmt.Println(id, " ", c)
		// check for a stop order
		if cancelled() {
			break
		}
	}
	if foundit {
		toServer <- id // declare we are done
	}
	fmt.Printf("[%d] .. BYE\n", id)
	toServer <- id
}

var done = make(chan struct{}) // messages to clients

// this will return false until channel done is closed
func cancelled() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

// this will return false until channel done is closed
func clientsfoundit() bool {
	select {
	case <-done:
		return true
	default:
		return false
	}
}

var network = make(chan struct{}) // messages from the network

// external is from outside - tells us to stop
func external() {
	tick := time.Tick(1 * time.Second)
	for cd := 10; cd > 0; cd-- {
		if cancelled() {
			fmt.Printf("network cancelled!\n ")
			return
		}
		if clientsfoundit() {
			fmt.Printf("network cancelled!\n ")
			return
		}
		<-tick // wait a sec
	}
	// we get here - kickoff
	fmt.Printf("NETWORK CALLING!\n ")
	network <- struct{}{}
}

// server doles out work
func server(work []string) {
	listen := make(chan int) // where server waits for word of success
	for i, w := range work {
		go func(i int, w string) {
			client(i, w, listen)
		}(i, w)
	}
	// wait for a response or an external timeout
	select {
	case theOne := <-listen:
		fmt.Printf("server: found by %d\n ", theOne)
		close(done)
		// drain the listen channel
		for i := 0; i < len(work); i++ {
			fmt.Printf(">>from %d\n", <-listen)
		}
	case <-network:
		close(done)
		// drain the listen channel
		for i := 0; i < len(work); i++ {
			fmt.Printf(">>from %d\n", <-listen)
		}
	}
	// ... before quitting
	fmt.Println("server done ... bye!")
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	// for {
	go external()
	server([]string{"al", "ph", "bet", "ic"})
	done = make(chan struct{})
	network = make(chan struct{})
	go external()
	server([]string{"al", "ph", "bet", "ic"})
	// }
}
