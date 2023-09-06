package gogpt

import (
	"fmt"
)

const (
	// Since token estimates are inexact, how much of a buffer should we leave?
	BUFF_MARGIN = 48
)

/*

	The following helper functions are provided to allow infinite chats.
	The max length for a reply should be set using MAX_TOKENS.
	The initial SYSTEM prompt is immutable.
	The most recent message is too important to summarize.

	This means we have a number of tokens available for context equal to:

	MAX_MODEL_TOKENS - MAX_TOKENS - TokenEstimator(SYSTEM_PROMPT) - TokenEstimator(MOST_RECENT_MESSAGE) = MAX_CONTEXT_SIZE

	If this number is smaller than the number of tokens in the message history, we need to summarize.

	We do this by asking ChatGPT to summarize the message history into a single message of less than MAX_CONTEXT_SIZE tokens.
	This summary is then appended to the SYSTEM prompt at query time and the message history is cleared.

	The summary is stored seperately from the message history, however, since it will need to be updated in future summaries:

	GetSummary(currentSummary, messageHistory) -> newSummary

*/

func MaxQueryTokens(model string) int {
	switch model {
	case MODEL_35_TURBO:
		return 1024
		//TEMP: return 4096
	default:
		return 2048
	}
}

// This is a rough estimate of the number of tokens in a string.
func TokenEstimator(m GoGPTMessage) int {
	return len(m.Content) / 4
}

func NewGoGPTChat(key string) *GoGPTChat {
	return &GoGPTChat{
		Query: NewGoGPTQuery(key),
	}
}

type GoGPTChat struct {
	Query   *GoGPTQuery
	Summary string
}

// A convenience function for method chaining
func (c *GoGPTChat) AddMessage(role string, content string) *GoGPTChat {
	c.Query.AddMessage(role, content)
	return c
}

// A function that encapsulates the query generation method and handles summariation.
func (g *GoGPTChat) Generate() (*GoGPTResponse, error) {

	var messages []GoGPTMessage
	var messages_to_summarize []GoGPTMessage

	total_tokens := MaxQueryTokens(g.Query.Model)
	prompt_tokens := 0
	func_tokens := 0
	context_tokens := 0
	new_message_tokens := 0
	first_sys_msg := true

	for i, msg := range g.Query.Messages {
		if msg.Role == ROLE_SYSTEM && first_sys_msg {
			prompt_tokens += TokenEstimator(msg)
			messages = append(messages, msg)
			first_sys_msg = false
		} else if msg.Role == ROLE_FUNCTION {
			func_tokens += TokenEstimator(msg)
			messages = append(messages, msg)
		} else if i == len(g.Query.Messages)-1 {
			new_message_tokens += TokenEstimator(msg)
		} else {
			context_tokens += TokenEstimator(msg)
			messages_to_summarize = append(messages_to_summarize, msg)
		}
	}

	max_context := total_tokens - prompt_tokens - new_message_tokens - func_tokens - BUFF_MARGIN

	if max_context < g.Query.MaxTokens {
		return nil, fmt.Errorf("not enough tokens left for a reply")
	}

	if max_context < context_tokens {

		summary, err := g.summarize(messages_to_summarize, max_context)

		messages = append(messages, GoGPTMessage{Role: ROLE_SYSTEM, Content: summary})
		messages = append(messages, g.Query.Messages[len(g.Query.Messages)-1])

		if err != nil {
			return nil, err
		}

		g.Query.Messages = messages
	}

	resp, err := g.Query.Generate()

	if err != nil {
		return nil, err
	}

	g.AddMessage(ROLE_ASSISTANT, resp.Choices[0].Message.Content)

	return resp, nil
}

func (c *GoGPTChat) summarize(msgs []GoGPTMessage, max_tokens int) (string, error) {

	if len(msgs) == 0 {
		return "", fmt.Errorf("no messages to summarize")
	}

	q := NewGoGPTQuery(c.Query.Key)
	q.AddMessage(ROLE_SYSTEM, fmt.Sprintf("Summarize the following chat history in less than %d words.", max_tokens))

	for _, msg := range msgs {
		q.AddMessage(msg.Role, msg.Content)
	}

	generated, err := q.Generate()

	if err != nil {
		return "", err
	}

	return generated.Choices[0].Message.Content, nil

}
