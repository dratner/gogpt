package gogpt

import (
	"fmt"
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
		return 4096
	default:
		return 2048
	}
}

// This is a rough estimate of the number of tokens in a string.
func TokenEstimator(s string) int {
	return len(s) / 4
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

// A converience function so you don't have to call Summarize() yourself.
func (c *GoGPTChat) AddMessage(role string, content string) *GoGPTQuery {
	_ = c.Summarize()
	return c.Query.AddMessage(role, content)
}

// A convenience function for method chaining.
func (g *GoGPTChat) Generate() (*GoGPTResponse, error) {
	return g.Query.Generate()
}

func (c *GoGPTChat) Summarize() error {

	fmt.Println("Summarizing...")

	if len(c.Query.Messages) == 0 {
		fmt.Println("No messages to summarize.")
		return nil
	}

	history := c.Summary + "\n\n"

	var messages []GoGPTMessage
	tokens_system, tokens_chat, tokens_max := 0, 0, 0

	tokens_max = MaxQueryTokens(c.Query.Model)

	for _, msg := range c.Query.Messages {
		if msg.Role == ROLE_SYSTEM {
			messages = append(messages, msg)
			tokens_system += TokenEstimator(msg.Content)
		} else {
			history += msg.Role + ": " + msg.Content + "\n"
			tokens_chat += TokenEstimator(msg.Content)
		}
	}

	// If we have enough tokens left, no need to summarize. Just return.
	if tokens_max-tokens_system-tokens_chat > c.Query.MaxTokens {
		fmt.Println("No need to summarize.")
		return nil
	}

	q := NewGoGPTQuery(c.Query.Key)
	generated, err := q.AddMessage(ROLE_SYSTEM, "Summarize the following chat history in less than 500 words.\n\n"+history).Generate()

	if err != nil {
		return err
	}

	fmt.Printf("Summarizing chat history...\n")

	c.Summary = generated.Choices[0].Message.Content
	messages = append(messages, GoGPTMessage{Role: ROLE_SYSTEM, Content: c.Summary})
	c.Query.Messages = messages

	fmt.Printf("History: %s\n", history)
	fmt.Printf("Summary: %s\n", c.Summary)

	return nil
}
