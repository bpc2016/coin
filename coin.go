// this is an implementation of bitcoin  structures that will
// be exchanged between server and worker pi miners in the pool

// Package coin provides structures for manipualting bitcoin.
// Our use case is a single Server running a full node and a community
// of Clients running lightweight Go programs to perform pooled mining.
//
package coin

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
)

// Block is the blockheaderobject - 80 bytes
type Block []byte

// Task type is the set of byte slices that each client receives.
// Ultimately, the blockheader is filled then its double hash compared
// to target
type Task struct {
	BH       []byte // partially filled blockheader 80 bytes
	MH       []byte // merkle header 32 bytes (TODO - remove)
	Target   []byte // computed value of target
	Share    []byte // lesser target for still-alive messaging
	Skeleton []byte // for constructing new Merkle roots 32*n bytes
	CoinBase string // equiv of both coinb1, coinb2 (TODO --> []byte)
}

// InComing captures the raw data that the server assembles to create
// work for client miners
type InComing struct {
	Version    uint32   // version
	PrevBlock  string   // previous block header hash string
	MerkleRoot string   // merkle root hash hex string (TODO - remove)
	TxHashes   []string // hex transaction info - all except coinbase
	CoinBase   string   // hex encoded coinbase
	TimeStamp  uint32   // time
	Bits       uint32   // target <--> difficulty
	Share      int      // lesser target for still-alive messaging
}

/*
type Coinbase struct {
	Version   uint32 // version (4)
	InCount   byte   // input count - will always be just 01 (1)
	PrevBlock string // previous block header hash string (32)
	Index     uint32 // always 0xffffffff	(4)
	Script    []byte // probably 32 bytes
	Sequence  uint32 // always 0x00000000
	OutCount  byte   // output count - will always be just 01
	Value     string // hex of amount due here = coinbase + fees uint64 (8)
	Scriptlen byte   // 0x19 => 25byte length of PayScript??
	PayScript string // likely 0x76a9....
	LockTime  uint32 // will always be just 00
}
*/

// Sha256 returns the Sha256 hash of its input byte slice data
// and returns a slice rather than the normal 32-byte array
func Sha256(data []byte) []byte {
	mid := sha256.Sum256(data)
	return mid[:]
}

// DoubleSha256 applies Sha256 twice to data
// returning the digest as a 32-byte slice
func DoubleSha256(data []byte) ([]byte, error) {
	if data == nil {
		return nil, errors.New("Empty bytes cannot be hashed")
	}
	mid1 := sha256.Sum256(data)
	mid2 := sha256.Sum256(mid1[:])
	return mid2[:], nil
}

// Hash2 applies Sha256 twice to the concatenation
// of byte sequences seq1 and seq2
func Hash2(seq1, seq2 []byte) []byte {
	concat := append(seq1, seq2...)
	hash, err := DoubleSha256(concat)
	if err != nil {
		log.Fatal(err)
	}
	return hash
}

// Reverse reverses byte slice s
func Reverse(s []byte) []byte {
	L := len(s)
	t := make([]byte, L)
	for i, v := range s {
		t[L-1-i] = v
	}
	return t
}

// GenLogin generates the login for use by client given user, key and time
func GenLogin(user uint32, key string, time string) (string, error) {
	login := ""
	// the userid
	userbytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(userbytes, user)
	// the key
	keyhex, err := hex.DecodeString(fmt.Sprintf("%x", key)) // *key
	if err != nil {
		return login, err
	}
	timehex, err := hex.DecodeString(time)
	if err != nil {
		return login, err
	}
	concat1 := append(userbytes, keyhex...)
	concat2 := append(timehex, concat1...)
	hash := Sha256(concat2)
	login = fmt.Sprintf("%x", hash[0:6])
	return login, nil
}
