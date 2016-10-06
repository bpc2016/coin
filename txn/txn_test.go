package txn

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func piece(a int, b int, v []byte, label string) []byte {
	slice := v[a : b+1]
	fmt.Printf(">> %d - %d\n%s: %v %x\n", a, b, label, slice, slice)
	return slice
}

func littleEndian(v []byte) string {
	length := len(v)
	vReversed := make([]byte, length)
	for i := 0; i < length; i++ {
		vReversed[i] = v[length-i-1]
	}
	str := fmt.Sprintf("%x", vReversed)
	return strings.TrimLeft(str, "0")
}

func TestCoinbaseData(t *testing.T) {
	testbh := 19839016
	testextra := 129456
	testminerid := 261789
	testminer := "an Arbitrarily long name ? again!:"
	cb, err := coinbaseData(testbh, testextra, testminerid, testminer)
	if err != nil {
		t.Error(err)
	}
	fmt.Printf("gets:\n%v\nhex=%x\nlength=%d\n", cb, cb, len(cb))
	p := piece(0, 0, cb, "bh length")
	n := int(p[0]) // either 3 or 4 - bytes length of hex(blockheight)
	p = piece(1, n, cb, "blockheight")
	h1 := littleEndian(p)
	h2 := fmt.Sprintf("%x", testbh)
	if h1 != h2 {
		t.Errorf("wrong blockheight length %s != %s", h1, h2)
	}
	nce := piece(n+1, n+4, cb, "extranonce")
	h1 = littleEndian(nce)
	h2 = fmt.Sprintf("%x", testextra)
	if h1 != h2 {
		t.Errorf("extranonce error %s != %s", h1, h2)
	}
	p = piece(n+5, n+7, cb, "miner id")
	h1 = littleEndian(p)
	h2 = fmt.Sprintf("%x", testminerid)
	if h1 != h2 {
		t.Errorf("miner id error %s != %s", h1, h2)
	}
	p = piece(n+8, n+8+20-1, cb, "miner hash")
	m := []byte(testminer)
	v := make([]byte, 4+len(m)) // to accommodate nonce and the name ...
	copy(v[0:3], nce)
	copy(v[4:], m)
	hv, err := Hash160(v)
	if err != nil {
		t.Error(err)
	}
	h1 = fmt.Sprintf("%x", hv)
	h2 = fmt.Sprintf("%x", p)
	if h1 != h2 {
		t.Errorf("miner hash error %s != %s", h1, h2)
	}
	p = piece(n+28, n+28+9-1, cb, "pool identity")
	h1 = fmt.Sprintf("%x", p)
	h2 = "2f5a6f6368657a612f"
	if h1 != h2 {
		t.Errorf("pool identity error %s != %s", h1, h2)
	}
}

