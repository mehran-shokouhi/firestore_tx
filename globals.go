package main

import (
	"encoding/base64"
	"os"

	"go.uber.org/zap"
)

var L *zap.Logger
var (
	FirestoreCredentials []byte
	FirestoreProjectID   string
)

func setupLogger() {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.FunctionKey = "func"
	L, _ = config.Build()
}

func setupEnv() {
	envName := "FIRESTORE_CREDENTIALS"
	firestoreCreds := os.Getenv(envName)
	if firestoreCreds == "" {
		L.Fatal(envName + " environment variable is not set")
	}

	var err error
	FirestoreCredentials, err = base64.StdEncoding.DecodeString(firestoreCreds)
	if err != nil {
		L.Fatal("Invalid Firestore credentials")
	}

	envName = "FIRESTORE_PROJECT_ID"
	FirestoreProjectID = os.Getenv(envName)
	if FirestoreProjectID == "" {
		L.Fatal(envName + " environment variable is not set")
	}
}

func SetupGlobals() {
	setupLogger()
	setupEnv()
}
