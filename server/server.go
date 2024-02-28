// Package server provides a REST API with three endpoints: /listTransactions , /showBalance, /spendBalance at localhost:8080
package server

import (
	db "bitcoin_wallet_rest_api/database"
	transactionTypes "bitcoin_wallet_rest_api/transactionTypes"
	"fmt"
	"math"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// Init initializes and starts the HTTP server for the application.
// It sets up the server's routing and starts listening on port 8080.
func Init() {
	mux := http.NewServeMux()

	mux.HandleFunc("/listTransactions", listTransactions) // List all transactions
	mux.HandleFunc("/showBalance", showBalance)           // Show current balance in BTC and EUR
	mux.HandleFunc("/spendBalance", spendBalance)         // New transfer, input data in EUR
	mux.HandleFunc("/addBalance", addBalance)             // New transfer, input data in EUR

	fmt.Println("starting server at \033[32mhttp://localhost:8080\033[0m")
	fmt.Println("In order to stop the server use :\033[31m CTRL + C\033[0m")

	http.ListenAndServe(":8080", mux) //define REST API endpoint
}

// listTransactions retrieves and returns a list of all transactions from the database.
// It handles the HTTP request and response, sending back a JSON response with the transactions
// or an error message if the database operation fails.
func listTransactions(w http.ResponseWriter, r *http.Request) {

	var transactions, err = db.GetAllTransactions()

	if err != nil {
		fmt.Println("Database error:", err)
		sendInternalServerErrorResponse(w)
		return
	}

	response := transactionTypes.APIResponse{
		Data: transactions,
	}
	sendHTTPResponse(w, response, http.StatusOK)
}

// showBalance calculates and returns the total balance of Bitcoin in EUR for the user.
// It fetches all transactions from the database, filters out spent transactions,
// calculates the total amount of Bitcoin, fetches the current Bitcoin value in EUR,
// and then calculates the total balance in EUR. The result is sent back as a JSON response.
func showBalance(w http.ResponseWriter, r *http.Request) {

	allTransactions, err := db.GetAllTransactions()

	if err != nil {
		fmt.Println("Database error:", err)
		sendInternalServerErrorResponse(w)
		return
	}

	totalAmountOfBitcoin := getTotalAmountOfBitcoin(allTransactions)
	BTCToEURExchangeRate, err := getBitcoinValue()

	if err != nil {
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusServiceUnavailable)
		return
	}

	totalAmountInEUR := math.Round(BTCToEURExchangeRate*totalAmountOfBitcoin*100) / 100

	bitcoinTotalString := fmt.Sprintf("%f", totalAmountOfBitcoin)
	totalAmountInEURString := fmt.Sprintf("%.2f", totalAmountInEUR)

	balanceReport := "You have " + bitcoinTotalString + " of Bitcoin that equals to " + totalAmountInEURString + "€"

	response := transactionTypes.APIResponse{
		Data: balanceReport,
	}
	sendHTTPResponse(w, response, http.StatusOK)
}

