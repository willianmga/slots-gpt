package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/bedrock"
	"github.com/joho/godotenv"
)

type PromptRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

type PromptResponse struct {
	Response string `json:"response"`
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	awsRegion := os.Getenv("AWS_REGION")
	awsAccessKey := os.Getenv("AWS_ACCESS_KEY_ID")
	awsSecretKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(awsRegion),
		Credentials: credentials.NewStaticCredentials(awsAccessKey, awsSecretKey, ""),
	})
	if err != nil {
		log.Fatalf("Failed to create AWS session: %v", err)
	}

	svc := bedrock.New(sess)

	http.HandleFunc("/api/send-prompt", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
			return
		}

		var req PromptRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request payload", http.StatusBadRequest)
			return
		}

		if req.Prompt == "" || req.Model == "" {
			http.Error(w, "Prompt and model are required", http.StatusBadRequest)
			return
		}

		params := &bedrock.InvokeModelInput{
			InputText: aws.String(req.Prompt),
			ModelId:  aws.String(req.Model),
		}

		resp, err := svc.InvokeModel(params)
		if err != nil {
			log.Printf("Error invoking Bedrock model: %v", err)
			http.Error(w, "Failed to invoke Bedrock model", http.StatusInternalServerError)
			return
		}

		response := PromptResponse{
			Response: aws.StringValue(resp.OutputText),
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	})

	log.Printf("Server is running on port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

/**
 * To use this app:
 * 1. Create a .env file in the root directory and add the following:
 *    AWS_ACCESS_KEY_ID=<your_aws_access_key_id>
 *    AWS_SECRET_ACCESS_KEY=<your_aws_secret_access_key>
 *    AWS_REGION=<your_aws_region>
 *    PORT=<optional_port>
 * 
 * 2. Install dependencies:
 *    go get github.com/aws/aws-sdk-go github.com/joho/godotenv
 * 
 * 3. Run the app:
 *    go run main.go
 *
 * 4. Send POST requests to http://localhost:<port>/api/send-prompt
 *    with JSON payloads like:
 *    {
 *      "prompt": "Hello, Bedrock!",
 *      "model": "example-model-id"
 *    }
 */

