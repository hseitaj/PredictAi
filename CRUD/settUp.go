package main

/*
	CMPSC 488
	Fall 2023
	PredictAI
*/

// Import required packages
import (
	"database/sql"  // For SQL database interaction
	"encoding/json" // For JSON handling
	"fmt"           // For formatted I/O
	"github.com/google/uuid"
	"io/ioutil" // For I/O utility functions
	"log"       // For logging
	"time"      // For time manipulation

	_ "github.com/go-sql-driver/mysql" // Import MySQL driver
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

// JSON_Data_Connect struct models the structure of database credentials in config.json
type JSON_Data_Connect struct {
	Username string
	Password string
	Hostname string
	Database string
}

// init initializes the program, reading the database configuration and establishing a connection
func init() {
	// Read database credentials from config.json
	config, err := readJSONConfig("config.json")
	if err != nil {
		log.Fatal("Error reading JSON config:", err)
		return
	}

	// Establish a new database connection
	var connErr error
	db, connErr = Connection(config)
	if connErr != nil {
		log.Fatal("Error establishing database connection:", connErr)
	}
}

// WriteLog writes a log entry to the database
func WriteLog(logID string, status_code string, message string, goEngineArea string, dateTime time.Time) error {
	// Validate the statusCode by checking if it exists in the `log_status_codes` table
	var existingStatusCode string
	err := db.QueryRow("SELECT status_code FROM log_status_codes WHERE status_code = ?", status_code).Scan(&existingStatusCode)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("Invalid statusCode: %s", status_code)
		}
		return err
	}

	// Prepare the SQL statement for inserting into the log table
	stmt, err := db.Prepare("INSERT INTO log(logID, status_code, message, go_engine_area, date_time) VALUES (? ,? ,? ,? ,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	// Execute the SQL statement
	_, errExec := stmt.Exec(logID, existingStatusCode, message, goEngineArea, dateTime)
	if errExec != nil {
		return errExec
	}

	return nil
}

// Insertstatus_code inserts a new status code into the `log_status_codes` table
func InsertStatusCode(status_code, description string) error {
	config, err := readJSONConfig("config.json")
	if err != nil {
		return err
	}
	db, err := Connection(config)
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO log_status_codes(status_code, status_message) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, errExec := stmt.Exec(status_code, description)
	if errExec != nil {
		return errExec
	}

	return nil
}

