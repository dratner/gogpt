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

	f := "./testconfig.json"
	file, err := os.ReadFile(f)

	if err != nil {
		return nil, err
	}

	conf := new(TestConfig)
	err = json.Unmarshal([]byte(file), &conf)

	if err != nil {
		return nil, err
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

	gpt.AddMessage(ROLE_SYSTEM, "Take this game command: 'Walk forward three steps' and make it go into json format.")
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

func TestGenerateChat(t *testing.T) {

	gpt1, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("Error building test query: %v", err)
	}

	gpt2, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("Error building test query: %v", err)
	}

	generated, err := gpt1.AddMessage(ROLE_SYSTEM, "You are a bumbling but confident French detective talking to your superintendant.").AddMessage(ROLE_USER, "Can you solve the great train robbery?").Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
	}

	t.Logf("GPT1: %s\n", generated.Choices[0].Message.Content)

	generated, err = gpt2.AddMessage(ROLE_SYSTEM, "You are a serious and dour English police superintendant talking to a detective. You asked the detective to solve the great train robbery.").AddMessage(ROLE_USER, generated.Choices[0].Message.Content).Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
	}

	t.Logf("GPT2: %s\n", generated.Choices[0].Message.Content)

	generated, err = gpt1.AddMessage(ROLE_USER, generated.Choices[0].Message.Content).Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
	}

	t.Logf("GPT1: %s\n", generated.Choices[0].Message.Content)

	generated, err = gpt2.AddMessage(ROLE_USER, generated.Choices[0].Message.Content).Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
	}

	t.Logf("GPT2: %s\n", generated.Choices[0].Message.Content)

	t.Logf("Query1: %+v", gpt1)

	t.Logf("Query2: %+v", gpt2)

	t.Logf("Generated: %+v", generated)
}
