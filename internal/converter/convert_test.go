package converter

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
)

func makeJWT(payload string) string {
	return "header." + base64.RawURLEncoding.EncodeToString([]byte(payload)) + ".signature"
}

func TestDetectFormat(t *testing.T) {
	cpaInput := []byte(`{"type":"codex","email":"demo@example.com"}`)
	sub2Input := []byte(`{"exported_at":"2026-05-14T00:00:00Z","accounts":[]}`)

	if got := DetectFormat(cpaInput); got != "cpa" {
		t.Fatalf("expected cpa, got %s", got)
	}
	if got := DetectFormat(sub2Input); got != "sub2" {
		t.Fatalf("expected sub2, got %s", got)
	}
}

func TestCPAToSub2(t *testing.T) {
	authPayload := `{"https://api.openai.com/auth":{"chatgpt_account_id":"acct_demo","chatgpt_user_id":"user_demo","chatgpt_plan_type":"plus"}}`
	idPayload := `{"https://api.openai.com/auth":{"organizations":[{"id":"org_demo"}]}}`

	input := []byte(`{
		"type":"codex",
		"email":"demo@example.com",
		"expired":"2026-12-31T00:00:00Z",
		"id_token":"` + makeJWT(idPayload) + `",
		"account_id":"acct_demo",
		"access_token":"` + makeJWT(authPayload) + `",
		"last_refresh":"2026-05-14T00:00:00Z",
		"refresh_token":"refresh_demo"
	}`)

	out, err := CPAToSub2(input)
	if err != nil {
		t.Fatalf("CPAToSub2 returned error: %v", err)
	}

	var parsed Sub2APIExport
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	if len(parsed.Accounts) != 1 {
		t.Fatalf("expected 1 account, got %d", len(parsed.Accounts))
	}
	if parsed.Accounts[0].Credentials.Email != "demo@example.com" {
		t.Fatalf("unexpected email: %s", parsed.Accounts[0].Credentials.Email)
	}
	if parsed.Accounts[0].Credentials.OrganizationID != "org_demo" {
		t.Fatalf("unexpected organization id: %s", parsed.Accounts[0].Credentials.OrganizationID)
	}
}

func TestSub2ToCPA(t *testing.T) {
	input := []byte(`{
		"exported_at":"2026-05-14T00:00:00Z",
		"proxies":[],
		"accounts":[
			{
				"name":"demo@example.com",
				"platform":"openai",
				"type":"oauth",
				"credentials":{
					"_token_version":1747180800000,
					"access_token":"access_demo",
					"chatgpt_account_id":"acct_demo",
					"chatgpt_user_id":"user_demo",
					"client_id":"client_demo",
					"email":"demo@example.com",
					"expires_at":"2026-12-31T00:00:00Z",
					"id_token":"id_demo",
					"plan_type":"plus",
					"refresh_token":"refresh_demo"
				},
				"extra":{"email":"demo@example.com"},
				"concurrency":10,
				"priority":1,
				"rate_multiplier":1,
				"auto_pause_on_expired":true
			}
		]
	}`)

	out, err := Sub2ToCPA(input)
	if err != nil {
		t.Fatalf("Sub2ToCPA returned error: %v", err)
	}

	var parsed CPAAccount
	if err := json.Unmarshal(out, &parsed); err != nil {
		t.Fatalf("failed to unmarshal output: %v", err)
	}
	if parsed.Type != "codex" {
		t.Fatalf("unexpected type: %s", parsed.Type)
	}
	if parsed.AccountID != "acct_demo" {
		t.Fatalf("unexpected account id: %s", parsed.AccountID)
	}
}

func TestSub2ToCPARejectsEmptyAccounts(t *testing.T) {
	_, err := Sub2ToCPA([]byte(`{"exported_at":"2026-05-14T00:00:00Z","proxies":[],"accounts":[]}`))
	if err == nil {
		t.Fatal("expected error for empty accounts")
	}
	if !strings.Contains(err.Error(), "does not contain any account data") {
		t.Fatalf("unexpected error: %v", err)
	}
}
