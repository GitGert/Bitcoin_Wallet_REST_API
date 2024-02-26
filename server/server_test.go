// package server

// import (
// 	transactionTypes "bitcoin_wallet_rest_api/transaction_types"
// 	"github.com/stretchr/testify/assert"
// 	"github.com/stretchr/testify/mock"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"
// )

// // MockDB is a mock for the database layer.
// type MockDB struct {
// 	mock.Mock
// }

// func (m *MockDB) GetAllTransactions() ([]transactionTypes.Transaction, error) {
// 	args := m.Called()
// 	return args.Get(0).([]transactionTypes.Transaction), args.Error(1)
// }

// // MockResponseSender is a mock for the sendHTTPResponse function.
// type MockResponseSender struct {
// 	mock.Mock
// }

// func (m *MockResponseSender) sendHTTPResponse(w http.ResponseWriter, response transactionTypes.APIResponse, statusCode int) {
// 	m.Called(w, response, statusCode)
// }

// // TestListTransactions tests the listTransactions function.
// func TestListTransactions(t *testing.T) {
// 	// Create a mock database.
// 	db := new(MockDB)
// 	// Create a mock response sender.
// 	responseSender := new(MockResponseSender)

// 	// Set up the mock database to return a list of transactions.
// 	db.On("GetAllTransactions").Return([]transactionTypes.Transaction{
// 		{Transaction_ID: "1", Amount: 100.0, Spent: false, Created_at: time.Now()},
// 	}, nil)

// 	// Set up the mock response sender to expect a call with the correct parameters.
// 	responseSender.On("sendHTTPResponse", mock.Anything, transactionTypes.APIResponse{
// 		Data: []transactionTypes.Transaction{
// 			{TransactionID: "1", Amount: 100.0, Spent: false, CreatedAt: "2023-04-01T10:00:00Z"},
// 		},
// 	}, http.StatusOK)

// 	// Create a request to pass to our handler.
// 	req, err := http.NewRequest("GET", "/listTransactions", nil)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	// Create a ResponseRecorder to record the response.
// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		listTransactions(w, r)
// 	})

// 	// Call the handler.
// 	handler.ServeHTTP(rr, req)

// 	// Assert the response status code is what we expect.
// 	assert.Equal(t, http.StatusOK, rr.Code, "handler returned wrong status code")

// 	// Assert the response body is what we expect.
// 	expected := `{"data":[{"transaction_id":"1","amount":100.0,"spent":false,"created_at":"2023-04-01T10:00:00Z"}]}`
// 	assert.Equal(t, expected, rr.Body.String(), "handler returned unexpected body")

// 	// Assert the mock database and response sender were called as expected.
// 	db.AssertExpectations(t)
// 	responseSender.AssertExpectations(t)
// }
