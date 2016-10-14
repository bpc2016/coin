package main

import (
	"coin"
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
	allowedTime          = 2  // "number of seconds before miner declared NOT alive"
	allowedConductorTime = 20 // number of seconds for the conductor
)

var (
	index     = flag.Int("index", 0, "RPC port is 50051+index") //; debug port is 36661+index")
	numMiners = flag.Int("miners", 3, "number of miners")       // DOESNT include the external one
	debug     = flag.Bool("d", false, "debug mode")
)

type lockMap struct {
	sync.Mutex
	nextID   int
	loggedIn map[string]int
}

type blockdata struct {
	u      []byte // upper coinbase
	l      []byte // lower coinbase
	height uint32 // blockheight
	blk    []byte // 80 byte block header partially filled
	merk   []byte // merkle root skeleton - multiple of 32 bytes
}

type lockBlock struct {
	sync.Mutex
	data blockdata
}

type lockChan struct {
	sync.Mutex
	winnerFound bool
	ch          chan struct{}
}

var (
	users      lockMap
	block      lockBlock      // models the block information - basis of 'work'
	run        lockChan       // channel that controls start of run
	signIn     chan string    // for registering users in getwork
	signOut    chan string    // for registering leaving users in getcancel
	stop       sync.WaitGroup // control cancellation issue
	blockchan  chan blockdata // for incoming block
	resultchan chan cpb.Win   // for the winner decision
)

// Login implements cpb.CoinServer
func (s *server) Login(ctx context.Context, in *cpb.LoginRequest) (*cpb.LoginReply, error) { // HL
	users.Lock()
	defer users.Unlock()
	if _, ok := users.loggedIn[in.Name]; ok {
		return nil, errors.New("You are already logged in!")
	}
	users.nextID++
	users.loggedIn[in.Name] = users.nextID // HL
	return &cpb.LoginReply{Id: uint32(users.nextID)}, nil
}

// GetWork implements cpb.CoinServer, synchronises start of miners, hands out work
func (s *server) GetWork(ctx context.Context, in *cpb.GetWorkRequest) (*cpb.GetWorkReply, error) {
	debugF("Work request: %+v\n", in)
	signIn <- in.Name // HL
	<-run.ch          // HL
	// customise work for this miner
	work := setWork(in.Name)
	return &cpb.GetWorkReply{Work: work}, nil
}

func setWork(name string) *cpb.Work {
	if name == "EXTERNAL" {
		return &cpb.Work{Coinbase: []byte{}, Block: []byte{}, Skel: []byte{}}
	}
	block.Lock()
	minername := fmt.Sprintf("%d:%s", *index, name)
	miner := users.loggedIn[name]
	upper := block.data.u
	lower := block.data.l
	blockHeight := block.data.height
	partblock := block.data.blk
	merkSkel := block.data.merk
	// generate actual coinbase txn
	coinbaseBytes, err := coin.GenCoinbase(upper, lower, blockHeight, miner, minername)
	fatalF("failed to set block data", err)
	block.Unlock()
	fmt.Printf("miner: %s\ncoinbase:\n%x\n", minername, coinbaseBytes)
	return &cpb.Work{Coinbase: coinbaseBytes, Block: partblock, Skel: merkSkel}
}

// Announce responds to a proposed solution : implements cpb.CoinServer
func (s *server) Announce(ctx context.Context, soln *cpb.AnnounceRequest) (*cpb.AnnounceReply, error) {
	run.Lock()
	defer run.Unlock()
	if run.winnerFound {
		return &cpb.AnnounceReply{Ok: false}, nil
	}
	// we have a  winner
	run.winnerFound = true  // HL
	resultchan <- *soln.Win // HL
	fmt.Println("starting signout numminers = ", *numMiners)
	WaitFor(signOut, "out")
	run.ch = make(chan struct{}) // HL
	stop.Done()                  // HL
	return &cpb.AnnounceReply{Ok: true}, nil
}

// GetCancel broadcasts a cancel instruction : implements cpb.CoinServer
func (s *server) GetCancel(ctx context.Context, in *cpb.GetCancelRequest) (*cpb.GetCancelReply, error) {
	signOut <- in.Name // HL
	stop.Wait()        // HL
	serv := *index
	return &cpb.GetCancelReply{Index: uint32(serv)}, nil // HL
}

// server is used to implement cpb.CoinServer.
type server struct{}

// IssueBlock receives the new block from Conductor : implements cpb.CoinServer
func (s *server) IssueBlock(ctx context.Context, in *cpb.IssueBlockRequest) (*cpb.IssueBlockReply, error) {
	blockchan <- blockdata{in.Lower, in.Upper, in.Blockheight, in.Block, in.Merkle}
	users.loggedIn["EXTERNAL"] = 1 // we login conductor here
	return &cpb.IssueBlockReply{Ok: true}, nil
}

// GetResult sends back win to Conductor : implements cpb.CoinServer
func (s *server) GetResult(ctx context.Context, in *cpb.GetResultRequest) (*cpb.GetResultReply, error) {
	result := <-resultchan                             // wait for a result
	fmt.Printf("sendresult: %d, %v\n", *index, result) // OMIT
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

// WaitFor allows for the loss of a miners
func WaitFor(sign chan string, direction string) {
	alive := make(map[string]bool) // HL
	count := 1
	alive[<-sign] = true // we need at least one! ... then the rest ...
	for i := 1; i < *numMiners; i++ {
		select {
		case <-time.After(allowedTime * time.Second): // exit, time is up
			goto done
		case c := <-sign:
			alive[c] = true
			count++
		}
	}
done:
	if direction == "in" && count < *numMiners {
		for name := range users.loggedIn {
			if !alive[name] && name != "EXTERNAL" {
				fmt.Printf("DEAD: %s\n", name)
				delete(users.loggedIn, name)
			}
		}
	}
	fmt.Printf("miners %s = %d\n", direction, count)
}

// coinbase accepts data from work, result is tailored to miner
// func coinbase(upper []byte, lower []byte, blockHeight int,
// 	miner int, minername string) coin.Transaction {
// 	txn, err := coin.GenCoinbase(upper, lower, blockHeight, miner, minername)
// 	fatalF("failed to generate coinbase transaction", err)
// 	// fmt.Printf("%x", txn)
// 	return coin.Transaction(txn) // convert to a transaction type
// }

func main() {
	flag.Parse() // HL
	users.loggedIn = make(map[string]int)
	users.nextID = -1
	*numMiners++ // to include the Conductor (EXTERNAL)

	port := fmt.Sprintf(":%d", 50051+*index) // HL
	lis, err := net.Listen("tcp", port)
	fatalF("failed to listen", err)

	signIn = make(chan string, *numMiners)  // register incoming miners
	signOut = make(chan string, *numMiners) // register miners receipt of cancel instructions
	blockchan = make(chan blockdata, 1)     // transfer block data
	run.ch = make(chan struct{})            // signal to start mining
	resultchan = make(chan cpb.Win)         // transfer solution data

	go func() {
		for {
			for {
				select {
				case block.data = <-blockchan: // HL
					goto start
				case <-time.After(allowedConductorTime * time.Second): // HL
					fmt.Println("Need a live conductor!")
				}
			}
		start:
			WaitFor(signIn, "in") // HL
			fmt.Printf("\n--------------------\nNew race!\n")
			run.winnerFound = false // HL
			stop.Add(1)             // HL
			close(run.ch)           // HL
		}
	}()
	s := new(server)
	g := grpc.NewServer()
	cpb.RegisterCoinServer(g, s)
	g.Serve(lis)
}