func TestNewRawTransaction(t *testing.T) {
	testInputTx := "3ad337270ac0ba14fbce812291b7d95338c878709ea8123a4d88c3c29efbc6ac"
	testAmount := 65600
	testOutputindex := 0
	testScriptSig := []byte{118, 169, 20, 146, 3, 228, 122, 22, 247, 153, 222, 208, 53, 50, 227, 228, 82, 96, 111, 220, 82, 0, 126, 136, 172}
	testScriptPubKey := []byte{169, 20, 26, 139, 0, 38, 52, 49, 102, 98, 92, 116, 117, 240, 30, 72, 181, 237, 232, 192, 37, 46, 135}
	testRawTx := []byte{1, 0, 0, 0, 1, 172, 198, 251, 158, 194, 195, 136, 77, 58, 18, 168, 158, 112, 120, 200, 56, 83, 217, 183, 145, 34, 129, 206, 251, 20, 186, 192, 10, 39, 55, 211, 58, 0, 0, 0, 0, 25, 118, 169, 20, 146, 3, 228, 122, 22, 247, 153, 222, 208, 53, 50, 227, 228, 82, 96, 111, 220, 82, 0, 126, 136, 172, 255, 255, 255, 255, 1, 64, 0, 1, 0, 0, 0, 0, 0, 23, 169, 20, 26, 139, 0, 38, 52, 49, 102, 98, 92, 116, 117, 240, 30, 72, 181, 237, 232, 192, 37, 46, 135, 0, 0, 0, 0}

	rawTx, err := NewRawTransaction(testInputTx, testAmount, testOutputindex, testScriptSig, testScriptPubKey)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(rawTx, testRawTx) != 0 {
		t.Errorf("Raw transaction different from expected transaction.\n%v\n%v\n", testRawTx, rawTx)
	}

	// L := len(rawTx)

	// piece(0, 3, rawTx)        // version - 4 bytes
	// piece(4, 4, rawTx)        // number - 1 byte
	// piece(5, 32+5-1, rawTx)   // input tx hash - 32 bytes
	// piece(37, 40, rawTx)      // output index (starts at 0) - 4 bytes
	// piece(41, 41, rawTx)      // length of following script = 25 or 0x19 here
	// piece(42, 42+25-1, rawTx) // script sig "unlocking script"
	// piece(67, 67+4-1, rawTx)  // sequence set at 0xffffff -4 bytes
	// piece(71, 71, rawTx)      // number of outputs - 1 byte
	// piece(72, 72+8-1, rawTx)  // satoshis : little endian - 8 bytes
	// piece(80, 80, rawTx)      // length of following script = 23 - 1 (variable)
	// piece(81, 81+23-1, rawTx) // script pub key "locking script"
	// piece(104, 107, rawTx)    // locktime - 4 byte

	// fmt.Printf("Length: %d\nAmount: %x\n", L, testAmount)

	fmt.Printf("rawTX: %x\n", rawTx)
	r := Transaction(rawTx)
	r.detail()
}

/*
func TestCoinbase(t *testing.T) {
	// testInputTx := "0000000000000000000000000000000000000000000000000000000000000000" // coinbase
	// testAmount := 2509094928                                                          // satoshi  25.09094928 BTC
	// testOutputindex := -1
	// testBlockHeight := 277316
	// testExtraNonce := 0x858402
	// testUserName := "busiso"
	// testCoinBaseData, err := coinbaseData(testBlockHeight, testExtraNonce, testUserName)
	// if err != nil {
	// 	t.Error(err)
	// }
	// testScriptPubKey, err := P2PKH("0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086") //hex.DecodeString("2102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac")
	// if err != nil {
	// 	t.Error(err)
	// }

	testRawTx, err := hex.DecodeString("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff1403443b04038584020b757365723a62757369736fffffffff0110c08d95000000001976a914164f1d1d6fce7e2e491352b95b4ea47b880c154688ac00000000")
	if err != nil {
		t.Error(err)
	}

	rawTx, err := NewCoinBase(277316, 9094928, "0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086", 0x858402, "user:busiso")
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("rawtx:\n%x\n", rawTx)

	if !reflect.DeepEqual(rawTx, testRawTx) {
		fmt.Printf("Raw transaction different from expected transaction.\n%v\n%v\n", testRawTx, rawTx)
	}

	cb, err := CoinBase(277316, 9094928, "0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086", 0x858402, "user:busiso")
	if err != nil {
		t.Error(err)
	}

	// fmt.Printf("cb:\n%v\nposition 40 : %v\n", cb, cb.entry(40))
	// l := len(cb) - 6
	// cb.put(l, []byte{1, 2, 3, 4, 5, 6})

	// fmt.Printf("cb:\n%v\nfrom position %d : %v\n", cb, l, cb.get(l, 6))

	cb.detail()
	// res, _ := P2PKH("0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086")

	// fmt.Printf("p2kh :\n%x\n", res)

	//01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff53038349040d00456c69676975730052d8f72ffabe6d6dd991088decd13e658bbecc0b2b4c87306f637828917838c02a5d95d0e1bdff9b0400000000000000002f73733331312f00906b570400000000e4050000ffffffff01bf208795000000001976a9145399c30930af4be1215d59b857b861ad5d88ac00000000

	// cb2, err := hex.DecodeString("01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff53038349040d00456c69676975730052d8f72ffabe6d6dd991088decd13e658bbecc0b2b4c87306f637828917838c02a5d95d0e1bdff9b0400000000000000002f73733331312f00906b570400000000e4050000ffffffff01bf208795000000001976a9145399c3093d31e4b0af4be1215d59b857b861ad5d88ac00000000")
	// if err != nil {
	// 	t.Error(err)
	// }
	// Transaction(cb2).detail()
	//"038abd07062f503253482f048725ee5208083865a409000000092f7374726174756d2f"
	//03799b0600048846f55704f8138d1d0c970f885860bc0100000000000a636b706f6f6c0d2f4b616e6f202f4245424f502f
	// cb2, _ := hex.DecodeString("03789b061d4d696e656420627920416e74506f6f6c20777920397cabee2057f54151da000000e0150200")
	// fmt.Printf("%v\n", unravel(cb2))

	w := []byte("/Zocheza/")
	fmt.Printf("/Zocheza/ =\n%x\nlength=%d", w, len(w))

	//433016//03 789b06 1d4d696e656420627920416e74506f6f6c20777920397cabee2057f54151da000000e0150200
	//433017//03 799b06 00048846 f55704f8 138d1d 0c 970f885860bc010000000000 0a 636b706f6f6c0d2f4b616e6f202f4245424f502f
	//433018//03 7a9b06 2f4254432e434f4d2ffabe6d6d d814995154785cd607ab216a80ebe497a3687ce1d490ebe10c8ad430d21d4d7a 0100000000000000021e b57199c8010000000000
	//433019//03 7b9b06 2f4254432e434f4d2ffabe6d6d cfc9ce9019b508e3738cecc4e3b3f027c3 04db8b3bf4 06326a02c12f84cd2165 0100000000000000021e a31a0bcd000000000000
	//433023//03 7f9b06 0004f551 f55704e6 a21b12 0c 2d8bdc58a91700ac00000000 0a 636b706f6f6c0d2f4b616e6f202f4245424f502f
}
*/
// 15 - length of script sig
// [3 68 59 4 3 133 132 2 6 47 80 50 83 72 47] - scriptsig  itself
// 35 - len scriptpubkey
// [33 2 170 151 12 89 38 64 209 157 224 63 246 243 41 214 253 46 236 176 35 38 59 155 165 209 184 28 41 181 35 218 139 33 172]

