# Connectly Send Campaign Go Package

![Version] 0.2.2
![License] Public Domain

The Connectly Send Campaign Go Package provides a set of functions to send campaign messages using the Connectly API. It includes the ability to download CSV files and send 
batched API requests.

## Installation

To use this package, you need to have Go installed on your system. You can install it using the following command:

bash
go get -u github.com/pabloxp/connectlysendcampaign


## Usage

package main

import (
    "fmt"
    "github.com/pabloxp/connectlysendcampaign"
)

func main() {
    // Replace these values with your own
    csvURL := "https://example.com/somefile.csv"
    outputPath := "downloaded.csv"
    apiKey := "<Your API Key>"

    // Download the CSV file
    err := connectlysendcampaign.DownloadCSVFile(csvURL, outputPath)
    if err != nil {
        fmt.Printf("Error downloading CSV file: %v\n", err)
        return
    }

    // Define the request
    request := connectlysendcampaign.BatchSendCampaignRequest{
        URL:         "https://example.com/api/send-campaign",
        APIKey:      apiKey,
        Headers:     map[string]string{"Content-Type": "application/json"},
        BatchSize:   5,
        CSVFilePath: outputPath,
    }

    // Send the campaign
    response, err := connectlysendcampaign.BatchSendCampaign(request)
    if err != nil {
        fmt.Printf("Error sending campaign: %v\n", err)
        return
    }

    fmt.Printf("Total batches sent: %d\n", response.NumBatches)
    fmt.Println("Message:", response.Message)
}


