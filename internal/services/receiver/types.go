package receiver

import "context"

//go:generate mockery --name=IPsRepository --with-expecter --case=underscore
type IPsRepository interface {
	HasIP(ctx context.Context, ip string) (bool, error)
	InsertIP(ctx context.Context, ip string) error
}

//go:generate mockery --name=HashesRepository --with-expecter --case=underscore
type HashesRepository interface {
	InsertHash(ctx context.Context, hash []byte) (bool, error)
}
