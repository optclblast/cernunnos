package server

import (
	"cernunnos/internal/pkg/dto"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

func buildRequest[R dto.Request](r *http.Request) (*R, error) {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("error read request body. %w", err)
	}

	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			err = errors.Join(fmt.Errorf("error close request body. %w", closeErr), err)
		}
	}()

	var request R
	if err = json.Unmarshal(rawBody, &request); err != nil {
		return nil, fmt.Errorf("error unmarshal request body. %w", err)
	}

	return &request, nil
}
