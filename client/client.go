// Package Client is the one that actually mines bitcoin.
// It receives a task from the server and creates associated merkle root hash
// then tests succesive nonces. The idea is to divide the search between N clients, 
// each assigned a search-class between 0 and N-1. The client examines nonces 
// starting from Cls in steps of N to the ceiling of 0xffffffff 
package client

import (
	"bytes"
	"coin"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"time"
)

// offsets of parts of the blockheader
const (
	b_version = 4  // 4 byte Version
	b_prevblk = 36 // 32 byte previous block hash
	b_merkle  = 68 // 32 byte merkle root hash
	b_time    = 72 // 4 byte Unix time
	b_bits    = 76 // 4 byte target (encoded) for difficulty rating
	b_nonce   = 80 // 4 byte nonce used
)

const (
	cores = 2          //  goroutines used per job - set to  number of cores
	top   = 0xffffffff // topmost uint32 = 2^32-1
)

// This information defines parameters of the client machine within
// the pool.
var (
	N   uint32 = 100 // total number in pool
	Cls uint32 = 28  // the residue class of this server
)

type Job struct {
	Block  []byte // Blockheader for this job (80 bytes)
	hash   []byte // Receives the computed hash (32 bytes)
	share  []byte // easy target
	target []byte // wanted hash target
}

type shareData struct {
	core  uint32
	nonce uint32
	time  string
	hash  []byte
}


type report struct{
	success bool
	block []byte
}

// Search takes uses the blockheader from task t to look for a winning hash with
// this client's preset residue class.  It exits with either
// (1) no solution and entire nonce range searched, or
// (2) a solution is found
func Search(t coin.Task) {
	foundit := make(chan report, cores) // signal to end search
	shares := make(chan shareData)    //  discovered shares

	for i := uint32(0); i < cores; i++ { // set off the cores searches
		go workon(t, i, foundit, shares)
	}

	// goroutine collecting share finds
	go func() {
		for {
			data := <-shares
			fmt.Printf("* [%d] share at nonce %d, after: %s\n%x\n",
				data.core, data.nonce, data.time, data.hash)
		}
	}()

	// main handles search results
	for i := 0; i < cores; i++ {
		endwith := <-foundit
		if endwith.success {
			fmt.Printf("Bye! block:\n%x\n", endwith.block)
			break
		}
	}
}

// workon performs a search of task t starting from nonce Cls+i*N
// communicates with calling routine via channels
// TODO - make this cycle rather than quit,
// TODO - make this respond to stop signal
func workon(t coin.Task, i uint32, foundit chan report, shares chan shareData) {
	j := Job{}
	j.PrepSearch(t) // make byte seq w/ Merkle root data
	start := time.Now()
	nce := Cls + i*N  // starting nonce
	step := cores * N // progression
	last := ((top-nce)/step)*step + nce
	fmt.Printf("goroutine: %d start from: %d at  %s\n",
		i, nce, start)
	for {
		gets := j.CheckHashAt(nce)
		if gets.goodHash {
			end := time.Now()
			duration := time.Since(start).String()
			fmt.Printf("*** [%d] Found nonce = %d\n*** At %s\n*** Duration: %s\n*** Hash: %x",
				i, nce, end, duration, gets.hash)
			foundit <- report{true,j.Block}	// signal success
			close(foundit)  				// and terminate search
			break
		}
		if gets.shareHash {
			shares <- shareData{i, nce, time.Since(start).String(), gets.hash}
		}
		if nce == last {
			fmt.Printf("* [%d] quitting, nonce:%d hex:%x\n", i, nce, nce)
			foundit <- report{false,nil} // signal failure
			break
		}
		nce += step
	}
}

// PrepSearch loads j.Block with data from task t. It is called
// at the start of any search and may be recalled to regenerate the
// Block if nonce values exhaust their 32-bit limit
func (j *Job) PrepSearch(t coin.Task) {
	j.Block = make([]byte, 80)
	copy(j.Block, t.BH)              //get the entire struct from here
	freshMerkle := buildMerkle(t.MH) // add freshly generated full merkle root
	copy(j.Block[b_prevblk:b_merkle], freshMerkle)
	j.target = t.Target // recall target, share
	j.share = t.Share
}

// writeUint32 is a convenience: write v into 4 bytes of slice
func writeUint32(slice []byte, v uint32) {
	binary.LittleEndian.PutUint32(slice, v)
}


type mineData struct{
	shareHash bool
	goodHash bool
	hash []byte
}
// CheckHashAt includes the given nonce in the block held by job j. Using
// the current share and target respectively, it calculates the block hash
// and returns easy=true / hard=true if the hash is below share / target
// resp. Byte sequence thehash returns the computed hash
func (j *Job) CheckHashAt(nonce uint32) mineData { //(easy bool, hard bool, thehash []byte) {
	writeUint32(j.Block[b_bits:b_nonce], nonce)
	hash, err := coin.DoubleSha256(j.Block)
	if err != nil {
		log.Fatal(err)
	}
	j.hash = coin.Reverse(hash)
	// return this as  mineData object
	out := mineData {
		bytes.Compare(j.hash, j.share) < 0,  // hash < share target
		bytes.Compare(j.hash, j.target) < 0, // hash < target
		j.hash,
	}
	return out
	// // named returns
	// out.hash = j.hash
	// thehash = j.hash
	// easy = bytes.Compare(j.hash, j.share) < 0  // hash < share target
	// hard = bytes.Compare(j.hash, j.target) < 0 // hash < target
	// return
}

// buildMerkle each client needs to rebuild the root hash
// using data from the server when completed, this will depend on
// the client's identity - and a construction of a bitcoin account
func buildMerkle(merk []byte) []byte {
	// right now, just all of it
	// later will use the coinb data as well
	return coin.Reverse(merk) // was not reversed to allow this step
}

// GetMerkle takes Skeleton and Coinbase an computes the
// Merkle root (in hex)
func GetMerkle(coinbase string, skeleton []byte) (string, error) {
	N := len(skeleton) / 32 // the number of partial hashes = loops
	cbBytes, err := hex.DecodeString(coinbase)
	if err != nil {
		return "", err
	}

	// we need to reverse coinbase
	part := coin.Reverse(cbBytes)

	// then climb up the tree
	for i := 0; i < N; i++ {
		part = coin.Hash2(part, skeleton[32*i:32*(i+1)])
	}
	root := fmt.Sprintf("%x", coin.Reverse(part))
	return root, nil
}
