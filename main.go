package main

import (
	"database/sql" //TODO: uncomment
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
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

	json.Unmarshal(body, ticker)
	fmt.Println(ticker.Data)
	fmt.Println(ticker.Data[0].Value)
	fmt.Println(ticker.Data[0].Symbol)

	//TODO: VERY IMPORTANT, add the cases wehre an error message is sent instead of the success json.

	bitcoinValueAsFloat64, err := strconv.ParseFloat(ticker.Data[0].Value, 64)
	ErrorHandler(err)
	return bitcoinValueAsFloat64
}

func SpendBalance(w http.ResponseWriter, r *http.Request) {
	//TODO:
	// get the value in eur from the request
	// validate the request via database
	// handle errors
	// mark fields spent in databse
	// create leftover transaction.
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

func ErrorHandler(err error) {
	fmt.Println(err)
}
