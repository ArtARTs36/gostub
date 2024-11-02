package contracts

import (
	"context"
	"net/http"
)

type Response struct {
	Status int
	Body   []byte
}

type UserService interface {
	List(ctx context.Context, r *http.Request) (*Response, error)
	Create(ctx context.Context, r *http.Request) (*Response, error)
}
