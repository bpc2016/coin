// Package network models a setup with a single server
// and multiple clients searching for bitcoin
package main

import (
	"fmt"
	"math/rand"
	"time"
)

//============================== clients ==========================================

// dice
func toss() int {
	return rand.Intn(6)
}

// client has an id and  works on a load of work - tosses two dice waiting for a double 5
func client(id int, load string, toServer chan int) {
	fmt.Printf("I am %d and have work: %s\n", id, load)
	foundit := false
	tick := time.Tick(1 * time.Second) // spin wheels
	for c := 0; c < 40; c++ {
		a, b := toss(), toss()
		if a == b && a == 5 {
			fmt.Printf("== %d == FOUND it\n", id)
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
	// fmt.Printf("[%d] .. BYE\n", id)
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

//========================= server ===========================================

var network = make(chan struct{}) // messages from the network

// external counts down for m seconds then tells us to stop unless we cancel with a find
func external(m int) {
	tick := time.Tick(100 * time.Millisecond)
	for cd := m * 10; cd > 0; cd-- { // effectively 10 second countdown
		if cancelled() {
			fmt.Printf("Network cancelled!\n")
			return
		}
		<-tick // wait a sec
	}
	// we get here - kickoff
	fmt.Printf("NETWORK CALLING!\n")
	network <- struct{}{}
}

// server doles out work
func server(work []string) {
	go external(8)           // start listening to the network
	listen := make(chan int) // where server waits for word of success
	for i, w := range work {
		go func(i int, w string) {
			client(i, w, listen)
		}(i, w)
	}
	// wait for a response or an external timeout - block
	select {
	case theOne := <-listen:
		fmt.Printf("Server: win found by %d\n", theOne)
		close(done)
		// drain the listen channel
		for i := 0; i < len(work); i++ {
			fmt.Printf("[%d] >>\n", <-listen)
		}
	case <-network:
		close(done)
		// drain the listen channel
		for i := 0; i < len(work); i++ {
			fmt.Printf("(%d) >>\n", <-listen)
		}
	}
	// ... before quitting
	fmt.Println("Server: ... bye!\n---------------------------")
	done = make(chan struct{}) // for next call
}

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	for {
		server([]string{"al", "ph", "bet", "ic"})
		time.Sleep(2 * time.Second)
	}
}
