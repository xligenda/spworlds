package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"os"
)

const MaxStoredDonations = 50

func generateRequestID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("Warning: failed to generate random bytes: %v", err)
		return "fallback-id"
	}
	return hex.EncodeToString(bytes)
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("%s not set", key)
	}
	return val
}