// raw: 01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0f03443b0403858402062f503253482fffffffff0110c08d9500000000232102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac00000000
// act: 01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff0f03443b0403858402062f503253482fffffffff0110c08d9500000000232102aa970c592640d19de03ff6f329d6fd2eecb023263b9ba5d1b81c29b523da8b21ac00000000

// from ... http://bitcoin.stackexchange.com/questions/21557/how-to-fully-decode-a-coinbase-transaction

// coinbase: 038abd07062f503253482f048725ee5208083865a409000000092f7374726174756d2f
// 03 - length opcode
// 8abd07 - data with length 03
// 06 - length opcode
// 2f503253482f - data with length 06
// 04 - length opcode
// 8725ee52 - data with length 04
// 08 - length opcode
// 083865a409000000 - data with length 08
// 09 - length opcode
// 2f7374726174756d2f - data with length 09

// scriptPubKey: 76a914975efcba1e058667594dc57146022ec46560a63c88ac
// 76 - OP_DUP opcode
// a9 - HASH160 opcode
// 14 - length opcode
// 975efcba1e058667594dc57146022ec46560a63c - data with length 14 (20 in dec)
// 88 - OP_EQUALVERIFY opcode
// ac - OP_CHECKSIG opcode

// 132xe93LdrdGa39vN7su1shRpcBwMdAX4J - wif
// 0225c141d69b74adac8ab984a8eb9fee42c4ce79cf6cb2be166b1ddc0356b37086 - pubkey
// 164f1d1d6fce7e2e491352b95b4ea47b880c1546 - after Hash160
// KyufBz2L22mZgxgeftJuDK7Fot4rMarX4sQ7v5SNE9eZhq1wSqVf - privkey
