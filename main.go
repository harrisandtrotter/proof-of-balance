package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/harrisandtrotter/proof-of-balance/backend/api"
	"github.com/harrisandtrotter/proof-of-balance/backend/blocks"
	"github.com/harrisandtrotter/proof-of-balance/backend/initialisers"
	"github.com/harrisandtrotter/proof-of-balance/backend/prices"
	"github.com/sqweek/dialog"
)

const (
	// Chain variables for Moralis API
	Ethereum          = "eth"
	Arbitrum          = "arbitrum"
	Polygon           = "polygon"
	BinanceSmartChain = "bsc"
	Fantom            = "fantom"
	Avalanche         = "avalanche"
	Cronos            = "cronos"

	// Token balance checker urls for revieiwing balance in CSV. Work-in-progress
	EthereumTokenChecker   = "https://etherscan.io/tokencheck-tool"
	ArbitrumTokenChecker   = "https://arbiscan.io/tokencheck-tool"
	PolygonTokenChecker    = "https://polygonscan.com/tokencheck-tool"
	BSCTokenChecker        = "https://bscscan.com/tokencheck-tool"
	FantomTokenChecker     = "https://ftmscan.com/tokencheck-tool"
	AvalancheTokenChecker  = "https://snowtrace.io/tokencheck-tool"
	CronosTokenChecker     = "https://cronoscan.com/tokencheck-tool"
	EthereumNativeChecker  = "https://etherscan.io/balancecheck-tool"
	ArbitrumNativeChecker  = "https://arbiscan.io/balancecheck-tool"
	PolygonNativeChecker   = "https://polygonscan.com/balancecheck-tool"
	BSCNativeChecker       = "https://bscscan.com/balancecheck-tool"
	FantomNativeChecker    = "https://ftmscan.com/balancecheck-tool"
	AvalancheNativeChecker = "https://snowtrace.io/balancecheck-tool"
	CronosNativeChecker    = "https://cronoscan.com/balancecheck-tool"

	// Native tokens
	ETH   = "ETH"
	ARB   = "ETH"
	BSC   = "BNB"
	FTM   = "FTM"
	MATIC = "MATIC"
	AVAX  = "AVAX"
	CRO   = "CRO"
)

var block blocks.Block

// Data structure
type TokenBalance struct {
	TokenAdress string `json:"token_address"`
	Name        string `json:"name"`
	Symbol      string `json:"symbol"`
	Logo        string `json:"logo"`
	Thumbnail   string `json:"thumbnail"`
	Decimals    int    `json:"decimals"`
	Balance     string `json:"balance"`
}

type NativeBalance struct {
	Balance string `json:"balance"`
}

type TokenFile struct {
	Address string
	Chain   string
	Block   string
}

type Block struct {
	Block          int    `json:"block"`
	Date           string `json:"date"`
	Timestamp      int    `json:"timestamp"`
	BlockTimestamp string `json:"block_timestamp"`
	Hash           string `json:"hash"`
	ParentHash     string `json:"parent_hash"`
}

func init() {
	initialisers.LoadEnvironment()
	initialisers.LoadAPIKey()
}

func main() {
	// for {
	// 	fmt.Println("\nWelcome to Proof-of-Balance. \nSee below for the available options. ")
	// 	fmt.Println("1. Block number")
	// 	fmt.Println("2. Retrieve balances")

	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Print("\nSelect the option you want to use today (e.g., 1 for receiving specific block number): \n")
	// 	input, _ := reader.ReadString('\n')

	// 	switch strings.TrimSpace(input) {
	// 	case "1":
	// 		fmt.Println("Coming soon.")
	// 	case "2":
	// 		fmt.Println("Please select the CSV file. \nThe format needs to be the wallet address in the column A and the relevant chain in column B. \n\nThe available chains and the required format are as follows: \neth\narbitrum\nfantom\nbsc\npolygon\navalanche\ncronos")
	// 		time.Sleep(time.Second * 1)
	// 		GetTokenBalance()
	// 	default:
	// 		fmt.Println("Invalid option selected.")
	// 	}

	// 	time.Sleep(time.Second * 3)
	// }

	api.Setup()

}

