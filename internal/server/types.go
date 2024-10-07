package server

import (
	"context"
	"io"
)

type Receiver interface {
	Auth(
		ctx context.Context,
		ip string,
		r io.Reader,
		w io.Writer,
	) error
}
