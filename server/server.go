// Package server provides a REST API with three endpoints: /listTransactions , /showBalance, /spendBalance at localhost:8080
package server

import (
	db "bitcoin_wallet_rest_api/database"
	financialDataTypes "bitcoin_wallet_rest_api/financialDataTypes"
	transactionTypes "bitcoin_wallet_rest_api/transactionTypes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
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
	// EXAMPLE:
	// http://localhost:8080/listTransactions

	var transactions, err = db.GetAllTransactions()
	if err != nil {
		fmt.Println("Database error:", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Database Error",
			Errors: []string{"Database Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
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
	// EXAMPLE:
	// http://localhost:8080/showBalance

	transaction, err := db.GetAllTransactions()
	if err != nil {
		fmt.Println("Database error:", err)

		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Database Error",
			Errors: []string{"Database Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	var totalAmountOfBitcoin float64

	for _, transaction := range transaction {
		if !transaction.Spent {
			totalAmountOfBitcoin += transaction.Amount
		}
	}

	data, err := getBitcoinValue()
	if err != nil {
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	totalAmountInEUR := math.Round(data*totalAmountOfBitcoin*100) / 100

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
	// EXAMPLE:
	// http://localhost:8080/spendBalance?amount=50

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
		return
	}
	bitcoinvalueAsFloat64, err := getBitcoinValue()
	if err != nil {
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	requestValueInBitcoin := requestValueAsFloat / bitcoinvalueAsFloat64

	if requestValueAsFloat < 0.00001 {
		fmt.Println("transfer amount cannot be smaller than 0.00001 BTC")
		fmt.Println("GetBitcoinValue Error: ", err)
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

		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Database Error",
			Errors: []string{"Database Error"},
		}

		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	unspentTransactionsIndexes := []int{}
	var unspentMoneyTotal float64

	for i, transaction := range allTransactions {
		if !transaction.Spent {
			unspentMoneyTotal += transaction.Amount
			unspentTransactionsIndexes = append(unspentTransactionsIndexes, i)
		}
		if unspentMoneyTotal >= requestValueInBitcoin {
			break
		}
	}

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

	for _, index_value := range unspentTransactionsIndexes {
		transactionID := allTransactions[index_value].TransactionID
		err = db.MarkTransactionUsed(transactionID)

		if err != nil {
			fmt.Println("Error while trying to mark transactions as spent: ", err)

			response := transactionTypes.APIResponse{
				Data:   "Database error",
				Errors: []string{"Error - Failed to edit transaction"},
			}

			sendHTTPResponse(w, response, http.StatusInternalServerError)
			return
		}
	}
	fmt.Println("difference: ", difference)
	if difference != 0.0 {
		err = db.CreateNewTransaction(difference)
		if err != nil {
			print("error creating transaction: ", err)
			response := transactionTypes.APIResponse{
				Data:   "Internal Server Error - Database Error",
				Errors: []string{"Error - Database Error"},
			}

			sendHTTPResponse(w, response, http.StatusInternalServerError)
			return
		}
	}

	response := transactionTypes.APIResponse{
		Data: request_value_string + " EUR has been spent",
	}

	sendHTTPResponse(w, response, http.StatusOK)
}

// addBalance is an HTTP handler that processes requests to add a specified amount
// of EUR to the Bitcoin wallet. It calculates the equivalent amount of Bitcoin for the
// given EUR value, fetches the current Bitcoin to EUR exchange rate, and creates a
// new transaction in the database.
func addBalance(w http.ResponseWriter, r *http.Request) {
	// EXAMPLE:
	// http://localhost:8080/addBalance?amount=50
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
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	bitcoinvalueAsFloat64, err := getBitcoinValue()
	if err != nil {
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	requestValueInBitcoin := requestValueAsFloat / bitcoinvalueAsFloat64

	if requestValueAsFloat < 0.00001 {
		fmt.Println("input amount cannot be smaller than 0.00001 BTC")
		fmt.Println("GetBitcoinValue Error: ", err)
		response := transactionTypes.APIResponse{
			Data:   "Bad Request - The minimum amount for a transfer is 0.00001 BTC",
			Errors: []string{"Error - Bad Request. BTC amount cannot be smaller than 0.00001"},
		}
		sendHTTPResponse(w, response, http.StatusBadRequest)
		return
	}

	if err = db.CreateNewTransaction(requestValueInBitcoin); err != nil {
		print("error creating transaction: ", err)

		response := transactionTypes.APIResponse{
			Data:   "Internal Server Error - Database Error",
			Errors: []string{"Error - Database Error"},
		}

		sendHTTPResponse(w, response, http.StatusInternalServerError)
	}

	response := transactionTypes.APIResponse{
		Data: request_value_string + " EUR has been added",
	}

	sendHTTPResponse(w, response, http.StatusOK)
}

// getBitcoinValue fetches the current Bitcoin to EUR exchange rate from an external API.
// It sends an HTTP GET request to the specified API endpoint, parses the JSON response,
// and extracts the Bitcoin to EUR exchange rate. The function returns the exchange rate
// as a float64 value and an error if any occurs during the process.
func getBitcoinValue() (float64, error) {
	var bitcoinValueAsFloat64 float64

	api_link := "http://api-cryptopia.adca.sh/v1/prices/ticker"

	response, err := http.Get(api_link)

	if err != nil {
		return 0, err
	}

	if response.StatusCode != 200 {
		return 0, err
	}

	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)

	if err != nil {
		return 0, err
	}

	ticker := financialDataTypes.Ticker{}
	err = json.Unmarshal(body, &ticker)

	if err != nil {
		return 0, err
	}

	// find the BTC/EUR field and save the value
	for _, v := range ticker.Data {
		if v.Symbol == "BTC/EUR" {
			bitcoinValueAsFloat64, err = strconv.ParseFloat(v.Value, 64)
			if err != nil {
				return 0, err
			}
			break
		}
	}

	return bitcoinValueAsFloat64, nil
}

// isStringValidEURValue checks if a given string represents a valid EUR value.
// It uses a regular expression to validate that the string is a positive number
// with up to two decimal places, which is the standard format for EUR values.
func isStringValidEURValue(input string) bool {
	re := regexp.MustCompile(`^\d+(\.\d{1,2})?$`) //match a string with 1 or 2 decimal points
	return re.MatchString(input)
}

// sendHTTPResponse sends an HTTP response with a given status code and JSON body.
// It sets the "Content-Type" header to "application/json", writes the status code,
// and encodes the provided response object into JSON format before sending it to the client.
func sendHTTPResponse(w http.ResponseWriter, response transactionTypes.APIResponse, httpSatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpSatus)
	// json.NewEncoder(w).Encode(response)
	jsonData, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}
