package player

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Public keys
const (
	AlicePK = 439403
	BobPK   = 545433
)

// Player for dice game
type Player struct {
	Name               string
	PublicKey          int
	A                  int    // the "message" to send to another player
	B                  int    // the random b for sending to opponent
	R                  int    // the random r for hashing
	Commitment         string // the commitment to send to opponent
	OpponentCommitment string // the commitment from the opponent
	OpponentB          int    // the b recieved from the opponent
	OpponentPublicKey  int
}

// RollDice rolls a dice between 1 and 6
func RollDice(seed int64) int {
	rand.Seed(seed)
	return rand.Intn(6) + 1
}

// Server starts a server for the player
func Server(p Player, port string) {
	l, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	c, err := l.Accept()
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		netData, err := bufio.NewReader(c).ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		name, key, data := readMessage(netData, p.OpponentPublicKey)
		switch key {
		case "ROLLING":
			p.OpponentCommitment = string(data)
			rand.Seed(time.Now().UnixNano())
			p.B = rand.Intn(1000000)
			fmt.Printf("Generated B: %v\n", p.B)
		case "B":
			_b, _ := strconv.Atoi(data)
			p.OpponentB = _b
			fmt.Println(fmt.Sprintf("B from %s: %v", name, data))
		case "COMM":
			s := strings.Split(data, "|") // r = s[0], a = s[1]
			isValid := compareCommitment(p, s[0], s[1])
			fmt.Printf("Seed and R sent from %s is: %t\n", name, isValid)

			if isValid {
				a, _ := strconv.Atoi(s[1])
				seed := aXORb(a, p.B)
				opponentRoll := RollDice(int64(seed))
				fmt.Printf("(%s rolls: %v)\n", name, opponentRoll)
			}
		case "GO":
			seed := aXORb(p.A, p.OpponentB)
			roll := RollDice(int64(seed))
			fmt.Println(fmt.Sprintf("I rolled a: %v", roll))
		default:
			fmt.Printf(name + ": " + string(data) + "\n")
		}
	}
}

// Client spawns a client for the player
func Client(p Player, port string) {
	c, err := net.Dial("tcp", "localhost"+port)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Print("READY\n")
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		_text := strings.TrimSpace(string(text))
		switch _text {
		case "INIT":
			rand.Seed(time.Now().UnixNano())
			p.R = rand.Intn(1000000)

			com := getCommitment(p)
			p.Commitment = string(com)
			sendClientMessage(p.Name, "ROLLING: "+string(com), p.PublicKey, c)
		case "SEND":
			sendClientMessage(p.Name, fmt.Sprintf("COMM: %v|%v", p.R, p.A), p.PublicKey, c)
		case "GO":
			sendClientMessage(p.Name, "GO: whatever", p.PublicKey, c)
		default:
			sendClientMessage(p.Name, _text, p.PublicKey, c)
		}
	}
}

func getCommitment(p Player) []byte {
	arr := append([]byte(strconv.Itoa(p.R)), strconv.Itoa(p.A)...)
	sum := sha256.Sum256(arr)

	return sum[:]
}

func compareCommitment(p Player, r, seed string) bool {
	arr := append([]byte(r), seed...)
	sum := sha256.Sum256(arr)
	return p.OpponentCommitment == string(sum[:])
}

func aXORb(a, b int) int {
	return a ^ b
}

func signedMessage(publicKey int, msg string) []byte {
	arr := append([]byte(strconv.Itoa(publicKey)), string(msg)...)
	sum := sha256.Sum256(arr)
	return sum[:]
}

func sendClientMessage(senderName, msg string, pk int, c net.Conn) {
	sum := signedMessage(pk, msg)
	fmt.Fprintf(c, "%s:: %s:: %s"+"\n", senderName, sum, msg)
}

func readMessage(msg string, pk int) (string, string, string) {
	str := strings.TrimSpace(string(msg))
	s := strings.Split(str, ":: ")

	name := s[0]
	checksum := s[1]
	_msg := ""
	if len(s) > 2 {
		_msg = s[2]
	}

	if checksum == string(signedMessage(pk, _msg)) {
		str2 := strings.TrimSpace(string(_msg))
		s2 := strings.Split(str2, ": ")
		if len(s2) == 1 {
			return name, "", s2[0]
		}
		return name, s2[0], s2[1]
	}

	return "", "", ""
}
