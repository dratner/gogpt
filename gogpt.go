package gogpt

import (
	"encoding/json"
	"fmt"

	"github.com/go-resty/resty/v2"
)

/*
	Based on the the ChatGPT documentation here: https://platform.openai.com/docs/api-reference/chat/object
*/

type GoGPTFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type GoGPTMessage struct {
	Role         string             `json:"role,omitempty"`
	Content      string             `json:"content"`
	FunctionCall *GoGPTFunctionCall `json:"function_call,omitempty"`
}

type GoGPTChoice struct {
	Index        int          `json:"index"`
	Message      GoGPTMessage `json:"message"`
	FinishReason string       `json:"finish_reason"`
}

type GoGPTUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GoGPTError struct {
	Message string `json:"message"`
	ErrType string `json:"type"`
	Param   string `json:"param"`
	Code    string `json:"code"`
}

type GoGPTResponse struct {
	Error   *GoGPTError   `json:"error,omitempty"`
	Id      string        `json:"id"`
	Object  string        `json:"object"`
	Created int32         `json:"created"`
	Model   string        `json:"model"`
	Choices []GoGPTChoice `json:"choices"`
	Usage   GoGPTUsage    `json:"usage"`
}

type GoGPTQuery struct {
	Model       string         `json:"model"`
	Temperature float32        `json:"temperature"`
	Messages    []GoGPTMessage `json:"messages"`
}

type GoGPT struct {
	Key         string  `json:"gptkey"`
	OrgName     string  `json:"gptorgname"`
	OrgId       string  `json:"gptorgid"`
	Endpoint    string  `json:"gptendpoint"`
	Model       string  `json:"gptmodel"`
	User        string  `json:"gptuser"`
	Role        string  `json:"gptrole"`
	Temperature float32 `json:"gpttemperature"`
}

func NewGoGPT(key string, orgName string, orgId string, endpoint string, model string, user string, role string, temperature float32) *GoGPT {
	return &GoGPT{
		Key:         key,
		OrgName:     orgName,
		OrgId:       orgId,
		Endpoint:    endpoint,
		Model:       model,
		User:        user,
		Role:        role,
		Temperature: temperature,
	}
}

func (g *GoGPT) Generate(prompt string) (*GoGPTResponse, error) {
	client := resty.New()

	msg := GoGPTMessage{
		Role:         g.Role,
		Content:      prompt,
		FunctionCall: nil,
	}

	query := GoGPTQuery{
		Model:       g.Model,
		Temperature: g.Temperature,
		Messages: []GoGPTMessage{
			msg,
		},
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+g.Key).
		SetHeader("Content-Type", "application/json").
		SetBody(query).
		Post(g.Endpoint)
	if err != nil {
		return nil, err
	}

	gptResp := new(GoGPTResponse)
	err = json.Unmarshal(resp.Body(), &gptResp)

	if err != nil {
		return nil, err
	}

	if gptResp.Error != nil {
		return nil, fmt.Errorf("Error: %v", gptResp.Error.Message)
	}

	return gptResp, nil
}
