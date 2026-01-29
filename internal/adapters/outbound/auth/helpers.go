// internal/adapters/outbound/auth/helpers.go
package auth

func stringPtr(v string) *string {
	return &v
}

func containsAudience(raw any, audience string) bool {
	for _, aud := range normalizeAudience(raw) {
		if aud == audience {
			return true
		}
	}
	return false
}

func normalizeAudience(raw any) []string {
	switch v := raw.(type) {
	case nil:
		return nil
	case string:
		if v == "" {
			return nil
		}
		return []string{v}
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	default:
		return nil
	}
}
