package app_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/aserto-dev/go-sample-project/pkg/cc/config"
	"github.com/aserto-dev/go-sample-project/pkg/testharness"
	"github.com/stretchr/testify/require"
)

func TestInfoEndpoint(t *testing.T) {
	// Arrange
	h := testharness.Setup(t, func(cfg *config.Config) {})
	defer h.Cleanup()
	assert := require.New(t)

	// Act
	client := h.CreateClient()
	url := "https://127.0.0.1:8383/api/v1/info"
	req, err := http.NewRequest("GET", url, bytes.NewBuffer([]byte{}))
	assert.NoError(err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	assert.NoError(err)
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	assert.NoError(err)

	assert.Equal(200, resp.StatusCode)

	type Build struct {
		Version string `json:"version"`
		Commit  string `json:"commit"`
		Date    string `json:"date"`
		Os      string `json:"os"`
		Arch    string `json:"arch"`
	}

	type Response struct {
		System  string `json:"system"`
		Version string `json:"version"`
		Build   Build  `json:"build"`
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()

	var result Response
	err = decoder.Decode(&result)
	assert.NoError(err)

	expected := Response{
		Build: Build{
			Version: "0.0.0",
			Commit:  "undefined",
			Date:    result.Build.Date,
			Os:      result.Build.Os,
			Arch:    result.Build.Arch,
		},
	}

	assert.Equal(expected, result)
}
