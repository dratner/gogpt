package gogpt

import (
	"testing"
)

func TestEmbedding(t *testing.T) {

	gpt, err := buildTestQueryHelper()

	if err != nil {
		t.Errorf("Error building config: %v", err)
	}

	emb, err := GetEmbedding("Hello, world!", gpt.Key)

	if err != nil {
		t.Errorf("Error embeddings: %v", err)
	}

	t.Logf("Embeddings: %+v", emb)

}
