// Package netsrv implements the single server against netcl clients
package netsrv

import (
	"coin/netcl"
	"fmt"
	"time"
)

var done = make(chan struct{}) // messages to clients

// cancelled  returns false until channel done is closed. It signals client to stop
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
	for cd := m * 10; cd > 0; cd-- { // m second countdown
		if cancelled() {
			fmt.Printf("Network cancelled!\n")
			return
		}
		<-tick // wait a sec
	}
	fmt.Printf("NETWORK CALLING!\n")
	network <- struct{}{}
}

// Server doles out work
func Server(work []string) {
	go external(8)           // start listening to the network
	listen := make(chan int) // where server waits for word of success
	for i, w := range work {
		go func(i int, w string) {
			netcl.Client(i, w, listen, cancelled)
		}(i, w)
	}
	// wait for a response or an external timeout - block
	format := "(%d) >>\n" // default is failure - (%d) indicates closed w/o win
	select {
	case theOne := <-listen:
		fmt.Printf("Server: win found by %d\n", theOne)
		format = "[%d] >>\n"
	case <-network:
	}
	close(done)
	// drain the listen channel
	for i := 0; i < len(work); i++ {
		fmt.Printf(format, <-listen)
	}
	fmt.Println("Server: ... bye!\n---------------------------")
	done = make(chan struct{}) // for next call
}
