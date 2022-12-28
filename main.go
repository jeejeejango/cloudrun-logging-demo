// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// [START cloudrun_pubsub_server]
// [START run_pubsub_server]

// Sample run-pubsub is a Cloud Run service which handles Pub/Sub messages.
package main

import (
	"cloud.google.com/go/logging"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	log.SetOutput(os.Stdout)

	ctx := context.Background()
	projectId := os.Getenv("PROJECT_ID")
	client, err := logging.NewClient(ctx, projectId)

	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}

	http.HandleFunc("/", LogContext{client}.LoggingPubSub)
	// Determine port for HTTP service.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
		log.Printf("Defaulting to port %s", port)
	}
	// Start HTTP server.
	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}

	defer func(client *logging.Client) {
		if err := client.Close(); err != nil {
			log.Fatalf("Failed to close client: %v", err)

		}
	}(client)
}

type LogContext struct {
	client *logging.Client
}

type PubSubMessage struct {
	Message struct {
		Data []byte `json:"data,omitempty"`
		ID   string `json:"id"`
	} `json:"message"`
	Subscription string `json:"subscription"`
}

func (logContext LogContext) LoggingPubSub(w http.ResponseWriter, r *http.Request) {
	var m PubSubMessage
	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("ioutil.ReadAll: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	// byte slice unmarshalling handles base64 decoding.
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("json.Unmarshal: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	var payload map[string]interface{}

	if err := json.Unmarshal(m.Message.Data, &payload); err != nil {
		log.Fatalf("Unable to parse JSON: %v", err)
	}

	logName := "demo_general_log"
	if payload["log_name"] != nil && payload["log_name"] != "" {
		logName = "demo_" + payload["log_name"].(string)
	}

	logger := logContext.client.Logger(logName)
	defer logger.Flush()

	sev := logging.Info
	if strings.Contains(strings.ToLower(logName), "error") {
		sev = logging.Error
	}

	logger.Log(logging.Entry{
		Severity: sev,
		Payload:  payload})
}
