package gogpt

import (
	"testing"
)

const (
	GPTKEY         = "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"
	GPTORGNAME     = "Personal"
	GPTORGID       = "org-11qnlvT2cl1MBRjIF8z8dGCc"
	GTPENDPOINT    = "https://api.openai.com/v1/chat/completions"
	GPTMODEL       = "gpt-3.5-turbo"
	GPTUSER        = "Personal"
	GPTROLE        = "system"
	GPTTEMPERATURE = 0.7
)

func TestGenerate(t *testing.T) {
	gpt := NewGoGPT(GPTKEY, GPTORGNAME, GPTORGID, GTPENDPOINT, GPTMODEL, GPTUSER, GPTROLE, GPTTEMPERATURE)
	prompt := "What is the factorial of 10?"
	generated, err := gpt.Generate(prompt)
	if err != nil {
		t.Errorf("Error generating: %v", err)
	}
	if generated == "" {
		t.Errorf("Generated string is empty")
	}

	t.Logf("Generated: %v", generated)
}
