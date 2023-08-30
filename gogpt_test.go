package gogpt

import (
	"encoding/json"
	"os"
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
