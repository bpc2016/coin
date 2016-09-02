// Package netcl implements a Client to netserver.
package netcl

import (
	"fmt"
	"math/rand"
	"time"
)

// dice
func toss() int {
	return rand.Intn(6)
}

// Client has an id and  works on a load of work - tosses two dice waiting for a double 5
func Client(id int, load string, toServer chan int, cancelled func() bool) {
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
