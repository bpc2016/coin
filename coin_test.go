package coin

import (
	"fmt"
	"testing"
)

func TestDoubleSHA256(t *testing.T) {
	testbytes := []byte("teststring")
	expected := "85e63514d692e0136925bec920b4ac50c297c36774587c2ec2b86e8075001e73"

	hash, err := DoubleSha256(testbytes)
	if err != nil {
		t.Error(err)
	}
	hashStr := fmt.Sprintf("%x", hash)
	if hashStr != expected {
		t.Errorf("expected %s but got %s", expected, hash)
	}
}
