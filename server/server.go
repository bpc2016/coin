package main

import (
	cpb "coin/service"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
	// numMiners = 3
	timeOut = 14
)

var numMiners = flag.Int("n", 3, "number of miners")

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

var workgate chan struct{}
var signIn chan string

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	fmt.Printf("Work requested: %+v\n", in)
	signIn <- in.Name // register
	<-workgate        // all must wait
	return &cpb.GetWorkReply{Work: fetchWork(in.Name)}, nil
}

var block string

// prepares the candidate block and also provides user specific coibase data
func fetchWork(name string) *cpb.Work { // TODO -this should return err as well
	return &cpb.Work{Coinbase: name, Block: []byte(block)}
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
		vetWin(win{who: "outsider", data: cpb.Win{Coinbase: "", Nonce: 0}}) // bogus
		return
	}
}

var endrun chan struct{}
var signOut chan string

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	signOut <- in.Name // register
	<-endrun           // wait for valid solution  OR timeout
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
		fmt.Printf("\nWinner is: %+v\n", thewin)
		close(settled.ch) // until call for new run resets this one
		for i := 0; i < *numMiners; i++ {
			miner := <-signOut
			fmt.Printf("[%d] de_register %s\n", i, miner)
		}
		close(endrun) // issue cancellation to our clients
		workgate = make(chan struct{})
		block = fmt.Sprintf("BLOCK: %v", time.Now())

		fmt.Printf("\nNew race!\n--------------------\n")
		return true
	}
}

func startRun() {
	for {
		for i := 0; i < *numMiners; i++ {
			miner := <-signIn
			fmt.Printf("(%d) registered %s\n", i, miner)
		}
		endrun = make(chan struct{})
		settled.Lock()
		settled.ch = make(chan struct{})
		settled.Unlock()
		close(workgate)  // start our miners
		go extAnnounce() // start external miners
	}
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	signIn = make(chan string)
	signOut = make(chan string)
	block = fmt.Sprintf("BLOCK: %v", time.Now())
	workgate = make(chan struct{})

	go startRun()

	s := grpc.NewServer()
	cpb.RegisterCoinServer(s, &server{})
	s.Serve(lis)
}
