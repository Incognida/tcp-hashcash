package receiver

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/Incognida/tcp-hashcash/internal/services/receiver/mocks"
)

const (
	successStream = "useless data\n1:20:240809:localhost:RklwNHVnZ09iSWVqTndyQXJGU21zY3RwcmJna3dXbW5IRExEdm1nVQ==:MjEyNTY1MQ==\n"
)

func TestReceiver_Auth(t *testing.T) {
	tests := []struct {
		name          string
		ipsMock       func() IPsRepository
		hashesMock    func() HashesRepository
		now           func() time.Time
		ttl           time.Duration
		randGenerator func() []byte
		reader        *bytes.Buffer
		writer        *bytes.Buffer
	}{
		{
			name: "success",
			ipsMock: func() IPsRepository {
				m := mocks.NewIPsRepository(t)
				m.EXPECT().InsertIP(mock.Anything, "localhost").Return(nil)
				m.EXPECT().HasIP(mock.Anything, "localhost").Return(true, nil)

				return m
			},

			hashesMock: func() HashesRepository {
				m := mocks.NewHashesRepository(t)
				m.EXPECT().InsertHash(
					mock.Anything,
					[]byte{0, 0, 8, 178, 241, 29, 26, 7, 176, 33, 233, 131, 21, 253, 189, 89, 226, 105, 43, 217},
				).Return(false, nil)

				return m
			},
			now: func() time.Time {
				return time.Date(2024, time.August, 9, 0, 0, 0, 0, time.UTC)
			},
			ttl: time.Hour * 24 * 14,
			randGenerator: func() []byte {
				return []byte("FIp4uggObIejNwrArFSmsctprbgkwWmnHDLDvmgU")
			},
			reader: bytes.NewBufferString(successStream),
			writer: bytes.NewBuffer(nil),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Receiver{
				ipsRepository:    tt.ipsMock(),
				hashesRepository: tt.hashesMock(),
				now:              tt.now,
				ttl:              tt.ttl,
				randGenerator:    tt.randGenerator,
			}

			err := s.Auth(context.Background(), "localhost", tt.reader, tt.writer)
			if err != nil {
				t.Errorf("Receiver.Auth() error = %v", err)
				return
			}
		})
	}
}
