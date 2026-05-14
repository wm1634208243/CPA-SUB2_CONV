package converter

import (
	"encoding/json"
	"fmt"
	"time"
)

// CPAToSub2 converts one or more CPA JSON objects into Sub2API export format.
// Input can be a single object {} or an array [{},...].
func CPAToSub2(input []byte) ([]byte, error) {
	var accounts []CPAAccount
	if err := json.Unmarshal(input, &accounts); err != nil {
		var single CPAAccount
		if err2 := json.Unmarshal(input, &single); err2 != nil {
			return nil, fmt.Errorf("failed to parse CPA JSON: %v", err2)
		}
		accounts = []CPAAccount{single}
	}

	export := Sub2APIExport{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Proxies:    []interface{}{},
		Accounts:   make([]Sub2Account, 0, len(accounts)),
	}

	for _, cpa := range accounts {
		claims := parseAccessTokenClaims(cpa.AccessToken)
		authClaims := map[string]interface{}{}
		if claims != nil {
			if v, ok := claims["https://api.openai.com/auth"]; ok {
				if m, ok := v.(map[string]interface{}); ok {
					authClaims = m
				}
			}
		}

		chatgptAccountID := getNestedString(authClaims, "chatgpt_account_id")
		if chatgptAccountID == "" {
			chatgptAccountID = cpa.AccountID
		}
		chatgptUserID := getNestedString(authClaims, "chatgpt_user_id")
		planType := getNestedString(authClaims, "chatgpt_plan_type")
		if planType == "" {
			planType = "plus"
		}

		orgID := ""
		idClaims := parseAccessTokenClaims(cpa.IDToken)
		if idClaims != nil {
			if authData, ok := idClaims["https://api.openai.com/auth"]; ok {
				if m, ok := authData.(map[string]interface{}); ok {
					if orgs, ok := m["organizations"].([]interface{}); ok && len(orgs) > 0 {
						if org, ok := orgs[0].(map[string]interface{}); ok {
							orgID, _ = org["id"].(string)
						}
					}
				}
			}
		}

		var tokenVersion int64
		if t, err := time.Parse(time.RFC3339, cpa.LastRefresh); err == nil {
			tokenVersion = t.UnixMilli()
		} else {
			tokenVersion = time.Now().UnixMilli()
		}

		creds := Sub2Credentials{
			TokenVersion:     tokenVersion,
			AccessToken:      cpa.AccessToken,
			ChatGPTAccountID: chatgptAccountID,
			ChatGPTUserID:    chatgptUserID,
			ClientID:         "app_EMoamEEZ73f0CkXaXp7hrann",
			Email:            cpa.Email,
			ExpiresAt:        cpa.Expired,
			IDToken:          cpa.IDToken,
			OrganizationID:   orgID,
			PlanType:         planType,
			RefreshToken:     cpa.RefreshToken,
		}

		acc := Sub2Account{
			Name:               cpa.Email,
			Platform:           "openai",
			Type:               "oauth",
			Credentials:        creds,
			Extra:              Sub2Extra{Email: cpa.Email, PrivacyMode: "training_off"},
			Concurrency:        10,
			Priority:           1,
			RateMultiplier:     1,
			AutoPauseOnExpired: true,
		}
		export.Accounts = append(export.Accounts, acc)
	}

	return json.MarshalIndent(export, "", "  ")
}

// Sub2ToCPA converts a Sub2API export into CPA format.
// Returns array if multiple accounts, single object if one.
func Sub2ToCPA(input []byte) ([]byte, error) {
	var export Sub2APIExport
	if err := json.Unmarshal(input, &export); err != nil {
		return nil, fmt.Errorf("failed to parse Sub2API JSON: %v", err)
	}
	if len(export.Accounts) == 0 {
		return nil, fmt.Errorf("Sub2API JSON does not contain any account data")
	}

	cpas := make([]CPAAccount, 0, len(export.Accounts))
	for _, acc := range export.Accounts {
		c := acc.Credentials

		expired := c.ExpiresAt
		if expired == "" {
			expired = time.Now().Add(10 * 24 * time.Hour).Format(time.RFC3339)
		}

		lastRefresh := ""
		if c.TokenVersion > 0 {
			lastRefresh = time.UnixMilli(c.TokenVersion).Format(time.RFC3339)
		} else {
			lastRefresh = time.Now().Format(time.RFC3339)
		}

		cpa := CPAAccount{
			Type:         "codex",
			Email:        c.Email,
			Expired:      expired,
			IDToken:      c.IDToken,
			AccountID:    c.ChatGPTAccountID,
			AccessToken:  c.AccessToken,
			LastRefresh:  lastRefresh,
			RefreshToken: c.RefreshToken,
		}
		cpas = append(cpas, cpa)
	}

	if len(cpas) == 1 {
		return json.MarshalIndent(cpas[0], "", "  ")
	}
	return json.MarshalIndent(cpas, "", "  ")
}

// DetectFormat tries to determine whether input is CPA or Sub2API format.
func DetectFormat(input []byte) string {
	var obj map[string]interface{}
	if err := json.Unmarshal(input, &obj); err == nil {
		if t, _ := obj["type"].(string); t == "codex" {
			return "cpa"
		}
		if _, ok := obj["accounts"]; ok {
			if _, ok2 := obj["exported_at"]; ok2 {
				return "sub2"
			}
		}
	}

	var arr []map[string]interface{}
	if err := json.Unmarshal(input, &arr); err == nil && len(arr) > 0 {
		if t, _ := arr[0]["type"].(string); t == "codex" {
			return "cpa"
		}
	}
	return "unknown"
}
