package main

import (
	"crypto/rand"
	"database/sql" //TODO: uncomment
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// use interfaces

// TODO:
// create database table with the given rows.
// create mock data, maybe save them.
type Transaction struct {
	Transaction_ID string
	Amount         float64
	Spent          bool
	Created_at     time.Time
}

var databaseDriver = "sqlite3"
var databasePath = "bitcoin_wallet.db"

func main() {
	fmt.Println("wat")
	mux := http.NewServeMux()

	mux.HandleFunc("/ListTransactions", List_Transactions) //List All transactions
	mux.HandleFunc("/ShowBalance", ShowBalance)            // show current balance in BTC and EUR
	mux.HandleFunc("/SpendBalance", SpendBalance)          // new transfer, input data in EUR

	fmt.Println("started server at http://localhost:8080")
	http.ListenAndServe(":8080", mux) //define REST API endpoint
}

func List_Transactions(w http.ResponseWriter, r *http.Request) {
	//query all of the transactions from the database and send them as JSON
	Transactions := getAllTransactions()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Transactions)
}

func ShowBalance(w http.ResponseWriter, r *http.Request) {

	//for showing balance I will need to read all of the transactions and calculate the amount that is left over.
	// if transaction.Spent == False then + else -
	Transaction := getAllTransactions()
	var TotalAmountOfBitcoin float64
	for _, transaction := range Transaction {
		if !transaction.Spent {
			TotalAmountOfBitcoin += transaction.Amount
		}
		//  else {
		// 	TotalAmountOfBitcoin -= transaction.Amount
		// }

		fmt.Println(transaction)
		//TODO: add the calculation logic here.
	}
	data := getBitcoinValue()
	TotalAmountInEUR := data * TotalAmountOfBitcoin
	TotalAmountInEUR = math.Round(TotalAmountInEUR*100) / 100
	float64AmountAsString := fmt.Sprintf("%f", TotalAmountOfBitcoin)
	TotalAmountInEUR_as_String := fmt.Sprintf("%.2f", TotalAmountInEUR)
	balanceReport := "You have " + float64AmountAsString + "of Bitcoin that equals to " + TotalAmountInEUR_as_String + "€"

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(balanceReport)

}

type Ticker struct {
	Data []struct {
		Symbol    string `json:"symbol"`
		Value     string `json:"value"`
		Sources   int    `json:"sources"`
		UpdatedAt string `json:"updated_at"`
	} `json:"data"`
}

func getBitcoinValue() float64 {
	api_link := "http://api-cryptopia.adca.sh/v1/prices/ticker"

	response, err := http.Get(api_link)
	ErrorHandler(err)

	defer response.Body.Close()

	ticker := &Ticker{}
	// Read the response body
	body, err := io.ReadAll(response.Body)
	ErrorHandler(err)
	fmt.Println("tikcer api response: ", string(body))
	json.Unmarshal(body, ticker)
	fmt.Println(ticker.Data)
	fmt.Println(ticker.Data[0].Value)
	fmt.Println(ticker.Data[0].Symbol)

	//TODO: VERY IMPORTANT, add the cases wehre an error message is sent instead of the success json.

	bitcoinValueAsFloat64, err := strconv.ParseFloat(ticker.Data[0].Value, 64)
	ErrorHandler(err)
	return bitcoinValueAsFloat64
}

// http://localhost:8080/SpendBalance?amount=50
func SpendBalance(w http.ResponseWriter, r *http.Request) {
	request_value_string := r.URL.Query().Get("amount")

	if request_value_string == "" {
		//TODO: send back an error message that the value is empty.
		fmt.Println("No EUR value provided")
		return
	} else if !isStringValidEURValue(request_value_string) {
		fmt.Println("Error: The input must have exactly  2 decimal points.")
		return
	}
	//validate the request, it needs to be a valid EUR value

	request_value_as_float64, err := strconv.ParseFloat(request_value_string, 64)
	if err != nil {
		fmt.Println("Error parsing string to float64:", err)
		return
	}
	bitcoinvalueAsFloat64 := getBitcoinValue()

	request_value_in_Bitcoin := request_value_as_float64 / bitcoinvalueAsFloat64

	if request_value_as_float64 < 0.00001 {
		fmt.Println("transfer amount cannot be smaller than 0.00001 BTC")
		return
	}

	allTransactions := getAllTransactions()

	unspentTransactionsIndexes := []int{}
	var unspentMoneyTotal float64
	for i, transaction := range allTransactions {
		if !transaction.Spent {
			unspentMoneyTotal += transaction.Amount
			unspentTransactionsIndexes = append(unspentTransactionsIndexes, i)
		}
		if unspentMoneyTotal >= request_value_in_Bitcoin {
			//if t
			break
		}
	}

	fmt.Println("unspentMoneyTotal:")
	fmt.Println(unspentMoneyTotal)
	fmt.Println("request_value_in_Bitcoin:")
	fmt.Println(request_value_in_Bitcoin)

	if unspentMoneyTotal < request_value_in_Bitcoin {
		fmt.Println("not enough funds")
		return
		//TODO: send error to client.
	}

	//If ther is enough money, calculate the differnece, mark everything as used using the indexes and create new transaction with the difference.

	difference := unspentMoneyTotal - request_value_in_Bitcoin

	for _, index_value := range unspentTransactionsIndexes {
		transactionID := allTransactions[index_value].Transaction_ID
		mark_Transaction_Used(transactionID)
	}
	fmt.Println("difference: ", difference)
	if difference != 0.0 {
		create_New_Transaction(difference)
	}

	//if the transfer amount is smaller than 0.00001 BTC, then the API should reject the request.
	//TODO:
	// get the value in eur from the request
	// validate the request via database
	// handle errors
	// mark fields spent in databse
	// create leftover transaction.
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(request_value_string + " EUR has been spent")
}

func create_New_Transaction(value float64) {
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

func mark_Transaction_Used(transactionID string) {
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

func isStringValidEURValue(input string) bool {
	// Regular expression to match a string with exactly  2 decimal points
	re := regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

	// Check if the input matches the regular expression
	return re.MatchString(input)
}

// example func
func (u *Transaction) Insert() error {
	// Database insertion logic here
	// ...
	return nil
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

func getAllTransactions() []Transaction {
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

// generateTransactionID generates a random unique hexadecimal string.
func generateTransactionID() (string, error) {
	bytes := make([]byte, 16) // Generate  16 random bytes
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func ErrorHandler(err error) {
	fmt.Println(err)
	print("exiting now...")
	// os.Exit(1)
}
