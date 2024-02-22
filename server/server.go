// Package server provides a REST API with three endpoints: /listTransactions , /showBalance, /spendBalance at localhost:8080
package server

import (
	db "bitcoin_wallet_rest_api/database"
	financial_data_types "bitcoin_wallet_rest_api/financial_data_types"
	transaction_types "bitcoin_wallet_rest_api/transaction_types"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

func Init() {
	mux := http.NewServeMux()

	mux.HandleFunc("/listTransactions", listTransactions) //List All transactions
	mux.HandleFunc("/showBalance", showBalance)           // show current balance in BTC and EUR
	mux.HandleFunc("/spendBalance", spendBalance)         // new transfer, input data in EUR

	fmt.Println("starting server at http://localhost:8080")
	http.ListenAndServe(":8080", mux) //define REST API endpoint
}

func listTransactions(w http.ResponseWriter, r *http.Request) {
	var transactions, err = db.GetAllTransactions()
	if err != nil {
		fmt.Println("Database error:", err)
		response := transaction_types.APIResponse{
			Data:   "Internal Server Error - Database Error",
			Errors: []string{"Database Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	response := transaction_types.APIResponse{
		Data: transactions,
	}
	sendHTTPResponse(w, response, http.StatusOK)
}

func showBalance(w http.ResponseWriter, r *http.Request) {
	//for showing balance I will need to read all of the transactions and calculate the amount that is left over.
	transaction, err := db.GetAllTransactions()
	if err != nil {
		fmt.Println("Database error:", err)

		response := transaction_types.APIResponse{
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
		response := transaction_types.APIResponse{
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

	response := transaction_types.APIResponse{
		Data: balanceReport,
	}
	sendHTTPResponse(w, response, http.StatusOK)
}

// http://localhost:8080/SpendBalance?amount=50
func spendBalance(w http.ResponseWriter, r *http.Request) {
	request_value_string := r.URL.Query().Get("amount")
	//TODO handle error

	if request_value_string == "" {
		fmt.Println("Error - No EUR value provided")
		//TODO: send back an error message that the value is empty.
		response := transaction_types.APIResponse{
			Data:   "please provide a value",
			Errors: []string{"Error - No EUR value provided"},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)

		return
	}
	if !isStringValidEURValue(request_value_string) {
		fmt.Println("Error: poor input value")
		response := transaction_types.APIResponse{
			Data:   "A poor value was provided - please make sure to provide a valid number",
			Errors: []string{"Error - Input value was improper"},
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
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
		response := transaction_types.APIResponse{
			Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
			Errors: []string{"Internal Server Error"},
		}
		sendHTTPResponse(w, response, http.StatusInternalServerError)
		return
	}

	requestValueInBitcoin := requestValueAsFloat / bitcoinvalueAsFloat64

	if requestValueAsFloat < 0.00001 {
		fmt.Println("transfer amount cannot be smaller than 0.00001 BTC")
		return
	}

	allTransactions, err := db.GetAllTransactions()
	if err != nil {
		fmt.Println("Database error:", err)

		response := transaction_types.APIResponse{
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
			//if t
			break
		}
	}

	fmt.Println("unspentMoneyTotal:")
	fmt.Println(unspentMoneyTotal)
	fmt.Println("requestValueInBitcoin:")
	fmt.Println(requestValueInBitcoin)

	if unspentMoneyTotal < requestValueInBitcoin {
		fmt.Println("not enough funds")
		return
		//TODO: send error to client.
	}

	//If ther is enough money, calculate the differnece, mark everything as used using the indexes and create new transaction with the difference.

	difference := unspentMoneyTotal - requestValueInBitcoin

	for _, index_value := range unspentTransactionsIndexes {
		transactionID := allTransactions[index_value].Transaction_ID
		db.Mark_Transaction_Used(transactionID)
	}
	fmt.Println("difference: ", difference)
	if difference != 0.0 {
		db.CreateNewTransaction(difference)
	}

	//if the transfer amount is smaller than 0.00001 BTC, then the API should reject the request.
	//TODO:
	// get the value in eur from the request
	// validate the request via database
	// handle errors
	// mark fields spent in databse
	// create leftover transaction.
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(request_value_string + " EUR has been spent")
}

func getBitcoinValue() (float64, error) {
	api_link := "http://api-cryptopia.adca.sh/v1/prices/ticker"

	response, err := http.Get(api_link)
	if err != nil {
		return 0, err
	}
	if response.StatusCode != 200 {
		return 0, err
	}
	defer response.Body.Close()

	ticker := financial_data_types.Ticker{}
	// Read the response body
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, err
	}
	// fmt.Println("tikcer api response: ", string(body))
	json.Unmarshal(body, &ticker)

	bitcoinValueAsFloat64, err := strconv.ParseFloat(ticker.Data[0].Value, 64)
	if err != nil {
		return 0, err
	}
	return bitcoinValueAsFloat64, nil
}

// This is good, I should use this one and remove the other onese that take uneccessary space
func sendHTTPResponse(w http.ResponseWriter, response transaction_types.APIResponse, httpSatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpSatus)
	json.NewEncoder(w).Encode(response)
}

func ErrorHandler(err error) {
	if err != nil {
		fmt.Println(err)
		print("exiting now...")
		// os.Exit(1)
	}
}

func isStringValidEURValue(input string) bool {
	re := regexp.MustCompile(`^\d+(\.\d{1,2})?$`) //match a string with 1 or 2 decimal points
	return re.MatchString(input)
}

func sendSuccessfulHTTPresponse(w http.ResponseWriter, data string, errors []string) { // THIS IS DEPRECATED
	response := transaction_types.APIResponse{
		Data:   data,
		Errors: errors,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func sendHTTPResponseAfterDBError(w http.ResponseWriter) { // THIS IS DEPRECATED
	response := transaction_types.APIResponse{
		Data:   "Internal Server Error - Database Error",
		Errors: []string{"Database Error"},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}

func sendHTTPResponseAfterGetBitcoinValueError(w http.ResponseWriter) { // THIS IS DEPRECATED
	response := transaction_types.APIResponse{
		Data:   "Internal Server Error - Failed to fetch Bitcoin Value",
		Errors: []string{"Internal Server Error"},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusInternalServerError)
	json.NewEncoder(w).Encode(response)
}

// /////////////////////////////////
// data := map[string]string{"message": "Hello, world!"}

// // Create the API response
// response := transaction_types.APIResponse{
// 	Data: data,
// }

// // Set the content type to application/json
// w.Header().Set("Content-Type", "application/json")

// // Marshal the response into JSON and write it to the response writer
// json.NewEncoder(w).Encode(response)
// /////////////////////////////////////////////
