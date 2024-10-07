package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os/signal"
	"syscall"
	"time"

	"github.com/Incognida/tcp-hashcash/internal/config"
	"github.com/Incognida/tcp-hashcash/internal/services/sender"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to parse config: %v\n", err)
	}

	s := sender.NewSender(cfg.ZerosCount, time.Now, cfg.ClientMaxCounter)

	ticker := time.NewTicker(cfg.ClientDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err = do(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), s); err != nil {
				log.Printf("client error, addr: (%s:%d): %v\n", cfg.Host, cfg.Port, err)
			}
		}
	}
}

func do(address string, s *sender.Sender) error {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return fmt.Errorf("failed at dial: %w", err)
	}

	defer conn.Close()

	quote, err := s.GetQuote(conn, conn)
	if err != nil {
		return fmt.Errorf("failed to get quote: %w", err)
	}

	log.Printf("quote result: %s\n", quote)

	return nil
}
