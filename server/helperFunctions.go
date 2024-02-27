package server

import (
	db "bitcoin_wallet_rest_api/database"
	financialDataTypes "bitcoin_wallet_rest_api/financialDataTypes"
	transactionTypes "bitcoin_wallet_rest_api/transactionTypes"
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strconv"
)

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

// getTotalAmountOfBitcoin calculates the total amount of Bitcoin from a list of transactions,
// excluding those marked as spent.
func getTotalAmountOfBitcoin(transaction []transactionTypes.Transaction) float64 {
	var totalAmountOfBitcoin float64

	for _, transaction := range transaction {
		if !transaction.Spent {
			totalAmountOfBitcoin += transaction.Amount
		}
	}

	return totalAmountOfBitcoin
}

// sendInternalServerErrorResponse sends a  500 Internal Server Error response to the client
func sendInternalServerErrorReponse(w http.ResponseWriter) {
	response := transactionTypes.APIResponse{
		Data:   "Internal Server Error",
		Errors: []string{"Error - Internal Server Error"},
	}
	sendHTTPResponse(w, response, http.StatusInternalServerError)
}

// getUnspentTransactionsSumAndIndexes calculates the sum of unspent transactions and their indexes
// from a given list of transactions. It iterates through the transactions, summing up the amounts
// of unspent transactions (those marked as not spent) until the total amount reaches or exceeds
// the current transfer requests value in Bitcoin. The function then returns the indexes of these
// unspent transactions and the total sum of their amounts.
func getUnspentTranactionsSumAndIndexeses(allTransactions []transactionTypes.Transaction, requestValueInBitcoin float64) ([]int, float64) {
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

	return unspentTransactionsIndexes, unspentMoneyTotal
}

// markTransactionsUsed iterates over a list of indexes corresponding to unspent transactions and
// updates their status in the database.
func markTransactionsUsed(unspentTransactionsIndexesList []int, allTransactions []transactionTypes.Transaction) error {
	for _, index_value := range unspentTransactionsIndexesList {
		transactionID := allTransactions[index_value].TransactionID
		if err := db.MarkTransactionUsed(transactionID); err != nil {
			return err
		}
	}
	return nil
}

// sendHTTPResponse sends an HTTP response with a given status code and JSON body.
// It sets the "Content-Type" header to "application/json", writes the status code,
// and encodes the provided response object into JSON format before sending it to the client.
func sendHTTPResponse(w http.ResponseWriter, response transactionTypes.APIResponse, httpSatus int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpSatus)
	jsonData, err := json.MarshalIndent(response, "", " ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(jsonData)
}
