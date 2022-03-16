package main

import (
	"math/rand"
	"time"

	"github.com/xnoga/sec_assignment2/player"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Alice "chooses" a random seed with which she wants to throw a dice with
	alice := player.Player{
		Name:              "Alice",
		A:                 rand.Intn(1000000),
		PublicKey:         player.AlicePK,
		OpponentPublicKey: player.BobPK,
	}

	go player.Server(alice, ":8080")
	time.Sleep(3 * time.Second)
	player.Client(alice, ":8081")
}
