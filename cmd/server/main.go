package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Incognida/tcp-hashcash/internal/config"
	hashesRepository "github.com/Incognida/tcp-hashcash/internal/repositories/hashes/postgres"
	ipsRepository "github.com/Incognida/tcp-hashcash/internal/repositories/ips/postgres"
	"github.com/Incognida/tcp-hashcash/internal/server"
	"github.com/Incognida/tcp-hashcash/internal/services/receiver"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	cfg, err := config.ParseConfig()
	if err != nil {
		log.Fatalf("failed to parse config: %v\n", err)
	}

	pgxCfg, err := pgxpool.ParseConfig(fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s default_query_exec_mode=simple_protocol",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName,
	))
	if err != nil {
		log.Fatalf("failed to parse pgx config: %v\n", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		log.Fatalf("failed to create pgx pool: %v\n", err)
	}

	defer pool.Close()

	rcv := receiver.NewReceiver(
		ipsRepository.NewRepository(pool),
		hashesRepository.NewRepository(pool),
		cfg.ZerosCount,
		cfg.TTL,
		time.Now,
		func() []byte {
			return randString(cfg.RandThreshold)
		},
	)

	srv := server.New(rcv, fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), cfg.Quotes)
	go func() {
		if err = srv.Listen(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("failed to listen: %v", err)
		}
	}()

	log.Println("started listen")

	<-ctx.Done()
	log.Println("got interruption signal")
	srv.Stop()

	log.Println("gracefully shut down")
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ123456789")

func randString(n int64) []byte {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return []byte(string(b))
}
