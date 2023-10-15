package gogpt

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
)

type TestConfig struct {
	GptKey     string `json:"gpt_key"`
	GptOrgName string `json:"gpt_org_name"`
	GptOrgId   string `json:"gpt_org_id"`
}

/*
	Simple helper function to build a test query.
*/

func buildTestQueryHelper() (*GoGPTQuery, error) {

	conf := new(TestConfig)

	// Try to pull the config vars from the environment.

	conf.GptKey = os.Getenv("OPENAI_KEY")
	conf.GptOrgId = os.Getenv("OPENAI_ORG_ID")
	conf.GptOrgName = os.Getenv("OPENAI_ORG_NAME")

	// If that fails, try to pull them from a file. Use key to test.
	if len(conf.GptKey) == 0 {

		f := "./testconfig.json"
		file, err := os.ReadFile(f)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal([]byte(file), &conf)

		if err != nil {
			return nil, err
		}
	}

	gpt := NewGoGPTQuery(conf.GptKey)
	gpt.OrgName = conf.GptOrgName
	gpt.OrgId = conf.GptOrgId
	gpt.MaxTokens = 100

	return gpt, nil
}

/*
This is the simplest test - just a round-trip to the API.
*/

func TestGenerate(t *testing.T) {

	gpt, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("Error building test query: %v", err)
	}

	t.Logf("Query: %+v", gpt)

	generated, err := gpt.AddMessage(ROLE_SYSTEM, "", "Can pigs fly?").Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
		return
	}

	t.Logf("Generated: %+v", generated)
}

func TestGenerateWithFunctions(t *testing.T) {

	type Event struct {
		Action    string `json:"action"`
		Direction string `json:"direction"`
		Distance  string `json:"distance"`
	}

	gpt, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("Error building test query: %v", err)
		return
	}

	gpt.AddMessage(ROLE_SYSTEM, "", "Interpret user input as game commands. If the user wants to do something, call the appropriate function.")
	gpt.AddMessage(ROLE_USER, "", "Walk forward three steps.")

	gpt.AddFunction("get_game_instruction_from_user_input", "Get game instruction from user input", Event{})

	if err != nil {
		t.Errorf("Error building test query: %v", err)
		return
	}

	t.Logf("Query: %+v", gpt)

	resp, err := gpt.Generate()

	if err != nil {
		t.Errorf("error generating: %v", err)
		return
	}

	t.Logf("Raw Response: %+v", resp)

	reply := resp.Choices[0].Message
	e := new(Event)

	if reply.FunctionCall != nil {
		t.Logf("\nFunction Call: %+v\n", reply.FunctionCall)
		err = json.Unmarshal([]byte(reply.FunctionCall.Arguments), e)
		if err != nil {
			t.Errorf("could not unmarshal: %v", err)
			return
		}
	} else {
		t.Errorf("no function to decode")
		return
	}

	if strings.EqualFold(e.Action, "Walk") {
		t.Logf("It walks!\n")
	} else {
		t.Errorf("Error: %+v", e)
		return
	}
}

func TestGenerateInfiniteChat(t *testing.T) {

	gpt, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("error building test query: %s", err)
	}

	chat1 := NewGoGPTChat(gpt.Key)
	chat1.Query.OrgName = gpt.OrgName
	chat1.Query.OrgId = gpt.OrgId
	chat1.Query.MaxTokens = 100

	chat1.AddMessage(ROLE_SYSTEM, "", "You are a bumbling but confident French detective talking to your superintendant.").AddMessage(ROLE_USER, "", "Solve the great train robbery.")

	chat2 := NewGoGPTChat(gpt.Key)
	chat2.Query.OrgName = gpt.OrgName
	chat2.Query.OrgId = gpt.OrgId
	chat2.Query.MaxTokens = 100

	chat2.AddMessage(ROLE_SYSTEM, "", "You are a serious and dour English police superintendant talking to a detective. You want him to solve the great train robbery.")

	t.Logf("generating seed message")

	generated, err := chat1.Generate()

	if err != nil {
		t.Errorf("error generating: %s", err)
		return
	}

	t.Logf("GPT1: %s\n\n", generated.Choices[0].Message.Content)
	t.Logf("USAGE: %d\n\n", generated.Usage.TotalTokens)

	for i := 0; i < 3; i++ {

		t.Logf("\n\nITERATION %d\n\n", i)

		generated, err = chat2.AddMessage(ROLE_USER, "", generated.Choices[0].Message.Content).Generate()

		if err != nil {
			t.Errorf("error generating: %s", err)
			return
		}

		t.Logf("GPT2: %s\n\n", generated.Choices[0].Message.Content)
		t.Logf("USAGE: %d\n\n", generated.Usage.TotalTokens)

		generated, err = chat1.AddMessage(ROLE_USER, "", generated.Choices[0].Message.Content).Generate()

		if err != nil {
			t.Errorf("error generating: %s", err)
			return
		}

		t.Logf("GPT1: %s\n\n", generated.Choices[0].Message.Content)
		t.Logf("USAGE: %d\n\n", generated.Usage.TotalTokens)
	}

	t.Logf("Completed successfully.")
}
