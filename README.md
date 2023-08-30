# gogpt

This is a simple module for accessing ChatGPT from golang.

## Using

Run a basic query in one line...

```
generated, err := NewGoGPTQuery(OPENAI_KEY).AddMessage(gogpt.ROLE_SYSTEM, "You are a detective.").AddMessage(gogpt.ROLE_USER, "Solve the Great Train Mystery").Generate()
fmt.Printf("Result: %s\n",generated.Choices[0].Message.Content)
```

Just keep adding messages to make a simple chat...

```
gpt := NewGoGPTQuery(OPENAI_KEY)
gpt.AddMessage(gogpt.ROLE_SYSTEM, "You are a detective.").AddMessage(gogpt.ROLE_USER, "Solve the Great Train Mystery.").AddMessage(gogpt.ROLE_ASSISTANT,"Ok! I got this.").AddMessage(gogpt.ROLE_USER,"And hurry!").Generate()
```

## Testing

If you want to test this module, copy the file testconfig-sample.json to testconfig.json and replace the org id and api key with your settings. You can change anything else as well, but you'll need a working API key.

Run ```go test -v``` to see verbose output.

BE AWARE TESTS ARE ON THE LIVE API. You will be using tokens, although max_tokens is set to 100 for each query. Total usage for the test suite is around 1,000 tokens.
