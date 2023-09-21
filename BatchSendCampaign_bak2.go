package main

import (
    "encoding/csv"
    "fmt"
    "os"
    "bytes"
    "net/http"
)

type CSVData struct {
    ChannelType    string
    ExternalID     string
    TemplateFields map[string]string
}

func main() {
    // Llama a la función para cargar el CSV y obtener la estructura de datos
    data, err := loadCSV("sample_connectly_campaign.csv")
    if err != nil {
        fmt.Println("Error al cargar el CSV:", err)
        return
    }

    // Listar todos los CSVEntry
    for i, entry := range data {
        fmt.Printf("Entrada %d:\n", i+1)
        fmt.Printf("ChannelType: %s\n", entry.ChannelType)
        fmt.Printf("ExternalID: %s\n", entry.ExternalID)

        fmt.Println("Campos de la plantilla:")
        for fieldName, fieldValue := range entry.TemplateFields {
            fmt.Printf("%s: %s\n", fieldName, fieldValue)
        }

        fmt.Println()
    }


    // URL de destino
    url := "https://cde176f9-7913-4af7-b352-75e26f94fbe3.mock.pstmn.io/v1/businesses/f1980bf7-c7d6-40ec-b665-dbe13620bffa/send/whatsapp_templated_messages"

    // Datos JSON que deseas enviar
    jsonData := []byte(`{"content": "here"}`)

    // Cabeceras personalizadas
    headers := map[string]string{
        "Content-Type":        "application/json",
        "Accept":              "application/json",
        "X-API-Key":           "<API Key>",
        "x-mock-response-code": "200",
    }

    // Realiza la solicitud POST
    //resp, err := performPOSTRequest(url, jsonData, headers)
    performPOSTRequest(url, jsonData, headers)
}

func loadCSV(filePath string) ([]CSVData, error) {
    // Abre el archivo CSV
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    // Lee el archivo CSV
    reader := csv.NewReader(file)
    lines, err := reader.ReadAll()
    if err != nil {
        return nil, err
    }

    // Define la estructura de datos para almacenar los campos CSV
    var data []CSVData

    // Procesa cada línea del archivo CSV
    for _, line := range lines {
        if len(line) < 4 {
            fmt.Println("Error: la línea no contiene suficientes campos")
            continue
        }

        // Extrae los campos del CSV
        channelType := line[0]
        externalID := line[1]
        templateFields := make(map[string]string)

        // Itera a través de los campos de plantilla
        for i := 2; i < len(line); i++ {
            fieldName := fmt.Sprintf("template_name:body_%d", i-1)
            fieldValue := line[i]
            templateFields[fieldName] = fieldValue
        }

        // Crea una instancia de CSVData y agrega los datos a la estructura de datos
        csvEntry := CSVData{
            ChannelType:    channelType,
            ExternalID:     externalID,
            TemplateFields: templateFields,
        }
        data = append(data, csvEntry)
    }

    // Retorna la estructura de datos cargada
    return data, nil
}


func performPOSTRequest(url string, jsonData []byte, headers map[string]string) (*http.Response, error) {
    // Crea un objeto Request para la solicitud POST
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
    if err != nil {
        return nil, err
    }

    // Agrega las cabeceras personalizadas
    for key, value := range headers {
        req.Header.Set(key, value)
    }

    // Realiza la solicitud POST
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }

    //

    if err != nil {
        fmt.Println("Error al realizar la solicitud POST:", err)
        return resp, nil
    }
    defer resp.Body.Close()

    // Verifica el código de respuesta
    if resp.StatusCode != http.StatusOK {
        fmt.Println("La solicitud POST no fue exitosa. Código de respuesta:", resp.Status)
        return resp, nil
    }
    return resp, nil

}
