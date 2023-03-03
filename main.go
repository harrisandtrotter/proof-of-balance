package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
)

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

func main() {
	for {
		fmt.Println("\nWelcome to Proof-of-Balance. \nSee below for the available options. ")
		fmt.Println("1. Block number")
		fmt.Println("2. Retrieve balances")

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("\nSelect the option you want to use today (e.g., 1 for receiving specific block number): \n")
		input, _ := reader.ReadString('\n')

		switch strings.TrimSpace(input) {
		case "1":
			fmt.Println("Coming soon.")
		case "2":
			fmt.Println("Please select the CSV file. \nThe format needs to be the wallet address in the column A and the relevant chain in column B. \n\nThe available chains and the required format are as follows: \neth\narbitrum\nfantom\nbsc\npolygon\navalanche\ncronos")
			time.Sleep(time.Second * 1)
			GetTokenBalance()
		default:
			fmt.Println("Invalid option selected.")
		}

		time.Sleep(time.Second * 3)
	}

}

// Main get balance function to be ran in the main script.
func GetTokenBalance() {
	// Get user to select CSV input file
	filename, err := dialog.File().Filter("CSV file", "csv").Load()
	if err != nil {
		fmt.Println("Error finding file:", &err)
	}

	data := CsvToToken(filename)

	// Create CSV output file
	output, err := os.Create("retrieved-data.csv")
	if err != nil {
		log.Fatalf("Error creating csv file: %v.", err)
	}

	defer output.Close()

	writer := csv.NewWriter(output)
	defer writer.Flush()

	// CSV output file headers
	headers := []string{"Address", "Chain", "Token Name", "Token Symbol", "Token Address", "Balance", "Native Balance", "Block number", "Token checker"}
	// Write the headers to the CSV
	err = writer.Write(headers)
	if err != nil {
		log.Fatalf("Error writing headers to CSV file: %v.", err)
	}

	// Range over parsed data for tokens
	for _, value := range data {
		// Retrieve block number for specified time
		block := GetBlock(value.Chain, Timestamp("31/12/2022 23:59:59 UTC"))

		// Perform token balance request and store in variable "response"
		response := getTokenBalance(value.Address, value.Chain, strconv.Itoa(block.Block), "")
		_ = response

		// Perform native token balance request and store in variable "nativeResponse"
		nativeResponse := getBalance(value.Address, value.Chain, strconv.Itoa(block.Block))
		_ = nativeResponse

		// Parse native token response data to accessible format for the script
		nativeData := ResponseToNative(nativeResponse)

		// Parse token response data to accessible format for the script
		newData := ResponseToToken(response)

		for _, tokens := range newData {

			// Convert token balance data from string to number
			balanceString, err := strconv.ParseFloat(tokens.Balance, 64)
			if err != nil {
				log.Fatalf("Error reading decimals amount: %v.", err)
			}

			// Convert native token balance data from string to number
			nativeBalanceStr, err := strconv.ParseFloat(nativeData.Balance, 64)
			if err != nil {
				log.Fatalf("Error reading decimals amount for native token: %v.", err)
			}

			// Deal with decimals conversion to readable number for balances
			nativeBalanceNumber := nativeBalanceStr / math.Pow10(18)
			balance := balanceString / math.Pow10(tokens.Decimals)

			// Write the retrieved data to the CSV output file
			record := []string{value.Address, value.Chain, tokens.Name, tokens.Symbol, tokens.TokenAdress, fmt.Sprintf("%f", balance), fmt.Sprintf("%f", nativeBalanceNumber), strconv.Itoa(block.Block), ""}
			err = writer.Write(record)
			if err != nil {
				log.Fatalf("Error writing tokens to CSV file: %v.", err)
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
	req.Header.Add("X-API-Key", "ADD YOUR MORALIS API KEY HERE")

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
