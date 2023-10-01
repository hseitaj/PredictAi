package main

// Import required packages
import (
	"database/sql"                     // For SQL database interaction
	"encoding/json"                    // For JSON handling
	"fmt"                              // For formatted I/O
	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
	"io/ioutil"                        // For I/O utility functions
	"log"                              // For logging
	"sync"                             // For multi-threading
	"time"                             // For simulating machine learning model processing time
)

// Declare db at the package level for global use
var db *sql.DB

// Log struct models the data structure of a log entry in the database
type Log struct {
	LogID        string
	status_code  string
	Message      string
	GoEngineArea string
	DateTime     []uint8
}

// Prediction struct models the data structure of a prediction in the database
type Prediction struct {
	PredictionID   string
	EngineID       string
	InputData      string
	PredictionInfo string
	PredictionTime string
}

// JSON_Data_Connect struct models the structure of database credentials in config.json
type JSON_Data_Connect struct {
	Username string
	Password string
	Hostname string
	Database string
}

// init initializes the program, reading the database configuration and establishing a connection
func init() {
	config, err := readJSONConfig("config.json")
	if err != nil {
		log.Fatal("Error reading JSON config:", err)
	}

	var connErr error
	db, connErr = Connection(config)
	if connErr != nil {
		log.Fatal("Error establishing database connection:", connErr)
	}
}

// Connection establishes a new database connection based on provided credentials
func Connection(config JSON_Data_Connect) (*sql.DB, error) {
	connDB, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", config.Username, config.Password, config.Hostname, config.Database))
	if err != nil {
		return nil, err
	}

	err = connDB.Ping()
	if err != nil {
		return nil, err
	}

	return connDB, nil
}

// readJSONConfig reads database credentials from a JSON file
func readJSONConfig(filename string) (JSON_Data_Connect, error) {
	var config JSON_Data_Connect
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(file, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// Function to check if the engine_id exists in scraper_engine table
func engineIDExists(engineID string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM scraper_engine WHERE engine_id=?)"
	err := db.QueryRow(query, engineID).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Function to insert a new prediction
func insertPrediction(engineID string, predictionInfo string) error {
	exists, err := engineIDExists(engineID)
	if err != nil {
		return fmt.Errorf("Error checking engine ID: %v", err)
	}
	if !exists {
		return fmt.Errorf("engine_id %s does not exist", engineID)
	}

	query := "INSERT INTO predictions (engine_id, prediction_info) VALUES (?, ?)"
	_, err = db.Exec(query, engineID, predictionInfo)
	if err != nil {
		return fmt.Errorf("Error storing prediction: %v", err)
	}
	return nil
}

// Function to insert a sample engine ID into scraper_engine table
func insertSampleEngine(engineID, engineName, engineDescription string) error {
	query := "INSERT INTO scraper_engine (engine_id, engine_name, engine_description) VALUES (?, ?, ?)"
	_, err := db.Exec(query, engineID, engineName, engineDescription)
	if err != nil {
		return fmt.Errorf("Error inserting sample engine: %v", err)
	}
	return nil
}

// Simulated ML model prediction function
func performMLPrediction(inputData string) string {
	// Simulate some delay for ML model prediction
	time.Sleep(2 * time.Second)
	return fmt.Sprintf("Prediction result for %s", inputData)
}

// Convert prediction result to JSON
func convertPredictionToJSON(predictionResult string) (string, error) {
	predictionMap := map[string]string{"result": predictionResult}
	predictionJSON, err := json.Marshal(predictionMap)
	if err != nil {
		return "", err
	}
	return string(predictionJSON), nil
}

func main() {
	if db == nil {
		log.Fatal("Database connection is not initialized.")
	}

	// Using a WaitGroup for multi-threading
	var wg sync.WaitGroup

	// Insert a sample engine ID
	sampleEngineID := "sample_engine_id"
	sampleEngineName := "Sample Engine"
	sampleEngineDescription := "This is a sample engine."
	exists, err := engineIDExists(sampleEngineID)
	if err != nil {
		log.Fatalf("Error checking if engine ID exists: %v", err)
	}

	if !exists {
		err = insertSampleEngine(sampleEngineID, sampleEngineName, sampleEngineDescription)
		if err != nil {
			log.Fatalf("Failed to insert sample engine: %v", err)
		}
	}

	// Simulate getting some prediction data and performing ML prediction
	predictionResult := performMLPrediction("Test Data")

	// Convert the prediction result to JSON
	predictionMap := map[string]string{"result": predictionResult}
	predictionJSON, err := json.Marshal(predictionMap)
	if err != nil {
		log.Fatalf("Failed to convert prediction to JSON: %v", err)
	}
	predictionInfo := string(predictionJSON)

	// Use goroutine to insert prediction
	wg.Add(1)
	go func() {
		defer wg.Done()
		err := insertPrediction(sampleEngineID, predictionInfo)
		if err != nil {
			log.Fatalf("Failed to insert prediction: %v", err)
		} else {
			log.Println("Successfully inserted prediction.")
		}
	}()

	// Wait for all goroutines to complete
	wg.Wait()
}
