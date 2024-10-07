package protocol

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type HashCash struct {
	Version    int64
	ZerosCount int64
	DateUnix   int64
	Counter    int64
	IP         string
	Rand       []byte
}

func FromBytes(b []byte) (*HashCash, error) {
	hashCashString := string(b)
	parts := strings.Split(hashCashString, ":")
	if len(parts) != 7 {
		return nil, fmt.Errorf("invalid hashcash format: %s", hashCashString)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid version: %s", parts[0])
	}

	zerosCount, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid zeros count: %s", parts[1])
	}

	at, err := time.Parse("060102", parts[2])
	if err != nil {
		return nil, fmt.Errorf("invalid date: %s", parts[2])
	}

	ip := parts[3]

	rand, err := base64.StdEncoding.DecodeString(parts[5])
	if err != nil {
		return nil, err
	}

	counter, err := decodeBase64Int64(parts[6])
	if err != nil {
		return nil, fmt.Errorf("invalid counter: %s", parts[6])
	}

	return &HashCash{
		Version:    version,
		ZerosCount: zerosCount,
		DateUnix:   at.Unix(),
		Rand:       rand,
		Counter:    counter,
		IP:         ip,
	}, nil
}

func (h *HashCash) ToBytes() []byte {
	return []byte(fmt.Sprintf(
		"%d:%d:%s:%s::%s:%s",
		h.Version,
		h.ZerosCount,
		time.Unix(h.DateUnix, 0).Format("060102"),
		h.IP,
		base64.StdEncoding.EncodeToString(h.Rand),
		base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%d", h.Counter))),
	))
}
func (h *HashCash) SHA1Hash() []byte {
	// this code can be optimized by replacing only the counter of the byte slice to remove allocation
	// but im too lazy for that now

	sh := sha1.New()
	sh.Write(h.ToBytes())

	return sh.Sum(nil)
}

func decodeBase64Int64(s string) (int64, error) {
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return 0, err
	}

	return strconv.ParseInt(string(b), 10, 64)
}

// IsLeadingBitsZero checks if the first ZeroCounts leading bits of a byte slice are zero
// keep in mind that leading zero bit is a bit that is on the left side of the byte (i.e. big-endian)
func (h *HashCash) IsLeadingBitsZero() bool {
	data := h.SHA1Hash()

	// Calculate the number of full bytes and remaining bits
	fullBytes := h.ZerosCount / 8
	remainingBits := h.ZerosCount % 8

	// Check the full bytes
	for i := 0; i < int(fullBytes); i++ {
		if data[i] != 0 {
			return false
		}
	}

	// Check the remaining bits in the next byte
	// e.g. we need 4 bits to be zero, so we need to bitwise AND of the 1111 0000 with the byte,
	// if the result is not zero, then the first 4 bits are not zero
	// 1111 0000 &
	// 0001 1010
	// ---------
	// 0001 0000
	if remainingBits > 0 {
		mask := byte(0xFF << (8 - remainingBits))
		if data[fullBytes]&mask != 0 {
			return false
		}
	}

	return true
}
