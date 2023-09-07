package gogpt

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/invopop/jsonschema"
)

/*
	Based on the the ChatGPT documentation here: https://platform.openai.com/docs/api-reference/chat/object

	There are different token limit for the different models: https://openai.com/pricing
*/

const (
	API_ENDPOINT   = "https://api.openai.com/v1/chat/completions"
	MODEL_35_TURBO = "gpt-3.5-turbo"
	ROLE_SYSTEM    = "system"
	ROLE_USER      = "user"
	ROLE_ASSISTANT = "assistant"
	ROLE_FUNCTION  = "function"
)

type GoGPTFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

/*
	Role is an enum of system, user, assistant, or function.
*/

type GoGPTMessage struct {
	Role         string             `json:"role"`
	Content      string             `json:"content"`
	Name         string             `json:"name,omitempty"`
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

type GoGPTFunction struct {
	Name        string             `json:"name"`
	Description string             `json:"description,omitempty"`
	Parameters  *jsonschema.Schema `json:"parameters"`
}

/*
	Only Key, Model, and Messages are required.
*/

type GoGPTQuery struct {
	Model           string             `json:"model"`
	Messages        []GoGPTMessage     `json:"messages"`
	Functions       []GoGPTFunction    `json:"functions,omitempty"`
	FunctionCall    string             `json:"function_call,omitempty"`
	Temperature     float32            `json:"temperature,omitempty"`
	TopP            float32            `json:"top_p,omitempty"`
	N               int                `json:"n,omitempty"`
	Stream          bool               `json:"stream,omitempty"`
	Stop            string             `json:"stop,omitempty"`
	MaxTokens       int                `json:"max_tokens,omitempty"`
	PresencePenalty float32            `json:"presence_penalty,omitempty"`
	LogitBias       map[string]float32 `json:"logit_bias,omitempty"`
	User            string             `json:"user,omitempty"`
	Key             string             `json:"-"`
	OrgName         string             `json:"-"`
	OrgId           string             `json:"-"`
	Endpoint        string             `json:"-"`
	Timeout         time.Duration      `json:"-"`
}

func NewGoGPTQuery(key string) *GoGPTQuery {

	// Set minimal defaults

	d, _ := time.ParseDuration("30s")

	return &GoGPTQuery{
		Key:         key,
		Endpoint:    API_ENDPOINT,
		Model:       MODEL_35_TURBO,
		Temperature: 0.7,
		MaxTokens:   250,
		Timeout:     d,
	}
}

func (g *GoGPTQuery) AddFunction(name string, desc string, obj interface{}) (*GoGPTQuery, error) {

	fjson := jsonschema.Reflect(obj)
	tname := reflect.TypeOf(obj).Name()

	if tname == "" {
		return nil, fmt.Errorf("could not determine type name")
	}

	f := GoGPTFunction{
		Name:        name,
		Description: desc,
		Parameters:  fjson.Definitions[tname],
	}

	g.Functions = append(g.Functions, f)

	return g, nil
}

func (g *GoGPTQuery) AddMessage(role string, content string) *GoGPTQuery {

	msg := GoGPTMessage{
		Role:    role,
		Content: content,
	}

	g.Messages = append(g.Messages, msg)

	return g
}

func (g *GoGPTQuery) Generate() (*GoGPTResponse, error) {

	client := resty.New()
	client.SetTimeout(g.Timeout)

	if len(g.Messages) == 0 {
		return nil, fmt.Errorf("no messages provided")
	}

	resp, err := client.R().
		SetHeader("Authorization", "Bearer "+g.Key).
		SetHeader("Content-Type", "application/json").
		SetBody(g).
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
		return nil, fmt.Errorf("error: %v", gptResp.Error.Message)
	}

	return gptResp, nil
}
