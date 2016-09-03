package main

import (
	"log"
	"os"

	"coin/cpb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "busiso"
)

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
	log.Printf("Login successful. id: %d, work: %s", r1.Id, r1.Work)

	// get ready, get set ... this needs to block
	r2, err := c.Mine(context.Background(), &cpb.MineRequest{Name: name})
	if err != nil {
		log.Fatalf("could not login: %v", err)
	}
	if r2.Ok {
		log.Printf("Mine request successful")
	} else {
		log.Printf("Mine request FAILED")
	}
}