// spendBalance processes a request to spend a specified amount of EUR in Bitcoin.
// It calculates the equivalent amount of Bitcoin for the given EUR value, checks if
// there are sufficient unspent transactions to cover the request, and if so, marks
// the necessary transactions as spent and creates a new transaction for any leftover
// Bitcoin. The function sends back a JSON response indicating the success or failure
// of the operation.
func spendBalance(w http.ResponseWriter, r *http.Request) {

	request_value_string := r.URL.Query().Get("amount")

	if request_value_string == "" {
		fmt.Println("Error - No EUR value provided")
		response := transactionTypes.APIResponse{
			Data:   "please provide a value",
			Errors: []string{"Error - No EUR value provided"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	if !isStringValidEURValue(request_value_string) {
		fmt.Println("Error: poor input value")
		response := transactionTypes.APIResponse{
			Data:   "A poor value was provided - please make sure to provide a valid number",
			Errors: []string{"Error - Input value was improper"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	requestValueAsFloat, err := strconv.ParseFloat(request_value_string, 64)
	if err != nil {
		fmt.Println("Error parsing string to float64:", err)
		sendInternalServerErrorResponse(w)
		return
	}

	bitcoinvalueAsFloat64, err := getBitcoinValue()
	if err != nil {
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusServiceUnavailable)
		return
	}

	requestValueInBitcoin := requestValueAsFloat / bitcoinvalueAsFloat64

	if requestValueAsFloat < 0.00001 {
		fmt.Println("Bitcoin value to small")
		response := transactionTypes.APIResponse{
			Data:   "Bad Request - The minimum amount for a transfer is 0.00001 BTC",
			Errors: []string{"Error - Bad Request. BTC amount cannot be smaller than 0.00001"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	allTransactions, err := db.GetAllTransactions()
	if err != nil {
		fmt.Println("Database error:", err)
		sendInternalServerErrorResponse(w)
		return
	}

	unspentTransactionsIndexesList, unspentMoneyTotal := getUnspentTranactionsSumAndIndexeses(allTransactions, requestValueInBitcoin)

	if unspentMoneyTotal < requestValueInBitcoin {
		fmt.Println("not enough funds")
		response := transactionTypes.APIResponse{
			Data:   "Insufficient funds",
			Errors: []string{"You do not have enough Bitcoin to cover this transaction."},
		}
		sendHTTPResponse(w, response, http.StatusForbidden)
		return
	}

	difference := unspentMoneyTotal - requestValueInBitcoin

	if err = markTransactionsUsed(unspentTransactionsIndexesList, allTransactions); err != nil {
		fmt.Println("Error while trying to mark transactions as spent: ", err)
		sendInternalServerErrorResponse(w)
		return
	}

	if difference != 0.0 {
		if err = db.CreateNewTransaction(difference); err != nil {
			print("error creating transaction: ", err)
			sendInternalServerErrorResponse(w)
			return
		}
	}

	response := transactionTypes.APIResponse{
		Data: request_value_string + " EUR has been spent",
	}

	sendHTTPResponse(w, response, http.StatusOK)
}

// addBalance processes requests to add a specified amount of EUR to the Bitcoin wallet.
// It calculates the equivalent amount of Bitcoin for the given EUR value, fetches
// the current Bitcoin to EUR exchange rate, and creates a new transaction in the database.
func addBalance(w http.ResponseWriter, r *http.Request) {
	request_value_string := r.URL.Query().Get("amount") //input value will be in EUR

	if request_value_string == "" {
		fmt.Println("Error - No EUR value provided")
		response := transactionTypes.APIResponse{
			Data:   "please provide a value",
			Errors: []string{"Error - No EUR value provided"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	if !isStringValidEURValue(request_value_string) {
		fmt.Println("Error: poor input value")
		response := transactionTypes.APIResponse{
			Data:   "A poor value was provided - please make sure to provide a valid number",
			Errors: []string{"Error - Input value was improper"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	requestValueAsFloat, err := strconv.ParseFloat(request_value_string, 64)
	if err != nil {
		fmt.Println("Error parsing string to float64:", err)
		sendInternalServerErrorResponse(w)
		return
	}

	bitcoinvalueAsFloat64, err := getBitcoinValue()
	if err != nil {
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusServiceUnavailable)
		return
	}

	requestValueInBitcoin := requestValueAsFloat / bitcoinvalueAsFloat64

	if requestValueAsFloat < 0.00001 {
		fmt.Println("input amount cannot be smaller than 0.00001 BTC")
		response := transactionTypes.APIResponse{
			Data:   "Bad Request - The minimum amount for a transfer is 0.00001 BTC",
			Errors: []string{"Error - Bad Request. BTC amount cannot be smaller than 0.00001"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	if err = db.CreateNewTransaction(requestValueInBitcoin); err != nil {
		print("error creating transaction: ", err)
		sendInternalServerErrorResponse(w)
	}

	response := transactionTypes.APIResponse{
		Data: request_value_string + " EUR has been added",
	}

	sendHTTPResponse(w, response, http.StatusOK)
}