// Main get balance function
func GetTokenBalance() {
	// Get user to select CSV input file
	filename, err := dialog.File().Filter("CSV file", "csv").Load()
	if err != nil {
		fmt.Println("Error finding file:", &err)
	}

	// Parse inputted CSV to accessible data structure
	data := CsvToToken(filename)

	// Store user output filename in outputFile variable
	outputFile, err := dialog.File().Filter("CSV file", "csv").Title("Save the output file to CSV with your desired name.").Save()
	if err != nil {
		fmt.Println("Error saving file:", &err)
	}

	// Create output CSV file
	output, err := os.Create(outputFile + ".csv")
	if err != nil {
		log.Fatalf("Error creating csv file: %v", err)
	}

	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	// Create and write headers to output CSV file
	headers := []string{"Address", "Chain", "Token Name", "Token Symbol", "Token Address", "Balance", "Block number", "Token checker", "Usd rate", "Usd value"}
	err = writer.Write(headers)
	if err != nil {
		log.Fatalf("Error writing CSV headers: %v", err)
	}

	// Range over and access the data structure from "data" variable and assigned to "value" variable.
	for _, value := range data {
		// Used to store asset symbol for native tokens
		var asset string
		// Used to store balance checker url to print into the CSV
		var nativeTokenChecker string
		// Used to store native token name to print into the CSV
		var tokenName string
		// Used to access prices module to retrieve prices and calculate Usd value.
		var price prices.Price
		// Used to calculate USD value based on "price" variable
		var Uvalue float64

		// Get block number per chain per specified timestamp
		// block := GetBlock(value.Chain, Timestamp("31/12/2022 23:59:59 UTC"))
		blockNo := block.BlockNumber(value.Chain, block.TimestampToUnix("31/12/2022 23:59:59"))
		// Retrieve balance for native token per specified chain
		nativeResponse := getBalance(value.Address, value.Chain, strconv.Itoa(blockNo))

		// Parse the response for native token to an accessible native token data strucuture
		native := ResponseToNative(nativeResponse)

		// Logic to print native asset and balance checker url to the output CSV file.
		if value.Chain == Ethereum {
			asset = ETH
			nativeTokenChecker = EthereumNativeChecker
			tokenName = "Ethereum"
		} else if value.Chain == Arbitrum {
			asset = ARB
			nativeTokenChecker = ArbitrumNativeChecker
			tokenName = "Ethereum"
		} else if value.Chain == BinanceSmartChain {
			asset = BSC
			nativeTokenChecker = BSCNativeChecker
			tokenName = "Binance Coin"
		} else if value.Chain == Fantom {
			asset = FTM
			nativeTokenChecker = FantomNativeChecker
			tokenName = "Fantom"
		} else if value.Chain == Polygon {
			asset = MATIC
			nativeTokenChecker = PolygonNativeChecker
			tokenName = "Polygon (MATIC)"
		} else if value.Chain == Avalanche {
			asset = AVAX
			nativeTokenChecker = AvalancheNativeChecker
			tokenName = "Avalanche"
		} else if value.Chain == Cronos {
			asset = CRO
			nativeTokenChecker = CronosNativeChecker
			tokenName = "Cronos"
		}

		// fmt.Printf("chain: %v, Asset: %v, Balance: %v, address: %v, block: %v\n", value.Chain, asset, native.Balance, value.Address, block.Block)
		// Convert native token balance from string to number
		nativeString, err := strconv.ParseFloat(native.Balance, 64)
		if err != nil {
			fmt.Printf(`Error for address %v, when parsing asset %v, with a balance of "%v" on %v chain\n Error message below:\n.`, value.Address, asset, native.Balance, value.Chain)
			log.Fatalf("Error parsing native balance token: %v", err)
		}

		// Convert native token balance from wei to ether
		nativeToken := nativeString / math.Pow10(18)

		// Store values and write values for native token data
		nativeRecord := []string{value.Address, value.Chain, tokenName, asset, " ", fmt.Sprintf("%f", nativeToken), strconv.Itoa(blockNo), nativeTokenChecker, "", ""}
		err = writer.Write(nativeRecord)
		if err != nil {
			log.Fatalf("Error writing to csv file: %v.", err)
		}

		// Retrieve balance for ERC20 tokens per specified chain and timestamp
		tokenResponse := getTokenBalance(value.Address, value.Chain, strconv.Itoa(block.Block), "")

		// Parse the response for ERC20 token to an accessible ERC20 token data strucuture
		tokenData := ResponseToToken(tokenResponse)

		// Range over and access the data structure from "tokenData" variable and store in "token" variable
		for _, token := range tokenData {
			var erc20TokenChecker string

			// Convert ERC20 token balance from string to number
			tokenString, err := strconv.ParseFloat(token.Balance, 64)
			if err != nil {
				log.Fatalf("Error parsing token balances: %v", err)
			}

			// Convert ERC20 token balance from specified decimals to no decimals
			tokenBalance := tokenString / math.Pow10(token.Decimals)

			// Retrieve price for ERC20 token
			erc20Price := price.GetPrice(token.TokenAdress, value.Chain, block.Block)

			// Calculate USD value
			Uvalue = erc20Price * tokenBalance

			// Logic for printing ERC20 token balance checker tool to output CSV
			if value.Chain == Ethereum {
				erc20TokenChecker = EthereumTokenChecker
			} else if value.Chain == Polygon {
				erc20TokenChecker = PolygonTokenChecker
			} else if value.Chain == Arbitrum {
				erc20TokenChecker = ArbitrumTokenChecker
			} else if value.Chain == BinanceSmartChain {
				erc20TokenChecker = BSCTokenChecker
			} else if value.Chain == Fantom {
				erc20TokenChecker = FantomTokenChecker
			} else if value.Chain == Cronos {
				erc20TokenChecker = CronosTokenChecker
			} else if value.Chain == Avalanche {
				erc20TokenChecker = AvalancheTokenChecker
			}

			// Store and write values for ERC20 token data
			tokenRecord := []string{value.Address, value.Chain, token.Name, token.Symbol, token.TokenAdress, fmt.Sprintf("%f", tokenBalance), strconv.Itoa(blockNo), erc20TokenChecker, strconv.FormatFloat(erc20Price, 'f', 6, 64), strconv.FormatFloat(Uvalue, 'f', 6, 64)}
			err = writer.Write(tokenRecord)
			if err != nil {
				log.Fatalf("Error: %v.", err)
			}

		}
	}
}

