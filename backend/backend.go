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
	numMiners = 1
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

type win struct {
	who  string
	data cpb.Win
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	won := vetWin(win{who: "client", data: *soln.Win})
	return &cpb.AnnounceReply{Ok: won}, nil
}

// extAnnounce is the analogue of 'Announce'
func extAnnounce() {
	select {
	case <-settled.ch:
		return
	case <-time.After(timeOut * time.Second):
		vetWin(win{who: "outsider", data: cpb.Win{Coinbase: "external", Nonce: 0}}) // bogus
		return
	}
}

var minersOut, stop sync.WaitGroup

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	minersOut.Add(-1) // we work downwards from numMiners
	stop.Wait()       // wait for cancel signal
	return &cpb.GetCancelReply{Ok: true}, nil
}

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
		fmt.Printf("\nWinner: %+v\n", thewin)
		close(settled.ch)                            // until call for new run resets this one
		minersOut.Wait()                             // pause for all to get cancellation
		minersOut.Add(numMiners)                     // reset
		stop.Add(-1)                                 // issue cancellations
		block = fmt.Sprintf("BLOCK: %v", time.Now()) // new work
		run.Add(1)                                   // get ready for next work requests

		fmt.Printf("\nNew race!\n--------------------\n")
		return true
	}
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

var minersIn, run sync.WaitGroup

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("Work requested: %+v\n", in)
	minersIn.Add(-1) // we work downwards from numMiners
	run.Wait()       // all must wait here
	return &cpb.GetWorkReply{Work: fetchWork(in.Name)}, nil
}

var block string

// prepares the candidate block and also provides user specific coibase data
func fetchWork(name string) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: name, Block: []byte(block)}
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	block = fmt.Sprintf("BLOCK: %v", time.Now()) // new work
	run.Add(1)                                   // updated in vetWin

	minersIn.Add(1)
	minersOut.Add(1)

	go func() {
		for {
			minersIn.Wait() // for a work request
			// then prep the cancel request
			settled.Lock()
			settled.ch = make(chan struct{})
			settled.Unlock()
			stop.Add(1)

			run.Add(-1)      // start our miners
			go extAnnounce() // start external miners
			minersIn.Add(1)  // so that we pause above
		}
	}()

	s := grpc.NewServer()
	cpb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
