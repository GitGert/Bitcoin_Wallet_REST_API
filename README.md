# Bitcoin_Wallet_REST_API


## Table of Contents

- [Introduction](#introduction)
- [Features](#features)
- [Installation](#installation) //todo: mby move some things here from usage
- [Usage](#usage)
- [Author](#uthor)

## Introduction

Bitcoin_Wallet_REST_API is a simple Bitcoin Wallet REST API based on a simplified Bitcoin transaction model.

## Features

- API endpoint  1: /listTransactions will provide JSON data of all the transactions that are stored in the database.
- API endpoint  2: /showBalance will provide JSON data of the balance that is usable in the current account.
- API endpoint  3: /spendBalance?amount=[insert number] will "spend" x amount of euros from the account.

## Installation

``` bash
git clone https://github.com/GitGert/Bitcoin_Wallet_REST_API
```

## Usage

In order to run the application you will need to install [golang](https://go.dev/doc/install) and [SQLite](https://www.sqlite.org/download.html)

* navigate to the root folder:
``` bash
cd Bitcoin_Wallet_REST_API
```

* start the server:
``` bash
go run main.go
```
golangs dependencies should resolve themselves when running "main.go" for the first time.
* to test you will need to either use a browser or a tool like curl (linux/IOS)

LINKS:
- http://localhost:8080/listTransactions
- http://localhost:8080/showBalance
- http://localhost:8080/spendBalance?amount=100

you can change the amount of euros you wish to spend by changing the number
after "/spendBalance?amount="


### Author

[Gert Nõgene](https://github.com/GitGert/)