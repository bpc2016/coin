package main

import (
	cpb "coin/service"
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
	port      = ":50051"
	numMiners = 3
	timeOut   = 14
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

var getwork chan struct{}

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("GetWork req: %+v\n", in)
	inGate <- in.Name // register
	<-getwork         // all must wait
	return &cpb.GetWorkReply{Work: fetchWork(in.Name)}, nil
}

var block string

// prepares the candidate block and also provides user specific coibase data
func fetchWork(name string) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: name, Block: []byte(block)}
}

var inGate, outGate chan string

func incomingGate() {
	for {
		for i := 0; i < numMiners; i++ {
			fmt.Printf("(%d) registered %s\n", i, <-inGate)
		}
		close(getwork)
		endrun = make(chan struct{})
		settled.Lock()
		settled.ch = make(chan struct{})
		settled.Unlock()
		go extAnnounce()
		// go getNewBlocks()
	}
}

type win struct {
	who  string
	data cpb.Win
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {

	// if newblock.Cancel() { // cancel previous getNewBlocks
	// 	return &cpb.AnnounceReply{Ok: false}, nil // we are late
	// }
	// checked := true // TODO - accommodate possible mistaken solution
	// fmt.Printf("\nWe won!: %+v\n", soln)
	// endRun()

	won := vetWin(win{who: "client", data: *soln.Win})
	return &cpb.AnnounceReply{Ok: won}, nil
}

// extAnnounce is the analogue of 'Announce'
func extAnnounce() {
	select {
	case <-settled.ch:
		return
	case <-time.After(timeOut * time.Second):
		vetWin(win{who: "outsider", data: cpb.Win{Coinbase: "", Nonce: 0}}) // bogus
		return
	}
}

// var mu sync.Mutex // guards settled
// var settled chan struct{}

var settled struct {
	sync.Mutex
	ch chan struct{}
}

// vetWins handle wins - all are directed here
func vetWin(thewin win) bool {
	settled.Lock()
	defer settled.Unlock()
	select {
	case <-settled.ch: // closed if already have a winner
		return false
	default:
		fmt.Printf("\nWinner is: %+v\n", thewin)
		close(settled.ch) // until call for new run resets this one
		endRun()          // SOLE call to endRun
		return true
	}
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	outGate <- in.Name // register
	<-endrun           // wait for valid solution  OR timeout
	return &cpb.GetCancelReply{Ok: true}, nil
}

var endrun chan struct{}

// var hereFirst struct {
// 	sync.RWMutex
// 	set bool
// }

func endRun() {
	// hereFirst.RLock()
	// if hereFirst.set {
	// 	hereFirst.RUnlock()
	// 	return // skip this one
	// }
	// hereFirst.RUnlock()

	// // get exclusive Lock
	// hereFirst.Lock()
	// defer hereFirst.Unlock()
	// hereFirst.set = true

	for i := 0; i < numMiners; i++ {
		fmt.Printf("[%d] de_register %s\n", i, <-outGate)
	}
	close(endrun) // cancel waiting for a valid stop
	fmt.Printf("\nNew race!\n--------------------\n")
	//newblock.Revive()
	getwork = make(chan struct{})
	//hereFirst.set = false
	block = fmt.Sprintf("BLOCK: %v", time.Now())
}

// var newblock Abort

// getNewBlocks watches the network for external winners and stops search if we exceed period secs
// func getNewBlocks() {
// 	select {
// 	case <-newblock.Chan():
// 		return // let announce call endRun
// 	case <-time.After(timeOut * time.Second):
// 	}
// 	endRun() // otherwise reach this after timeOut seconds
// }

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1

	//newblock.New() // = make(chan struct{})
	getwork = make(chan struct{})

	inGate = make(chan string)
	outGate = make(chan string)

	block = fmt.Sprintf("BLOCK: %v", time.Now())
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
