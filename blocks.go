// Package coin implemenets bitcoin mining - the blocks file
// contains blockheader material
package coin

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

const (
	mrposition    = 36 // start of merkle root in block
	nonceposition = 76 // start of nonce
)

/*
original idea was a task object:
	return coin.Task{
		BH:       buffer.Bytes(),    // 80 byte slice
		MH:       merkle,            // 32 byte slice
		Target:   targetBytes(Bits), // 32 byte slice
		Share:    share(Share),      // 32 byte slice
		CoinBase: CoinBase,          // string
		Skeleton: skeleton,          // slice of 32*n bytes
	}, nil
} */

// BlockHeader returns a template 80 byte blockheader from version/prevblockhash/timestamp/bits
func BlockHeader(Version int, PrevBlock string, TimeStamp int, Bits int) (Block, error) {
	//Version field
	version := make([]byte, 4)
	binary.LittleEndian.PutUint32(version, uint32(Version))
	// previous block hash
	prevblock := make([]byte, 32)
	temp, err := hex.DecodeString(PrevBlock)
	if err != nil {
		return nil, err
	}
	prevblock = Reverse(temp)
	// merkle - blank for now
	merkle := make([]byte, 32)
	// time - unix timestamp NOTE uint32(time.Now().Unix())
	time := make([]byte, 4)
	binary.LittleEndian.PutUint32(time, uint32(TimeStamp))
	// bits (target)
	bits := make([]byte, 4)
	binary.LittleEndian.PutUint32(bits, uint32(Bits))
	// nonce - initially 0
	nonce := make([]byte, 4)

	var buffer bytes.Buffer
	buffer.Write(version)   // 4
	buffer.Write(prevblock) // 32
	buffer.Write(merkle)    // 32
	buffer.Write(time)      // 4
	buffer.Write(bits)      // 4
	buffer.Write(nonce)     // 4

	return Block(buffer.Bytes()), nil
}

// Bits2Target converts uint32 bits to a 32-byte sequence target
// which is compared to block hashes. A
// target is given by m*2**(8*(r-3)) where bits = r|m, r occupying
// the top byte and m the lower 3 bytes of this uint32.
func Bits2Target(bits uint32) []byte {
	r := bits >> 24           // top byte
	m := (bits << 8) >> 8     // remaining three bytes
	b := make([]byte, 32)     // expect 32 byte result but item is shorter
	mBytes := make([]byte, 4) // receives m
	pos := 32 - r
	binary.LittleEndian.PutUint32(mBytes, m)
	copy(b[pos:pos+3], mBytes[1:])
	return b
}

// ShareTarget  returns a 32 byte sequence with k leading 0's
// and the rest of the elements 0xff as a challenge that is
// easier than the actual target
func ShareTarget(k int) []byte {
	byteseq := make([]byte, k)
	for i := 0; i < 32-k; i++ {
		byteseq = append(byteseq, 0xff)
	}
	return byteseq
}

// Merkle computes the merkleroot hash and a skeleton from coinbase and transactions list
func Merkle(coinbase string, txns []string) ([]byte, []byte, error) {
	txes := append([]string{coinbase}, txns...)
	return merKle(txes)
}

// merKle takes a list of transaction hashes (hex strings), reverses each one then
// computes the Merkle root as a byte sequence.
//
// The reversals are to accomodate the use case of Merkle root of the list of
// hashes of transactions in a Block
//
// It also returns a byte sequence 'skeleton' of the  Merkle root computations which,
// together with the first transaction, can be used to regenerate the root quickly
func merKle(list []string) ([]byte, []byte, error) {
	// convert the hex strings to bytes, reversed
	bytelist := make([][]byte, len(list))
	for i, s := range list {
		b, err := hex.DecodeString(s)
		if err != nil {
			return nil, nil, err
		}
		bytelist[i] = Reverse(b)
	}
	// use private merkleBytes to work on the byte sequences
	resBytes, skeleton, err := merkleBytes(bytelist)
	if err != nil {
		return nil, nil, err
	}
	// protocol demands a final reversal
	mHashBytes := Reverse(resBytes)
	return mHashBytes, skeleton, nil
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
			txs[M] = Hash2(txs[i], txs[i+1])
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

//TODO - below should use []byte , noyt string for coinbase

// Skel2Merkle takes Skeleton and Coinbase an computes the Merkle root (in hex)
func Skel2Merkle(coinbase string, skeleton []byte) ([]byte, error) {
	N := len(skeleton) / 32 // the number of partial hashes = loops
	cbBytes, err := hex.DecodeString(coinbase)
	if err != nil {
		return nil, err
	}
	// we need to reverse coinbase
	part := Reverse(cbBytes)
	// then climb up the tree
	for i := 0; i < N; i++ {
		part = Hash2(part, skeleton[32*i:32*(i+1)])
	}
	root := Reverse(part)
	return root, nil
}

// Skeleton produces just the skeleton - NO coinbase involved given list txns of hashed txns in hex
func Skeleton(txns []string) ([]byte, error) {
	any := "0000000000000000000000000000000000000000000000000000000000000000" // 32 bytes of 0
	txes := append([]string{any}, txns...)
	_, b, err := merKle(txes) // we dont want the bogus merkle root
	return b, err
}

// below we use the coin.Block type [really 80 bytes] byte sequence

// AddMerkle is used to dynamicaly load new Merkkle hashes
func (b Block) AddMerkle(mrhash []byte) error {
	if len(mrhash) != 32 {
		return errors.New("wrong merkle root hash size")
	}
	copy(b[mrposition:], mrhash)
	return nil
}

// PutNonce sets the block's nonce
func (b Block) PutNonce(num uint32) {
	binary.LittleEndian.PutUint32(b[nonceposition:], num)
}
