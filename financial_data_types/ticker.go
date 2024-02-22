package financial_data_types

type Ticker struct {
	Data []struct {
		Symbol    string `json:"symbol"`
		Value     string `json:"value"`
		Sources   int    `json:"sources"`
		UpdatedAt string `json:"updated_at"`
	} `json:"data"`
}
