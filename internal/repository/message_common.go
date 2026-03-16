package repository

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/teamsphere/server/internal/model"
)

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func joinClauses(clauses []string, sep string) string {
	return strings.Join(clauses, sep)
}

func truncateRunes(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}

func marshalForwardMeta(meta *model.ForwardInfo) ([]byte, error) {
	if meta == nil {
		return nil, nil
	}
	raw, err := json.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("marshal forward_meta: %w", err)
	}
	return raw, nil
}

func unmarshalForwardMeta(raw []byte) *model.ForwardInfo {
	if len(raw) == 0 {
		return nil
	}
	var meta model.ForwardInfo
	if err := json.Unmarshal(raw, &meta); err != nil {
		return nil
	}
	return &meta
}
