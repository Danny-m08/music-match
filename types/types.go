package types

import (
	"encoding/json"
	"github.com/bojanz/currency"
	"math/rand"
	"strings"
	"time"
)

const (
	idLength     = 10
	alphaNumeric = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
)

type Listing struct {
	Price   currency.Amount `json:"price"`
	ID      string          `json:"id"`
	Track   Track           `json:"track"'`
	Created *time.Time      `json:"created"`
	Tx      *Transaction    `json:"transaction,omitempty"`
}

type Track struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

type Transaction struct {
	ID    string //`json:"id"`
	Buyer *User
	Date  time.Time //`json:"timestamp"`
}

func GenerateID() string {
	str := strings.Builder{}
	rand.Seed(time.Now().UnixNano())
	for it := 0; it < idLength; it++ {
		str.WriteByte(alphaNumeric[rand.Intn(len(alphaNumeric))])
	}
	return str.String()
}

func (l *Listing) String() string {
	data, _ := json.Marshal(l)
	l.Created.String()

	return string(data)
}
