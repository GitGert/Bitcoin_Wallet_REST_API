package databaseFunctions

import (
	transaction_types "bitcoin_wallet_rest_api/transaction_types"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"time"
)

type Transaction = transaction_types.Transaction //TODO: maybe change this later
var databaseDriver = "sqlite3"
var databasePath = "bitcoin_wallet.db"

func Create_New_Transaction(value float64) {
	db := openDatabase()
	defer db.Close()

	newTransactionID, err := generateTransactionID()
	ErrorHandler(err)

	insertTransactionQuery := `
	INSERT INTO transactions (id, amount, spent, created_at)
	VALUES (?, ?, ?, ?);
	`
	current_time := time.Now()

	_, err = db.Exec(insertTransactionQuery, newTransactionID, value, false, current_time)
	ErrorHandler(err)
}

func GetAllTransactions() []Transaction {
	db := openDatabase()
	defer db.Close()

	allTransactions := []Transaction{}

	queryTxt := `SELECT * FROM transactions`
	rows, err := db.Query(queryTxt)

	if err != nil {
		fmt.Println(err)
	}

	// err = rows.Scan(&transaction.Transaction_ID, &transaction.Amount, &transaction.Spent, &transaction.Created_at)
	for rows.Next() {
		transaction := Transaction{}
		err = rows.Scan(&transaction.Transaction_ID, &transaction.Amount, &transaction.Spent, &transaction.Created_at)
		if err != nil {
			fmt.Println(err)
		}
		allTransactions = append(allTransactions, transaction)
	}

	if err != nil {
		fmt.Println("Error while scanning transaction: ", err)
		os.Exit(1)
	}

	// fmt.Println(allTransactions)

	db.Close()

	return allTransactions
}

func Mark_Transaction_Used(transactionID string) {
	db := openDatabase()
	defer db.Close()

	stmt, err := db.Prepare("UPDATE transactions SET spent = true WHERE id = ?") //TODO: if I remember correctly this prepare thing is not really needed if I only call this once with one value
	//hovever if I change this so the db connection will be open in the loop and I can use the same db connection it will be more optimized... TLDR: look into this later.
	ErrorHandler(err)
	defer stmt.Close()

	// Execute the statement.
	_, err = stmt.Exec(transactionID)
	ErrorHandler(err)
}

// TODO: work on this.
func openDatabase() *sql.DB {
	database, err := sql.Open(databaseDriver, databasePath)
	if err != nil {
		fmt.Println("Error opening database: ", err)
	}
	// defer database.Close()

	return database
}

func ErrorHandler(err error) {
	fmt.Println(err)
	print("exiting now...")
	// os.Exit(1)
}

// generateTransactionID generates a random unique hexadecimal string.
func generateTransactionID() (string, error) {
	bytes := make([]byte, 16) // Generate  16 random bytes
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
