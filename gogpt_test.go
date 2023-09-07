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

	generated, err := gpt.AddMessage(ROLE_SYSTEM, "Can pigs fly?").Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
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

	gpt.AddMessage(ROLE_SYSTEM, "Take this game command: 'Walk forward three steps' and make it go into json format. Send a single function call.")
	gpt.AddFunction("get_game_instruction_from_user_input", "Get game instruction from user input", Event{})

	tmp, _ := json.Marshal(gpt)
	t.Logf("RAW: %+v", string(tmp))

	if err != nil {
		t.Errorf("Error building test query: %v", err)
		return
	}

	t.Logf("Query: %+v", gpt)

	resp, err := gpt.Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
		return
	}

	t.Logf("Raw Response: %+v", resp)
	t.Logf("\nJSON:\n %s\nEND\n\n", resp.Choices[0].Message.Content)

	e := new(Event)
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), e)

	if err != nil {
		t.Errorf("Error unmarshalling: %v", err)
		return
	}

	if strings.EqualFold(e.Action, "Walk") {
		t.Logf("It walks!\n")
	} else {
		t.Errorf("Error: %+v", e)
	}
}

func TestGenerateInfiniteChat(t *testing.T) {

	gpt, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("error building test query: %s", err)
	}

	chat1 := NewGoGPTChat(gpt.Key)
	chat2 := NewGoGPTChat(gpt.Key)

	chat1.AddMessage(ROLE_SYSTEM, "You are a bumbling but confident French detective talking to your superintendant.").AddMessage(ROLE_USER, "Solve the great train robbery.")
	chat2.AddMessage(ROLE_SYSTEM, "You are a serious and dour English police superintendant talking to a detective. You want him to solve the great train robbery.")

	t.Logf("GENERATING SEED MESSAGE\n")

	generated, err := chat1.Generate()

	if err != nil {
		t.Errorf("error generating: %s", err)
	}

	t.Logf("GPT1: %s\n\n", generated.Choices[0].Message.Content)
	t.Logf("USAGE: %d\n\n", generated.Usage.TotalTokens)

	for i := 0; i < 10; i++ {

		t.Logf("\n\nITERATION %d\n\n", i)

		generated, err = chat2.AddMessage(ROLE_USER, generated.Choices[0].Message.Content).Generate()

		if err != nil {
			t.Errorf("error generating: %s", err)
		}

		t.Logf("GPT2: %s\n\n", generated.Choices[0].Message.Content)
		t.Logf("USAGE: %d\n\n", generated.Usage.TotalTokens)

		generated, err = chat1.AddMessage(ROLE_USER, generated.Choices[0].Message.Content).Generate()

		if err != nil {
			t.Errorf("error generating: %s", err)
		}

		t.Logf("GPT1: %s\n\n", generated.Choices[0].Message.Content)
		t.Logf("USAGE: %d\n\n", generated.Usage.TotalTokens)
	}

	t.Logf("Completed successfully.")
}
