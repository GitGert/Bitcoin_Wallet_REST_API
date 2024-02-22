// Package financial_data_types currently provides just the Ticker struct
package financial_data_types

// Ticker represents the structure of a successful response from the http://api-cryptopia.adca.sh/v1/prices/ticker endpoint.
type Ticker struct {
	Data []struct {
		Symbol    string `json:"symbol"`
		Value     string `json:"value"`
		Sources   int    `json:"sources"`
		UpdatedAt string `json:"updated_at"`
	} `json:"data"`
}
