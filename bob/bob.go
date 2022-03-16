package main

import (
	"math/rand"
	"time"

	"github.com/xnoga/sec_assignment2/player"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	// Bob "chooses" a random seed with which he wants to throw a dice with
	bob := player.Player{
		Name:              "Bob",
		B:                 rand.Intn(1000000),
		PublicKey:         player.BobPK,
		OpponentPublicKey: player.AlicePK,
	}

	go player.Server(bob, ":8081")
	player.Client(bob, ":8080")
}
