package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"golang.org/x/time/rate"
)

// BatchSendCampaignRequest represents the input parameters for the BatchSendCampaign function.
type BatchSendCampaignRequest struct {
	CSVFilePath     string        // Path to the CSV file with message data
	APIKey          string        // API key for connecting to the remote API
	BatchSize       int           // Number of messages to send in each batch
	TimeoutPerBatch time.Duration // Timeout for each batch of API requests
	MsgsPerSecond   int           // Rate limit for API requests (msgs per second)
}

// BatchSendCampaignResponse represents the output of the BatchSendCampaign function.
type BatchSendCampaignResponse struct {
	TotalMessagesSent int           // Total number of messages sent
	SuccessfulBatches int           // Number of successful batches
	FailedBatches     int           // Number of failed batches
	ExecutionReport   []string      // Report of the script execution, including API call details
}

// APIResult simulates the result of an API call.
type APIResult struct {
	Success bool   // Indicates if the API call was successful
	Details string // Details of the API call (for the execution report)
}

// sendBatchAPIRequest simulates sending a batch of API requests and returns an APIResult.
func sendBatchAPIRequest(apiKey string, batch []string, rateLimiter *rate.Limiter) APIResult {
	// Wait for rate limiting permission with a context
	ctx := context.Background()
	if err := rateLimiter.WaitN(ctx, len(batch)); err != nil {
		return APIResult{
			Success: false,
			Details: fmt.Sprintf("Rate limiting error: %v", err),
		}
	}

	// Simulate sending API requests here
	// Replace this with your actual API call logic

	// For demonstration purposes, we'll just print the batch size
	details := fmt.Sprintf("Sending batch of %d messages to the remote API with API key: %s", len(batch), apiKey)
	fmt.Println(details)

	// Simulate success
	return APIResult{
		Success: true,
		Details: details,
	}
}

// BatchSendCampaign sends batches of API requests asynchronously and returns a report of the script execution.
func BatchSendCampaign(req *BatchSendCampaignRequest) *BatchSendCampaignResponse {
	// Initialize the response struct
	response := &BatchSendCampaignResponse{}

	// Open the CSV file
	file, err := os.Open(req.CSVFilePath)
	if err != nil {
		response.ExecutionReport = append(response.ExecutionReport, fmt.Sprintf("Error opening CSV file: %v", err))
		return response
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

	// Create a rate limiter
	rateLimiter := rate.NewLimiter(rate.Limit(req.MsgsPerSecond), req.MsgsPerSecond)

	// Create a channel for API call results
	resultChan := make(chan APIResult)

	// Read the CSV data and process it in batches
	var batch []string
	totalMessagesSent := 0

	for {
		record, err := reader.Read()
		if err != nil {
			break // End of file
		}

		// Process the CSV record and convert it to an API request
		message := record[0] // Assuming the CSV contains a single column for messages

		batch = append(batch, message)
		if len(batch) == req.BatchSize {
			// Send the batch of API requests asynchronously
			go func(apiKey string, batch []string) {
				apiCallResult := sendBatchAPIRequest(apiKey, batch, rateLimiter)
				resultChan <- apiCallResult
			}(req.APIKey, batch)

			totalMessagesSent += len(batch)

			// Reset the batch
			batch = nil

			// Sleep for the specified timeout
			time.Sleep(req.TimeoutPerBatch)
		}
	}

	// Check if there's a remaining batch to send
	if len(batch) > 0 {
		// Send the remaining batch of API requests asynchronously
		go func(apiKey string, batch []string) {
			apiCallResult := sendBatchAPIRequest(apiKey, batch, rateLimiter)
			resultChan <- apiCallResult
		}(req.APIKey, batch)

		totalMessagesSent += len(batch)
	}

	// Close the result channel when all API calls are done
	close(resultChan)

	// Collect API call results and update the response
	for apiResult := range resultChan {
		if apiResult.Success {
			response.SuccessfulBatches++
		} else {
			response.FailedBatches++
		}

		// Append API call details to the execution report
		response.ExecutionReport = append(response.ExecutionReport, apiResult.Details)
	}

	// Calculate execution time
	response.TotalMessagesSent = totalMessagesSent

	return response
}

func main() {
	// Example usage:
	request := &BatchSendCampaignRequest{
		CSVFilePath:     "sample_connectly_campaign.csv",
		APIKey:          "your-api-key",
		BatchSize:       10,
		TimeoutPerBatch: 2 * time.Second,
		MsgsPerSecond:   5, // Adjust the rate limit as needed
	}

	response := BatchSendCampaign(request)

	// Print the script execution report
	fmt.Printf("Total Messages Sent: %d\n", response.TotalMessagesSent)
	fmt.Printf("Successful Batches: %d\n", response.SuccessfulBatches)
	fmt.Printf("Failed Batches: %d\n", response.FailedBatches)
	fmt.Println("Execution Report:")
	for _, details := range response.ExecutionReport {
		fmt.Println(details)
	}
}