// Used to get one ERC20 token balance for an address for specified chain.
func getTokenBalance(address string, chain string, block string, tokens string) []byte {
	resp, err := getRequest(ConstructTokenURL(address, chain, block, tokens))
	if err != nil {
		log.Fatalf("Error performing request to get balance from Moralis API: %v.", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing response body to get token balance from Moralis API: %v.", err)

	}

	return body
}

// Used to getBalance from an address on specified blockchain
func getBalance(address string, chain string, block string) []byte {
	resp, err := getRequest(ConstructNativeURL(address, chain, block))
	if err != nil {
		log.Fatalf("Error performing request to get balance from Moralis API: %v.", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing response body for getBalance from Moralis API: %v.", err)
	}

	return body
}

// Helper function to perform HTTP GET requests
func getRequest(url string) (*http.Response, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Error creating HTTP GET request: %v.", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", initialisers.APIKEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatalf("Error creating HTTP GET request: %v.", err)
	}

	return resp, nil
}

// Used to create the URL to query the balance of an address for the native chain token.
func ConstructNativeURL(address string, chain string, block string) string {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%v/balance?chain=%v&to_block=%v", address, chain, block)

	return url
}

// Used to create the URL to query the balance of an address for ERC20 tokens.
func ConstructTokenURL(address string, chain string, block string, tokens string) string {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%v/erc20?chain=%v&to_block=%v&token_adresses%%5B0%%5D=%v", address, chain, block, tokens)
	// fmt.Println(url)
	return url
}

// Used to parse token balance response data to readable data to use in the programme.
func ResponseToToken(body []byte) []TokenBalance {
	var response []TokenBalance

	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error parsing token balance response data: %v.", err)
	}

	return response
}

// Used to parse native token balance response data to readable format to use in the programme.
func ResponseToNative(body []byte) NativeBalance {
	var response NativeBalance

	err := json.Unmarshal(body, &response)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON for native balance: %v.", err)
	}

	return response
}

// Parses the data in the input CSV file to an accessbile variable by the script.
func CsvToToken(filename string) []TokenFile {
	file, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening the file %v: %v.", filename, err)
	}

	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Error readint the file: %v.", err)
	}

	var data []TokenFile
	var jData TokenFile

	for _, record := range records {
		jData.Address = record[0]
		jData.Chain = record[1]
		data = append(data, jData)
	}

	return data
}

// Perform GET request to Moralis API in order to retrieve the block number at specifed timestamp.
func GetBlock(chain string, timestamp string) Block {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/dateToBlock?chain=%v&date=%v", chain, timestamp)

	resp, err := getRequest(url)
	if err != nil {
		log.Fatalf("Error performing GET request to get the block number: %v.", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error parsing block response body: %v.", err)
	}

	var block Block

	err = json.Unmarshal(body, &block)
	if err != nil {
		log.Fatalf("Error parsing block JSON response: %v.", err)
	}

	return block
}

// Used to return the specified timestamp in unix format for use with the Moralis API.
func Timestamp(date string) string {
	utc, err := time.Parse("02/01/2006 15:04:05 UTC", date)
	if err != nil {
		log.Fatalf("Error parsing UTC timestamp: %v.", err)
	}

	unixTime := utc.Unix()

	return strconv.Itoa(int(unixTime))
}
