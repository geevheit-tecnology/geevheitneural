package agents

import (
	"github.com/geevheit/neuralcenter/api/internal/domain/pipeline"
	"testing"
)

func TestCorrection004_RejectInvalidReply(t *testing.T) {
	for _, raw := range []string{"", "plain text", "null", "{}", `{"output":" "}`, `{"output":123}`} {
		if _, err := parseRoleOutput(pipeline.RoleDev, raw, nil); err == nil {
			t.Fatalf("accepted invalid reply %q", raw)
		}
	}
}
