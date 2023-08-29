package gogpt

import (
	"encoding/json"
	"os"
	"testing"
)

func TestGenerate(t *testing.T) {

	f := "./testconfig.json"
	file, err := os.ReadFile(f)

	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	gpt := new(GoGPT)
	err = json.Unmarshal([]byte(file), &gpt)

	if err != nil {
		t.Errorf("Error loading config: %v", err)
	}

	prompt := "Can pigs fly?"
	generated, err := gpt.Generate(prompt)
	if err != nil {
		t.Errorf("Error generating: %v", err)
	}

	t.Logf("Generated: %+v", generated)
}
