package server

import (
	"bytes"
	"cernunnos/internal/pkg/dto"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

func TestBuildRequest(t *testing.T) {
	request := map[string]any{
		"ids":              []string{"123", "234", "345"},
		"with_busy":        true,
		"with_unavailable": true,
		"limit":            5,
	}

	data, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	r := &http.Request{
		Body: io.NopCloser(bytes.NewBuffer(data)),
	}

	req, err := buildRequest[dto.StorageProductsRequest](r)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(req)
}
