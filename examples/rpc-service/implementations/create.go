package implementations

import (
	"context"
	"net/http"

	"github.com/artarts36/myproject/contracts"
)

func (s *StubUserService) Create(ctx context.Context, r *http.Request) (*contracts.Response, error) {
	return any(0).(*contracts.Response), errors.New("is not real method Create")
}
