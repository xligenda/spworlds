package spwmini

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

func CheckUser(user User, secretToken string) bool {
	if user.Hash == "" || secretToken == "" {
		return false
	}

	var sb strings.Builder
	sb.WriteString(user.ID)
	sb.WriteString(user.Username)
	sb.WriteString(user.MinecraftUUID)

	if user.Timestamp > 0 {
		fmt.Fprintf(&sb, "%d", user.Timestamp)
	}
	if len(user.Roles) > 0 {
		sb.WriteString(strings.Join(user.Roles, ","))
	}
	if user.IsAdmin {
		sb.WriteString("true")
	} else if sb.Len() > 0 && (user.Timestamp > 0 || len(user.Roles) > 0) {
		sb.WriteString("false")
	}

	mac := hmac.New(sha256.New, []byte(secretToken))
	mac.Write([]byte(sb.String()))
	expectedMac := mac.Sum(nil)
	expectedHash := hex.EncodeToString(expectedMac)

	return hmac.Equal([]byte(strings.ToLower(user.Hash)), []byte(strings.ToLower(expectedHash)))
}
