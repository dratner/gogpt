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

func TestGenerate(t *testing.T) {

	f := "./testconfig.json"
	file, err := os.ReadFile(f)

	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	conf := new(TestConfig)
	err = json.Unmarshal([]byte(file), &conf)

	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	gpt := NewGoGPTQuery(conf.GptKey)
	gpt.OrgName = conf.GptOrgName
	gpt.OrgId = conf.GptOrgId
	gpt.MaxTokens = 100

	gpt.AddMessage(ROLE_SYSTEM, "Can pigs fly?")

	t.Logf("Query: %+v", gpt)

	generated, err := gpt.Generate()

	if err != nil {
		t.Errorf("Error generating: %v", err)
	}

	t.Logf("Generated: %+v", generated)
}
