package converter

import (
	"encoding/base64"
	"strings"
)

func splitJWT(token string) []string {
	return strings.Split(token, ".")
}

func base64DecodeSegment(seg string) ([]byte, error) {
	// JWT uses raw base64url (no padding)
	switch len(seg) % 4 {
	case 2:
		seg += "=="
	case 3:
		seg += "="
	}
	return base64.URLEncoding.DecodeString(seg)
}
