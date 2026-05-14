package converter

import "encoding/json"

// ─── CPA (codex) format ───────────────────────────────────────────────────────

type CPAAccount struct {
	Type         string `json:"type"`
	Email        string `json:"email"`
	Expired      string `json:"expired"`
	IDToken      string `json:"id_token"`
	AccountID    string `json:"account_id"`
	AccessToken  string `json:"access_token"`
	LastRefresh  string `json:"last_refresh"`
	RefreshToken string `json:"refresh_token"`
}

// ─── Sub2API format ───────────────────────────────────────────────────────────

type Sub2APIExport struct {
	ExportedAt string        `json:"exported_at"`
	Proxies    []interface{} `json:"proxies"`
	Accounts   []Sub2Account `json:"accounts"`
}

type Sub2Account struct {
	Name               string          `json:"name"`
	Platform           string          `json:"platform"`
	Type               string          `json:"type"`
	Credentials        Sub2Credentials `json:"credentials"`
	Extra              Sub2Extra       `json:"extra"`
	Concurrency        int             `json:"concurrency"`
	Priority           int             `json:"priority"`
	RateMultiplier     float64         `json:"rate_multiplier"`
	AutoPauseOnExpired bool            `json:"auto_pause_on_expired"`
}

type Sub2Credentials struct {
	TokenVersion     int64             `json:"_token_version"`
	AccessToken      string            `json:"access_token"`
	ChatGPTAccountID string            `json:"chatgpt_account_id"`
	ChatGPTUserID    string            `json:"chatgpt_user_id"`
	ClientID         string            `json:"client_id"`
	Email            string            `json:"email"`
	ExpiresAt        string            `json:"expires_at"`
	IDToken          string            `json:"id_token"`
	ModelMapping     map[string]string `json:"model_mapping,omitempty"`
	OrganizationID   string            `json:"organization_id,omitempty"`
	PlanType         string            `json:"plan_type"`
	RefreshToken     string            `json:"refresh_token"`
}

type Sub2Extra struct {
	Email       string `json:"email"`
	PrivacyMode string `json:"privacy_mode,omitempty"`
	// codex usage fields (optional)
	Codex5hUsedPercent int    `json:"codex_5h_used_percent,omitempty"`
	Codex7dUsedPercent int    `json:"codex_7d_used_percent,omitempty"`
	Codex7dResetAt     string `json:"codex_7d_reset_at,omitempty"`
}

// ─── Conversion helpers ───────────────────────────────────────────────────────

// parseAccessTokenClaims extracts fields from JWT payload (no verification)
func parseAccessTokenClaims(token string) map[string]interface{} {
	parts := splitJWT(token)
	if len(parts) < 2 {
		return nil
	}
	payload, err := base64DecodeSegment(parts[1])
	if err != nil {
		return nil
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil
	}
	return claims
}

func getNestedString(m map[string]interface{}, keys ...string) string {
	cur := m
	for i, k := range keys {
		v, ok := cur[k]
		if !ok {
			return ""
		}
		if i == len(keys)-1 {
			if s, ok := v.(string); ok {
				return s
			}
			return ""
		}
		if next, ok := v.(map[string]interface{}); ok {
			cur = next
		} else {
			return ""
		}
	}
	return ""
}
