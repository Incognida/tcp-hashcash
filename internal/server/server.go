package server

import (
	"context"
	"log"
	"math/rand"
	"net"
	"sync"
)

type Server struct {
	receiver Receiver
	addr     string
	quotes   [][]byte
	quit     chan struct{}
	wg       *sync.WaitGroup
	listener net.Listener
}

func New(receiver Receiver, addr string, quotes []string) *Server {
	quotesAsBytes := make([][]byte, 0, len(quotes))
	for i := range quotes {
		quotesAsBytes = append(quotesAsBytes, []byte(quotes[i]))
	}

	return &Server{
		receiver: receiver,
		addr:     addr,
		quotes:   quotesAsBytes,
		quit:     make(chan struct{}),
		wg:       &sync.WaitGroup{},
	}
}

func (s *Server) Listen(ctx context.Context) error {
	listener, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}

	s.listener = listener
	s.wg.Add(1)

	go s.serve(ctx)

	return nil
}

func (s *Server) serve(ctx context.Context) {
	defer s.wg.Done()

	log.Println("im serving and i know it")
	for {
		// Listen for an incoming connection.
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.quit:
				return
			default:
				log.Printf("accept error: %s\n", err)
			}
		} else {
			log.Printf("accepted connection from %s\n", conn.RemoteAddr().String())
			s.wg.Add(1)
			go func() {
				s.handleConn(ctx, conn)
				s.wg.Done()
			}()
		}
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	log.Printf("new connection from %s\n", conn.RemoteAddr().String())
	defer conn.Close()

	host, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		log.Printf("failed to split host and port: %s\n", err)
		return
	}

	err = s.receiver.Auth(ctx, host, conn, conn)
	if err != nil {
		log.Printf("auth error: %s\n", err)
		return
	}

	_, err = conn.Write(append(s.quotes[rand.Intn(len(s.quotes))], '\n'))
	if err != nil {
		log.Printf("give quote err: %s\n", err)
		return
	}

	log.Println("successfully sent quote")
}

func (s *Server) Stop() {
	close(s.quit)
	s.listener.Close()
	s.wg.Wait()
}
