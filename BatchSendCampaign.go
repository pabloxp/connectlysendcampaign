package BatchSendCampaign

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
  "sync"
  "time"
)

var Version string = "1.0"


// BatchSendCampaignRequest represents the request structure for BatchSendCampaign.
type BatchSendCampaignRequest struct {
	URL         string            `json:"url"`       // URL for sending API requests
	APIKey      string            `json:"api_key"`   // API key for authentication
	Headers     map[string]string `json:"headers"`   // Custom headers for the request
	BatchSize   int               `json:"batch_size"` // Number of messages to send in each batch
	CSVFilePath string            `json:"csv_file_path"` // Path to the CSV file
}

// BatchSendCampaignResponse represents the response structure for BatchSendCampaign.
type BatchSendCampaignResponse struct {
	Message   string `json:"message"`    // Confirmation message
	NumBatches int    `json:"num_batches"` // Number of batches sent
}

// BatchSendCampaign sends batches of API requests and returns confirmation.
func BatchSendCampaign(request BatchSendCampaignRequest) (*BatchSendCampaignResponse, error) {
	// Create a logger for detailed logging
	logger := log.New(os.Stdout, "BatchSendCampaign: ", log.Ldate|log.Ltime|log.Lmicroseconds)

	logger.Printf("Starting BatchSendCampaign with CSV file: %s\n", request.CSVFilePath)

	// Open the CSV file
	file, err := os.Open(request.CSVFilePath)
	if err != nil {
		logger.Printf("Error opening CSV file: %v\n", err)
		return nil, err
	}
	defer file.Close()

	// Create a CSV reader
	reader := csv.NewReader(file)

  // Read the first line (header) and discard it
	_, err = reader.Read()
	if err != nil {
		logger.Printf("Error reading CSV header: %v\n", err)
		return nil, err
	}

	// Read all remaining CSV records
	records, err := reader.ReadAll()
	if err != nil {
		logger.Printf("Error reading CSV file: %v\n", err)
		return nil, err
	}


	logger.Printf("Total records in CSV file: %d\n", len(records))

	// Initialize the response
	response := BatchSendCampaignResponse{
		Message:   "Messages sent successfully",
		NumBatches: 0,
	}

	// Create a buffer to accumulate batch messages
	var batchData []map[string]string

  // Create a channel to control the rate of API requests
  rateLimit := make(chan struct{}, request.BatchSize) // Limit to BatchSize concurrent requests

  // Wait group to wait for all workers to finish
  var wg sync.WaitGroup

  // Timestamps and count for each send
  sendTimestamps := make([]time.Time, 0)
  sendCount := 0

  // Iterate through CSV records and send batches
  for _, record := range records {
      // Rate limit the requests
      rateLimit <- struct{}{}

      // Increment the wait group
      wg.Add(1)

      // Start a Goroutine to send the batch request
      go func(record []string) {
          defer func() {
              // Release the rate limiter and decrement the wait group when the Goroutine finishes
              <-rateLimit
              wg.Done()
          }()

          // Construct a map of data from CSV record (adjust as per your CSV structure)
          data := map[string]string{
              "channel_type":    record[0],
              "external_id":     record[1],
              "template_name:body_1": record[2],
              "template_name:body_2": record[3],
          }

          // Add data to the batch
          batchData = append(batchData, data)

          // Check if the batch size has been reached
          if len(batchData) >= request.BatchSize {
              // Send the batch request
              err := sendBatchRequest(request.URL, request.APIKey, request.Headers, batchData)
              if err != nil {
                  logger.Printf("Error sending batch request: %v\n", err)
                  return
              }

              // Record the timestamp and increment the send count
              sendTimestamps = append(sendTimestamps, time.Now())
              sendCount++

              // Clear the batch data
              batchData = nil

              // Increment the batch count
              response.NumBatches++
          }
      }(record)
    }

    // Wait for all workers to finish
    wg.Wait()

    // Send any remaining batch
    if len(batchData) > 0 {
        err := sendBatchRequest(request.URL, request.APIKey, request.Headers, batchData)
        if err != nil {
            logger.Printf("Error sending remaining batch request: %v\n", err)
            return nil, err
        }

        // Record the timestamp and increment the send count
        sendTimestamps = append(sendTimestamps, time.Now())
        sendCount++

        // Increment the batch count
        response.NumBatches++
    }

    logger.Printf("BatchSendCampaign completed successfully with %d batches\n", response.NumBatches)

    // Print the report with send count and timestamps
    fmt.Printf("Total sends: %d\n", sendCount)
    for i, timestamp := range sendTimestamps {
        fmt.Printf("Send %d timestamp: %s\n", i+1, timestamp.Format("2006-01-02 15:04:05"))
    }

    return &response, nil
  }

// sendBatchRequest sends a batch of API requests.
func sendBatchRequest(url, apiKey string, headers map[string]string, data []map[string]string) error {
	// Create JSON payload from data
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}


	// Create an HTTP client
	client := &http.Client{}

	// Create an HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	// Add headers to the request
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Set API key in headers
	req.Header.Set("X-API-Key", apiKey)

	// Set Content-Type header
	req.Header.Set("Content-Type", "application/json")

  // Set Content-Type header
	req.Header.Set("x-mock-response-code", "201")

	// Perform the HTTP request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check the response status code, you can handle success and error cases here
	if resp.StatusCode != http.StatusUnauthorized {
		fmt.Printf("HTTP status code: %d", resp.StatusCode)
	}else{
    return fmt.Errorf("HTTP request failed with status code: %d", resp.StatusCode)
  }

	return nil
}

func main() {
	// Example usage of BatchSendCampaign
	request := BatchSendCampaignRequest{
		URL:         "https://cde176f9-7913-4af7-b352-75e26f94fbe3.mock.pstmn.io/v1/businesses/f1980bf7-c7d6-40ec-b665-dbe13620bffa/send/whatsapp_templated_messages",
		APIKey:      "<API Key>",
		Headers:     map[string]string{"Accept": "application/json"},
		BatchSize:   5,
		CSVFilePath: "sample_connectly_campaign.csv",
	}

	response, err := BatchSendCampaign(request)
	if err != nil {
		log.Fatalf("Error: %v\n", err)
	}

	fmt.Printf("Total batches sent: %d\n", response.NumBatches)
	fmt.Println("Message:", response.Message)
}
