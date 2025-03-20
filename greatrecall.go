package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
)

type AddressRequest struct {
	Value     string `json:"value"`
	CreatedAt string `json:"created_at"`
}

type AddressResponse struct {
	StandardAddress string `json:"standardAddress"`
}

func standardAddress(input string) (string, error) {
	url := os.Getenv("GREATRECALL_API_URL")
	if url == "" {
		return input, errors.New("GREATRECALL_API_URL not found")
	}
	key := os.Getenv("GREATRECALL_API_KEY")
	if key == "" {
		return input, errors.New("GREATRECALL_API_KEY not found")
	}

	timestamp := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	text := fmt.Sprintf("%s:%s", key, timestamp)
	token := sha256.Sum256([]byte(text))
	saltedToken := fmt.Sprintf("%x", token)

	reqBody := AddressRequest{
		Value:     input,
		CreatedAt: timestamp,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return input, errors.Wrap(err, "marshal request body")
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return input, errors.Wrap(err, "create request")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", saltedToken))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return input, errors.Wrap(err, "send request")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return input, errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var result AddressResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return input, errors.Wrap(err, "decode response")
	}

	if result.StandardAddress == "" {
		return input, errors.New("no standard address found")
	}
	return result.StandardAddress, nil
}
