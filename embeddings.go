package gogpt

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type EmbeddingData struct {
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
	Object    string    `json:"object"`
}

type GoGPTEmbeddings struct {
	Model  string          `json:"model"`
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Usage  GoGPTUsage      `json:"usage"`
}

type GoGPTEmbeddingsRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

func GetEmbedding(input string, key string) (*GoGPTEmbeddings, error) {

	embeddingsReq := GoGPTEmbeddingsRequest{
		Input: input,
		Model: MODEL_EMBEDDING_ADA,
	}

	timeout, _ := time.ParseDuration("30s")

	client := resty.New()

	client.SetTimeout(timeout)

	if len(input) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+key).
		SetHeader("Content-Type", "application/json").
		SetBody(embeddingsReq).
		Post(EMBEDDINGS_ENDPOINT)

	if err != nil {
		return nil, err
	}

	embResp := new(GoGPTEmbeddings)
	err = json.Unmarshal(resp.Body(), &embResp)

	if err != nil {
		return nil, err
	}

	return embResp, nil
}
