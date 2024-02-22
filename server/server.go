package server

import (
	db "bitcoin_wallet_rest_api/database"
	financial_data_types "bitcoin_wallet_rest_api/financial_data_types"
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

	mux.HandleFunc("/ListTransactions", List_Transactions) //List All transactions
	mux.HandleFunc("/ShowBalance", ShowBalance)            // show current balance in BTC and EUR
	mux.HandleFunc("/SpendBalance", SpendBalance)          // new transfer, input data in EUR

	fmt.Println("started server at http://localhost:8080")
	http.ListenAndServe(":8080", mux) //define REST API endpoint
}

func ShowBalance(w http.ResponseWriter, r *http.Request) {

	//for showing balance I will need to read all of the transactions and calculate the amount that is left over.
	// if transaction.Spent == False then + else -
	Transaction := db.GetAllTransactions()
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

	allTransactions := db.GetAllTransactions()

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
		db.Mark_Transaction_Used(transactionID)
	}
	fmt.Println("difference: ", difference)
	if difference != 0.0 {
		db.Create_New_Transaction(difference)
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

func List_Transactions(w http.ResponseWriter, r *http.Request) {
	//query all of the transactions from the database and send them as JSON
	Transactions := db.GetAllTransactions()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(Transactions)
}

func getBitcoinValue() float64 {
	api_link := "http://api-cryptopia.adca.sh/v1/prices/ticker"

	response, err := http.Get(api_link)
	ErrorHandler(err)

	defer response.Body.Close()

	ticker := &financial_data_types.Ticker{}
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

func ErrorHandler(err error) {
	fmt.Println(err)
	print("exiting now...")
	// os.Exit(1)
}

func isStringValidEURValue(input string) bool {
	// Regular expression to match a string with exactly  2 decimal points
	re := regexp.MustCompile(`^\d+(\.\d{1,2})?$`)

	// Check if the input matches the regular expression
	return re.MatchString(input)
}
