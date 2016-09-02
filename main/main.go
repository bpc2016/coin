// Package main produces the executable implementing the single server against netcl clients
package main

import (
	"coin/netsrv"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())
	for {
		netsrv.Server([]string{"al", "ph", "bet", "ic"})
		time.Sleep(2 * time.Second)
	}
}
