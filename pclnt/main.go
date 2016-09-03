package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"coin/cpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "busiso"
)

func getWork(c cpb.CoinClient, name string) *cpb.Work {
	// get ready, get set ... this needs to block
	r2, err := c.GetWork(context.Background(), &cpb.GetWorkRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Got work %+v\n", r2.Work)
	return r2.Work
}

var cancelled bool

// search tosses two dice waiting for a double 5
func search(id uint32, work *cpb.Work) {
	foundit := false
	tick := time.Tick(1 * time.Second) // spin wheels
	for c := 0; c < 40; c++ {
		a, b := toss(), toss()
		if a == b && a == 5 {
			foundit = true
			break
		}
		<-tick
		fmt.Println(id, " ", c)
		// check for a stop order
		if cancelled {
			break
		}
	}
	if foundit {
		fmt.Printf("== %d == FOUND it\n", id)
		// 	toServer <- id // declare we are done
	}
	fmt.Printf("[%d] .. BYE\n", id)
}

// dice
func toss() int {
	return rand.Intn(6)
}

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := cpb.NewCoinClient(conn)

	// Contact the server and print out its response.
	name := defaultName
	if len(os.Args) > 1 {
		name = os.Args[1]
	}
	r1, err := c.Login(context.Background(), &cpb.LoginRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	log.Printf("Login successful. Assigned id: %d\n", r1.Id)

	// TODO - convert the context to have a timeout here
	rand.Seed(time.Now().UTC().UnixNano())

	// search
	task := getWork(c, name)
	search(r1.Id, task)
}
