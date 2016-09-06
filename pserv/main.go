package main

import (
	"coin/cpb"
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port          = ":50051"
	enoughWorkers = 3
)

// server is used to implement cpb.CoinServer.
type server struct{}

// logger type is for the users login details
type logger struct {
	sync.Mutex //anonymous
	nextID     int
	loggedIn   map[string]int
}

var users logger

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) {
	users.Lock()
	defer users.Unlock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

var getwork chan struct{} // cpb.Abort
var block string

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("GetWork req: %+v\n", in)
	inGate <- in.Name     // register
	<-getwork             // all must wait
	work := fetchWork(in) // work assigned this miner
	return &cpb.GetWorkReply{Work: work}, nil
}

// prepares the candidate block and also provides user specific coibase data
func fetchWork(in *cpb.GetWorkRequest) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: in.Name, Block: []byte(block)}
}

var inGate, outGate chan string // unbuffered

func incomingGate() {
	for {
		for i := 0; i < enoughWorkers; i++ {
			fmt.Printf("(%d) registered %s\n", i, <-inGate)
		}
		close(getwork)
		endrun = make(chan struct{})
		go getNewBlocks() // models watching the entire network, timeout our search
	}
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	if newblock.Cancel() { // cancel previous getNewBlocks
		// was previously called!!
		return &cpb.AnnounceReply{Ok: false}, nil // we are late
	}
	checked := true // TODO - accommodate possible mistaken solution
	fmt.Printf("\nWe won!: %+v\n", soln)
	endRun()
	return &cpb.AnnounceReply{Ok: checked}, nil
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	outGate <- in.Name // register
	<-endrun           // wait for valid solution  OR timeout
	return &cpb.GetCancelReply{Ok: true}, nil
}

var endrun chan struct{} // cpb.Abort

func endRun() {
	for i := 0; i < enoughWorkers; i++ {
		fmt.Printf("[%d] de_register %s\n", i, <-outGate)
	}
	close(endrun) // cancel waiting for a valid stop
	fmt.Printf("\nNew race!\n--------------------\n")
	newblock.Revive()
	getwork = make(chan struct{})

	block = fmt.Sprintf("BLOCK:\n%v", time.Now())
}

var newblock cpb.Abort

// getNewBlocks watches the network for external winners and stops search if we exceed period secs
func getNewBlocks() {
	select {
	case <-newblock.Chan():
		return // let announce call endRun
	case <-time.After(17 * time.Second):
	}
	// otherwise reach this after 17 seconds
	endRun()
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	newblock.New() // = make(chan struct{})
	getwork = make(chan struct{})

	inGate = make(chan string)
	outGate = make(chan string)

	block = fmt.Sprintf("BLOCK:\n%v", time.Now())
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	go incomingGate()

	s := grpc.NewServer()
	cpb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
