// Package databaseFunctions provides functions for reading, writing, and modifying the bitcoin_wallet.db
package databaseFunctions

import (
	transaction_types "bitcoin_wallet_rest_api/transactionTypes"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"
)

var databaseDriver = "sqlite3"
var databasePath = "bitcoin_wallet.db"

// CreateNewTransaction creates a new transaction record in the database.
// It opens a database connection, starts a new transaction, generates a new transaction ID,
// and inserts a new transaction record with the specified amount, marked as unspent, and the
// current timestamp.
// If any operation fails, it rolls back the transaction and logs the error.
func CreateNewTransaction(value float64) error {

	value = roundToDecimalPlaces(value, 8)

	db, err := openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer tx.Rollback()

	newTransactionID, err := generateTransactionID()
	if err != nil {
		fmt.Println(err)
		return err
	}

	insertTransactionQuery := `
	INSERT INTO transactions (id, amount, spent, created_at)
	VALUES (?, ?, ?, ?);
	`
	current_time := time.Now()
	formattedTime := current_time.Format("2006-01-02 15:04:05")

	_, err = tx.Exec(insertTransactionQuery, newTransactionID, value, false, formattedTime)
	if err != nil {
		fmt.Println(err)
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}

	db.Close()

	return nil
}

func roundToDecimalPlaces(number float64, decimalPlaces int) float64 {
	// Convert the number to a string with the desired precision
	str := fmt.Sprintf("%.*f", decimalPlaces, number)
	// Convert the string back to a float64
	roundedNumber, err := strconv.ParseFloat(str, 64)
	if err != nil {
		panic(err) // Handle error appropriately
	}
	return roundedNumber
}

// GetAllTransactions retrieves all transaction records from the database.
// It opens a database connection, executes a SELECT query to fetch all transactions,
// scans the results into a slice of Transaction structs, and then returns this slice.
// If any operation fails, it returns an empty slice and the error encountered.
func GetAllTransactions() ([]transaction_types.Transaction, error) {
	db, err := openDatabase()
	if err != nil {
		return nil, err
	}

	defer db.Close()

	allTransactions := []transaction_types.Transaction{}

	queryTxt := `SELECT * FROM transactions`
	rows, err := db.Query(queryTxt)

	if err != nil {
		return allTransactions, err
	}

	for rows.Next() {
		transaction := transaction_types.Transaction{}
		err = rows.Scan(&transaction.TransactionID, &transaction.Amount, &transaction.Spent, &transaction.CreatedAt)
		if err != nil {
			return []transaction_types.Transaction{}, err
		}
		allTransactions = append(allTransactions, transaction)
	}

	err = db.Close()

	if err != nil {
		return []transaction_types.Transaction{}, err
	}

	return allTransactions, nil
}

// Mark_Transaction_Used marks a transaction as used in the database.
// It opens a database connection, prepares an SQL UPDATE statement to set the 'spent'
// field of a transaction to true based on the provided transaction ID, and executes
// the statement. If any operation fails, it logs the error.
func MarkTransactionUsed(transactionID string) error {
	db, err := openDatabase()
	if err != nil {
		return err
	}
	defer db.Close()

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	queryTxt := "UPDATE transactions SET spent = true WHERE id = ?"
	_, err = tx.Exec(queryTxt, transactionID)
	if err != nil {
		fmt.Println(err)
		return err
	}

	tx.Commit()
	db.Close()

	return nil
}

// openDatabase opens a connection to the database using the specified database driver
// and path. It returns a pointer to the sql.DB object representing the database connection
// and an error if any occurs during the process.
func openDatabase() (*sql.DB, error) {
	database, err := sql.Open(databaseDriver, databasePath)
	if err != nil {
		fmt.Println("Error while opening database: ", err)
		return nil, err
	}
	return database, nil
}

// generateTransactionID generates a unique transaction ID by creating a random 16-byte
// sequence and encoding it as a hexadecimal string. This ID can be used as a unique identifier
// for transactions in the database.
func generateTransactionID() (string, error) {
	bytes := make([]byte, 16) // Generate  16 random bytes
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
