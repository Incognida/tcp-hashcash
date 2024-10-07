package sender

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/Incognida/tcp-hashcash/internal/pkg/protocol"
)

const (
	uselessMessage = "hi\n"
)

type Sender struct {
	zerosCount int64
	now        func() time.Time
	maxCounter int64
}

func NewSender(
	zerosCount int64,
	now func() time.Time,
	maxCounter int64,
) *Sender {
	return &Sender{
		zerosCount: zerosCount,
		now:        now,
		maxCounter: maxCounter,
	}
}

func (s *Sender) GetQuote(
	r io.Reader,
	w io.Writer,
) (string, error) {
	// initialization with any message
	_, err := w.Write([]byte(uselessMessage))
	if err != nil {
		return "", fmt.Errorf("failed to write useless message: %w", err)
	}

	reader := bufio.NewReader(r)

	// get a challenge from receiver
	challengeBytes, err := reader.ReadBytes('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read challenge: %w", err)
	}

	hashcash, err := protocol.FromBytes(challengeBytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse challenge: %w", err)
	}

	// solve the challenge
	for {
		if hashcash.Counter > s.maxCounter {
			return "", fmt.Errorf("cant find hashcash solution")
		}

		if hashcash.IsLeadingBitsZero() {
			break
		}

		hashcash.Counter++
	}

	log.Printf("found solution for %d iterations\n", hashcash.Counter)

	_, err = w.Write(append(hashcash.ToBytes(), '\n'))
	if err != nil {
		return "", fmt.Errorf("failed to write hashcash solution: %w", err)
	}

	// get the quote
	return reader.ReadString('\n')
}
