// Package Server directs bitcoin mining operations. It runs a full
// node and when a new block search begins, distributes search data
// to remote client machines.
package server

import (
	"coin"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
)

// pointers to the parts of the blockheader
const (
	version = 4  // 4 byte Version
	prevblk = 36 // 32 byte previous block hash
	merkle  = 68 // 32 byte merkle root hash
	time    = 72 // 4 byte Unix time
	bits    = 76 // 4 byte target for difficulty rating
	nonce   = 80 // 4 byte nonce used
)

// SetTask takes incoming block data in and returns
// a Task object which a miner uses to search for a winning hash
func SetTask(in coin.InComing) coin.Task {
	bhdr := make([]byte, 80)

	// version
	writeUint32(bhdr[0:version], in.Version)

	// prevous block
	copy(bhdr[version:prevblk], coin.Reverse(hexToBytes(in.PrevBlock)))

	// time
	writeUint32(bhdr[merkle:time], in.TimeStamp)

	// bits (target)
	writeUint32(bhdr[time:bits], in.Bits)

	// share target
	if in.Share == 0 {
		in.Share = 3 // minimum
	}

	// Merkle root data
	merkle := make([]byte, 32)
	if in.MerkleRoot != "" {
		merkle = hexToBytes(in.MerkleRoot)
	}

	skeleton := []byte{} // default send empty
	if len(in.TxHashes) > 0 && in.CoinBase != "" {
		txes := append([]string{in.CoinBase}, in.TxHashes...)
		mRootStr, skel, err := Merkle(txes)
		if err == nil {
			merkle = hexToBytes(mRootStr)
			skeleton = skel
		}
	}

	return coin.Task{
		BH:       bhdr,                 // 80 byte slice
		MH:       merkle,               // 32 byte slice
		Target:   targetBytes(in.Bits), // 32 byte slice
		Share:    share(in.Share),      // 32 byte slice
		CoinBase: in.CoinBase,          // string
		Skeleton: skeleton,             // slice of 32*n bytes
	}
}

// writeUint32 write v into 4 bytes of slice
func writeUint32(slice []byte, v uint32) {
	binary.LittleEndian.PutUint32(slice, v)
}

// hexToBytes - takes a string s and returns a byte sequence
// assume that s is in bigendian, so add extra 0 if odd length
func hexToBytes(s string) []byte {
	if len(s)%2 == 1 {
		s = "0" + s
	}
	bts, err := hex.DecodeString(s)
	if err != nil {
		log.Fatal(err)
	}
	return bts
}

// targetBytes converts uint32 bits to a 32-byte sequence target
// which is compared to block hashes. A
// target is given by m*2**(8*(r-3)) where bits = r|m, r occupying
// the top byte and m the lower 3 bytes of this uint32.
func targetBytes(bits uint32) []byte {
	r := bits >> 24           // top byte
	m := (bits << 8) >> 8     // remaining three bytes
	b := make([]byte, 32)     // expect 32 byte result but item is shorter
	mBytes := make([]byte, 4) // receives m
	pos := 32 - r
	writeUint32(mBytes, m)
	copy(b[pos:pos+3], mBytes[1:])
	return b
}

// share  returns a 32 byte sequence with k leading 0's
// and the rest of the elements 0xff as a challenge that is
// easier than the actual target
func share(k int) []byte {
	byteseq := make([]byte, k)
	for i := 0; i < 32-k; i++ {
		byteseq = append(byteseq, 0xff)
	}
	return byteseq
}

type TxHashes []string

// Merkle takes a list of transaction hashes (hex strings), reverses each one then
// computes the Merkle root which is then reversed.
//
// The reversals are to accomodate the use case of Merkle root of the list of
// hashes of transactions in a Block
//
// It also returns a byte sequence 'skeleton' of the  Merkle root computations which,
// together with the first transaction, can be used to regenerate the root quickly
func Merkle(list TxHashes) (string, []byte, error) {

	// convert the hex strings to bytes, reversed
	bytelist := make([][]byte, len(list))
	for i, s := range list {
		b, err := hex.DecodeString(s)
		if err != nil {
			return "", nil, err
		}
		bytelist[i] = coin.Reverse(b)
	}

	// use private merkleBytes to workon the byte sequences
	resBytes, skeleton, err := merkleBytes(bytelist)
	if err != nil {
		return "", nil, err
	}

	// protocol demands a final reversal
	resStr := fmt.Sprintf("%x", coin.Reverse(resBytes))
	return resStr, skeleton, nil
}

// merkleBytes takes list txs of byte sequences and returns
// the Merkle root. If the list has odd length, the final element is duplicated.
// Replacing successive pairs with the double hash of their concatenation
// yields a new list of half the length. This process is repeated until
// just one 32-bytes slice remains.
func merkleBytes(txs [][]byte) ([]byte, []byte, error) {
	L := len(txs)
	switch {
	case L == 0:
		return nil, nil, errors.New("Cannot handle empty data")
	case L == 1:
		return txs[0], nil, nil
	case L%2 == 1:
		{
			txs = append(txs, txs[L-1])
			L++
		}
	}
	var (
		i, M     int
		skeleton []byte
	)
	for {
		i, M = 0, 0
		for {
			if i == L {
				break
			}
			txs[M] = coin.Hash2(txs[i], txs[i+1])
			i, M = i+2, M+1
			if M == 1 { // extend skeleton byte
				skeleton = append(skeleton, txs[1]...)
			}
		}
		if M == 1 { // exit with the single root
			break
		}
		L = M
		if L%2 == 1 {
			txs[L] = txs[L-1]
			L++
		}
	}
	return txs[0], skeleton, nil
}

// getMerkle takes Skeleton and Coinbase an computes the
// Merkle root (hex). This is properly in Client. Here for testing
func getMerkle(coinbase string, skeleton []byte) (string, error) {
	N := len(skeleton) / 32 	// the number of partial hashes = loops
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
