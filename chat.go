package gogpt

import (
	"fmt"

	"github.com/pkoukk/tiktoken-go"
)

const (
	// Since token estimates are inexact, how much of a buffer should we leave?
	BUFF_MARGIN = 48
)

/*

	The following helper functions are provided to allow infinite chats.
	The max length for a reply should be set using MAX_TOKENS.
	The initial SYSTEM prompt is immutable.
	The most recent messages are too important to summarize.

	This means we have a number of tokens available for context equal to:

	MAX_MODEL_TOKENS - MAX_TOKENS - TokenEstimator(SYSTEM_PROMPT) - TokenEstimator(MOST_RECENT_MESSAGES) = MAX_CONTEXT_SIZE

	If this number is larger than the number of tokens the model supports, we need to summarize.

	We do this by asking ChatGPT to summarize the message history into a single message of less than MAX_TOKENS tokens.
	A new message history is then generated with the initial prompt, the summary, and the queue of new messages.
*/

func MaxQueryTokens(model string) int {
	switch model {
	case MODEL_4:
		return 8192
	case MODEL_35_TURBO:
		return 16385
	default:
		return 16385
	}
}

// This is an estimate of the number of tokens in a string.
func TokenEstimator(msg GoGPTMessage, model string) int {

	tkm, err := tiktoken.EncodingForModel(model)

	if err != nil {
		return 0
	}

	// encode
	token := tkm.Encode(msg.Content, nil, nil)

	return len(token)
}

func NewGoGPTChat(key string) *GoGPTChat {
	return &GoGPTChat{
		Query: NewGoGPTQuery(key),
	}
}

type GoGPTChat struct {
	Query        *GoGPTQuery
	Summary      string
	MessageQueue []GoGPTMessage
	prompt       *GoGPTMessage
}

// A convenience function for method chaining
func (c *GoGPTChat) AddMessage(role string, name string, content string) *GoGPTChat {

	msg := GoGPTMessage{
		Role:    role,
		Content: content,
		Name:    name,
	}

	c.MessageQueue = append(c.MessageQueue, msg)

	return c
}

func (g *GoGPTChat) summarize(queueSize int) error {

	// if we don't already know the prompt, find it
	if g.prompt == nil {
		for _, msg := range g.Query.Messages {
			if msg.Role == ROLE_SYSTEM {
				g.prompt = &msg
				break
			}
		}
	}

	// if we can't find the prompt, we can't summarize
	if g.prompt == nil {
		return fmt.Errorf("no prompt found")
	}

	promptSize := TokenEstimator(*g.prompt, g.Query.Model)

	// make sure the prompt, the summary, the queue, and a return message will fit
	if (g.Query.MaxTokens + queueSize + promptSize + BUFF_MARGIN) > MaxQueryTokens(g.Query.Model) {
		return fmt.Errorf("not enough tokens to summarize")
	}

	q := NewGoGPTQuery(g.Query.Key)

	for _, msg := range g.Query.Messages {
		if &msg != g.prompt {
			q.AddMessage(msg.Role, msg.Name, msg.Content)
		}
	}

	q.AddMessage(ROLE_SYSTEM, "", fmt.Sprintf("Summarize the following chat history. You must use less than %d words.", g.Query.MaxTokens))
	q.MaxTokens = g.Query.MaxTokens

	resp, err := q.Generate()

	if err != nil {
		return err
	}

	g.Query.Messages = []GoGPTMessage{}
	g.Query.AddMessage(ROLE_SYSTEM, "", g.prompt.Content)
	g.Query.AddMessage(ROLE_SYSTEM, "", resp.Choices[0].Message.Content)

	return nil
}

// A function that encapsulates the query generation method and handles summariation.
func (g *GoGPTChat) Generate() (*GoGPTResponse, error) {

	var err error

	usage := g.Query.MaxTokens + BUFF_MARGIN // the maximum size of the return message plus a buffer

	for _, msg := range g.Query.Messages {
		usage += TokenEstimator(msg, g.Query.Model)
	}

	queueSize := 0
	for _, msg := range g.MessageQueue {
		queueSize += TokenEstimator(msg, g.Query.Model)
	}
	usage += queueSize

	if usage > MaxQueryTokens(g.Query.Model) {
		err = g.summarize(queueSize)
		if err != nil {
			return nil, err
		}
	}

	g.Query.Messages = append(g.Query.Messages, g.MessageQueue...)
	g.MessageQueue = []GoGPTMessage{}

	resp, err := g.Query.Generate()

	if err != nil {
		return nil, err
	}

	g.Query.AddMessage(ROLE_ASSISTANT, "", resp.Choices[0].Message.Content)

	return resp, nil
}
