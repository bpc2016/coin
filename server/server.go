package main

import (
	cpb "coin/service"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"sync"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	index     = flag.Int("index", 0, "RPC port is 50051+index") //; debug port is 36661+index")
	numMiners = flag.Int("miners", 3, "number of miners")       // DOESNT include the external one
	debug     = flag.Bool("d", false, "debug mode")
)

// logger type is for the users login details  login OMIT
type logger struct {
	sync.Mutex
	nextID   int
	loggedIn map[string]int
}

var users logger

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) { // HL
	users.Lock()
	defer users.Unlock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

// nigol OMIT

type blockdata struct {
	sync.Mutex
	data string
}

var block blockdata // models the block information - basis of 'work'

type locked struct {
	sync.Mutex
	on bool
	ch chan struct{}
}

var tooLate locked

// GetWork implements cpb.CoinServer, synchronise start of miners
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	debugF("Work request: %+v\n", in)
	s.signIn <- in.Name // register
	// s.start.Wait()      // all must wait, start when this is Done()
	<-tooLate.ch // all must wait, start when this is closed
	block.Lock()
	work := &cpb.Work{Coinbase: in.Name, Block: []byte(block.data)}
	block.Unlock()
	return &cpb.GetWorkReply{Work: work}, nil
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	//won := s.vetWin(*soln.Win)
	tooLate.Lock()
	defer tooLate.Unlock()
	if tooLate.on {
		return &cpb.AnnounceReply{Ok: false}, nil
	}
	// we have a winner
	tooLate.on = true // until call for new run resets this one
	resultchan <- *soln.Win
	for i := 0; i < *numMiners; i++ {
		<-s.signOut
	}
	tooLate.ch = make(chan struct{})
	s.stop.Done() // issue cancellation to our clients
	// s.start.Add(1) // reset getwork start
	return &cpb.AnnounceReply{Ok: true}, nil
}

// GetCancel blocks until a valid stop condition then broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	s.signOut <- in.Name // register
	s.stop.Wait()        // wait for valid solution  OR timeout
	serv := *index
	return &cpb.GetCancelReply{Index: uint32(serv)}, nil
}

// vetWins handles wins - all are directed here
// func (s *server) vetWin(thewin cpb.Win) bool {
// 	s.search.Lock()
// 	defer s.search.Unlock()
// 	won := false
// 	select {
// 	case <-s.search.ch: // closed if already have a winner
// 	default: // should vet the solution, return false otherwise!!
// 		close(s.search.ch)   // until call for new run resets this one
// 		resultchan <- thewin // fmt.Sprintf("Winner: %+v", thewin)
// 		for i := 0; i < *numMiners; i++ {
// 			<-s.signOut
// 		}

// 		s.stop.Done()  // issue cancellation to our clients
// 		s.start.Add(1) // reset getwork start
// 		won = true
// 	}
// 	return won
// }

type privchannel struct {
	sync.Mutex
	ch chan struct{}
}

// server is used to implement cpb.CoinServer.
type server struct {
	signIn  chan string    // for registering users in getwork
	signOut chan string    // for registering leaving users in getcancel
	search  privchannel    // search.ch is closed when we have dclared a winner
	start   sync.WaitGroup // this is the tricky one, stop here for company
	stop    sync.WaitGroup
}

var blockchan chan string

// IssueBlock receives the new block from frontend : implements cpb.CoinServer
func (s *server) IssueBlock(ctx context.Context, in *cpb.IssueBlockRequest) (*cpb.IssueBlockReply, error) {
	blockchan <- in.Block
	return &cpb.IssueBlockReply{Ok: true}, nil
}

// this comes from this server's role with frontend as client
func getNewBlock() {
	temp := <-blockchan // note that this will block if EXTERNAL absent
	block.Lock()
	block.data = temp
	block.Unlock()
}

var resultchan chan cpb.Win //string

// GetResult sends back win to frontend : implements cpb.CoinServer
func (s *server) GetResult(ctx context.Context, in *cpb.GetResultRequest) (*cpb.GetResultReply, error) {
	result := <-resultchan                             // wait for a result
	fmt.Printf("sendresult: %d, %v\n", *index, result) // send this back to client
	return &cpb.GetResultReply{Winner: &result, Index: uint32(*index)}, nil
}

// utilities
func fatalF(message string, err error) {
	if err != nil {
		log.Fatalf(message+": %v", err)
	}
}

func debugF(format string, args ...interface{}) {
	if *debug {
		log.Printf(format, args...)
	}
}

// initalise
func init() {
	users.loggedIn = make(map[string]int)
	users.nextID = -1
}

func main() {
	flag.Parse()

	*numMiners++ // to include the EXTERNAL

	port := fmt.Sprintf(":%d", 50051+*index)
	lis, err := net.Listen("tcp", port) // RPC port - localhost?
	fatalF("failed to listen", err)

	s := new(server)

	s.signIn = make(chan string, *numMiners)
	s.signOut = make(chan string, *numMiners)
	//s.start.Add(1) // get work start
	tooLate.ch = make(chan struct{}) // prepare for getwork start

	blockchan = make(chan string, 1) // buffered
	resultchan = make(chan cpb.Win)

	go func() {
		for {
			tooLate.Lock()
			tooLate.on = false
			tooLate.Unlock()
			getNewBlock()
			for i := 0; i < *numMiners; i++ { // loop blocks here until miners are ready
				<-s.signIn
			}

			fmt.Printf("\n--------------------\nNew race!\n")
			s.stop.Add(1)     // prep channel for getcancels
			close(tooLate.ch) // start our miners
			// s.search.ch = make(chan struct{}) // reset this channel
			// s.start.Done() // start our miners
		}
	}()

	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
