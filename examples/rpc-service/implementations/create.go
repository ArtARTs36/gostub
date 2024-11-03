package implementations

import (
	"context"
	"github.com/artarts36/myproject/contracts"
	"net/http"
)

func (s *StubUserService) Create(ctx context.Context, r *http.Request) (*contracts.Response, error) {
	panic("method StubUserService.Create not implemented")
}
