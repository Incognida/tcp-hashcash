package receiver

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"time"

	"github.com/Incognida/tcp-hashcash/internal/pkg/protocol"
)

type Receiver struct {
	ipsRepository    IPsRepository
	hashesRepository HashesRepository
	zerosCount       int64
	now              func() time.Time
	ttl              time.Duration
	randGenerator    func() []byte
}

func NewReceiver(
	ipsRepository IPsRepository,
	hashesRepository HashesRepository,
	zerosCount int64,
	ttl time.Duration,
	now func() time.Time,
	randGenerator func() []byte,
) *Receiver {
	return &Receiver{
		ipsRepository:    ipsRepository,
		hashesRepository: hashesRepository,
		zerosCount:       zerosCount,
		now:              now,
		ttl:              ttl,
		randGenerator:    randGenerator,
	}
}

func (s *Receiver) Auth(
	ctx context.Context,
	ip string,
	r io.Reader,
	w io.Writer,
) error {
	reader := bufio.NewReader(r)

	// ignore the first message, it is just initialization
	_, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	// choose a challenge
	challenge, err := s.ChooseChallenge(ctx, ip)
	if err != nil {
		return err
	}

	_, err = w.Write(append(challenge.ToBytes(), '\n'))
	if err != nil {
		return err
	}

	// verify message
	answer, err := reader.ReadString('\n')
	if err != nil {
		return err
	}

	hashCash, err := protocol.FromBytes([]byte(answer))
	if err != nil {
		return err
	}

	return s.VerifyChallenge(ctx, hashCash, ip)
}

func (s *Receiver) ChooseChallenge(ctx context.Context, ip string) (*protocol.HashCash, error) {
	if err := s.ipsRepository.InsertIP(ctx, ip); err != nil {
		return nil, err
	}

	return &protocol.HashCash{
		Version:    1,
		ZerosCount: s.zerosCount,
		DateUnix:   s.now().Unix(),
		Rand:       s.randGenerator(),
		Counter:    0,
		IP:         ip,
	}, nil
}

// VerifyChallenge verifies the challenge according these rules: https://en.wikipedia.org/wiki/Hashcash#Recipient's_side
func (s *Receiver) VerifyChallenge(
	ctx context.Context,
	challenge *protocol.HashCash,
	ip string,
) error {
	if !challenge.IsLeadingBitsZero() {
		return fmt.Errorf("invalid hashcash: %+v", challenge)
	}

	if s.now().Sub(time.Unix(challenge.DateUnix, 0)) > s.ttl {
		return fmt.Errorf("date is expired: %d", challenge.DateUnix)
	}

	if challenge.IP != ip {
		return fmt.Errorf("invalid ip: %s", challenge.IP)
	}

	has, err := s.ipsRepository.HasIP(ctx, challenge.IP)
	if err != nil {
		return err
	}

	if !has {
		return fmt.Errorf("ip is not registered: %s", challenge.IP)
	}

	inserted, err := s.hashesRepository.InsertHash(ctx, challenge.SHA1Hash())
	if err != nil {
		return err
	}

	if !inserted {
		return fmt.Errorf("malicious attemp to re-use the hash string: %s", challenge.SHA1Hash())
	}

	return nil
}