// GetLog retrieves all logs from the database
func GetLog() ([]Log, error) {
	stmt, err := db.Prepare("CALL SelectAllLogs()")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var logItem Log
		var dateTimeStr []uint8
		err := rows.Scan(&logItem.LogID, &logItem.status_code, &logItem.Message, &logItem.GoEngineArea, &dateTimeStr)
		if err != nil {
			return nil, err
		}
		logItem.DateTime = dateTimeStr
		logs = append(logs, logItem)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// GetSuccess fetches all logs with a "Success" status code
func GetSuccess() ([]Log, error) {
	stmt, err := db.Prepare("CALL SelectAllLogsByStatusCode(?)")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query("Success")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []Log
	for rows.Next() {
		var logItem Log
		var dateTimeStr []uint8
		err := rows.Scan(&logItem.LogID, &logItem.status_code, &logItem.Message, &logItem.GoEngineArea, &dateTimeStr)
		if err != nil {
			return nil, err
		}
		logItem.DateTime = dateTimeStr
		logs = append(logs, logItem)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return logs, nil
}

// InsertOrUpdatestatus_code either inserts a new status code or updates an existing one
//func InsertOrUpdateStatusCode(status_code, description string) error {
//	var existingStatusCode string
//	err := db.QueryRow("SELECT status_code FROM log_status_codes WHERE status_code = ?", status_code).Scan(&existingStatusCode)
//
//	stmtStr := ""
//	if err == sql.ErrNoRows {
//		stmtStr = "INSERT INTO log_status_codes(status_code, status_message) VALUES (?, ?)"
//	} else if err == nil {
//		stmtStr = "UPDATE log_status_codes SET status_message = ? WHERE status_code = ?"
//	} else {
//		return err
//	}
//
//	stmt, err := db.Prepare(stmtStr)
//	if err != nil {
//		return err
//	}
//	defer stmt.Close()
//
//	if err == sql.ErrNoRows {
//		_, err = stmt.Exec(status_code, description)
//	} else {
//		_, err = stmt.Exec(description, status_code)
//	}
//	return err
//}

// StoreLog stores a log entry using a stored procedure
func StoreLog(status_code string, message string, goEngineArea string) error {
	stmt, err := db.Prepare("CALL InsertLog(?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, errExec := stmt.Exec(status_code, message, goEngineArea)
	if errExec != nil {
		return errExec
	}

	return nil
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

// User Management Functions

// CreateUser creates a new user in the database
func CreateUser(name, login, role, password string, active bool) error {
	stmt, err := db.Prepare("CALL goengine.create_user(?, ?, ?, ?, ?)") // Updated to match SQL
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, errExec := stmt.Exec(name, login, role, password, active)
	if errExec != nil {
		return errExec
	}

	return nil
}

//// UpdateUser updates an existing user in the database
//func UpdateUser(id, name, login, role, password string) error {
//	stmt, err := db.Prepare("CALL goengine.update_user(?, ?, ?, ?, ?)") // Updated to match SQL
//	if err != nil {
//		return err
//	}
//	defer stmt.Close()
//
//	_, errExec := stmt.Exec(id, name, login, role, password)
//	if errExec != nil {
//		return errExec
//	}
//
//	return nil
//}

//// UpdateUser updates an existing user in the database
//func UpdateUser(id, name, login, role, password string) error {
//	stmt, err := db.Prepare("CALL goengine.update_user(?, ?, ?, ?, ?)") // Updated to match SQL
//	if err != nil {
//		fmt.Println("Prepare Error:", err) // Debug line
//		return err
//	}
//	defer stmt.Close()
//
//	_, errExec := stmt.Exec(id, name, login, role, password)
//	if errExec != nil {
//		fmt.Println("Exec Error:", errExec) // Debug line
//		return errExec
//	}
//
//	return nil
//}

const maxLength = 3 // Add maxLength constant for validation

func InsertOrUpdateStatusCode(statusCode, statusMessage string) error {
	if len(statusCode) > maxLength { // maxLength should be defined to match your DB schema
		return fmt.Errorf("status code is too long: %s", statusCode)
	}

	stmt, err := db.Prepare("INSERT INTO log_status_codes(status_code, status_message) VALUES (?, ?) ON DUPLICATE KEY UPDATE message = VALUES(message)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, errExec := stmt.Exec(statusCode, statusMessage)
	if errExec != nil {
		return errExec
	}
	return nil
}

func FetchUserID(login string) (string, error) {
	var userID string
	query := "SELECT user_id FROM users WHERE user_login = ?"
	fmt.Printf("Executing query: %s with login = %s\n", query, login)
	err := db.QueryRow(query, login).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("no user with login: %s", login)
		}
		return "", err
	}
	return userID, nil
}

// UpdateUser updates an existing user in the database.
func UpdateUser(name, login, role, password string) error {
	userID, err := FetchUserID(login)
	if err != nil {
		return fmt.Errorf("failed to fetch user ID: %w", err)
	}
	stmt, err := db.Prepare("CALL goengine.update_user(?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, errExec := stmt.Exec(userID, name, login, role, password)
	if errExec != nil {
		return fmt.Errorf("failed to update user: %w", errExec)
	}
	return nil
}

// DeleteUser removes a user from the database
func DeleteUser(id string) error {
	stmt, err := db.Prepare("CALL delete_user(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, errExec := stmt.Exec(id)
	if errExec != nil {
		return errExec
	}

	return nil
}

// main function to test all existing methods
func main() {

	// Initialize database connection
	if db == nil {
		log.Fatal("Database connection is not initialized.")
	}

	// Insert or Update status code
	err := InsertOrUpdateStatusCode("POS", "noth")
	if err != nil {
		log.Println("Failed to insert or update status code:", err)
	}

	_, err = FetchUserID("jxo19")
	if err != nil {
		log.Fatalf("Failed to fetch user ID: %v", err)
	}

	// Update User
	err = UpdateUser("NewName", "jxo19", "ADM", "newpassword")
	if err != nil {
		fmt.Printf("Failed to update user: %s\n", err)
	} else {
		fmt.Println("Successfully updated user")
	}

	// Delete User
	err = DeleteUser("jxo19")
	if err != nil {
		fmt.Printf("Failed to delete user: %s\n", err)
	} else {
		fmt.Println("Successfully deleted user")
	}

	//Generate a unique logID
	uniqueLogID := uuid.New().String()

	//Write log
	currentTime := time.Now()
	err = WriteLog(uniqueLogID, "Pos", "Message logged successfully", "Engine1", currentTime)
	if err != nil {
		log.Println("Failed to write log:", err)
	}

	// Get and print all logs
	logs, err := GetLog()
	if err != nil {
		log.Println("Failed to get logs:", err)
	} else {
		for _, logItem := range logs {
			fmt.Println(logItem)
		}
	}

	//Store log using a stored procedure (uncomment if needed)
	err = StoreLog("Success", "Stored using procedure", "Engine1")
	if err != nil {
		log.Println("Failed to store log using stored procedure:", err)
	}

	//Insert a new status code
	err = InsertStatusCode("200", "OK")
	if err != nil {
		log.Println("Failed to insert new status code:", err)
	}

	//Create a new user
	err = CreateUser("John", "john123", "ADM", "password", true)
	if err != nil {
		log.Println("Failed to create a new user:", err)
	}

	//Delete a user
	//err = DeleteUser("john123")
	//if err != nil {
	//	log.Println("Failed to delete user:", err)
	//}
	//
	////Get and print all "Success" logs
	//successLogs, err := GetSuccess()
	//if err != nil {
	//	log.Println("Failed to get success logs:", err)
	//} else {
	//	fmt.Println("Success Logs:")
	//	for _, logItem := range successLogs {
	//		fmt.Println(logItem)
	//	}
	//}
}
