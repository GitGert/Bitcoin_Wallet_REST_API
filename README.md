# Bitcoin_Wallet_REST_API

Bitcoin_Wallet_REST_API is a simple Bitcoin Wallet REST API based on a simplified Bitcoin transaction model.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
- [Author](#uthor)


## Features

- API endpoint  1: /listTransactions - will provide JSON data of all the transactions that are stored in the database.
- API endpoint  2: /showBalance - will provide JSON data of the balance that is usable in the current account.
- API endpoint  3: /spendBalance?amount=[insert number] - will "spend" x amount of euros from the account.
- API endpoint  4: /addBalance?amount=[insert number] - will add x amount of euros to the account.

## Installation

``` bash
git clone https://github.com/GitGert/Bitcoin_Wallet_REST_API
```

## Usage

In order to run the application you will need have [Golang](https://go.dev/doc/install) and [SQLite](https://www.sqlite.org/download.html) installed on your machine.

* navigate to the root folder:
``` bash
cd Bitcoin_Wallet_REST_API
```

* start the server:
``` bash
go run main.go
```
Golangs dependencies will resolve themselves when running "main.go" for the first time.
* to test you will need to either use a browser or a CLI tool like [cURL](https://blog.hubspot.com/website/curl-command)

API ENDPOINT LINKS:
- http://localhost:8080/listTransactions
- http://localhost:8080/showBalance
- http://localhost:8080/spendBalance?amount=100
- http://localhost:8080/addBalance?amount=100

you can change the amount of euros you wish to add/spend by changing the number
after "?amount="

### Author

[Gert Nõgene](https://github.com/GitGert/)