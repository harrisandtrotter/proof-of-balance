package blocks

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/harrisandtrotter/proof-of-balance/server/initialisers"
)

type Block struct {
	Date           string `json:"date"`
	Block          int    `json:"block"`
	Timestamp      int    `json:"timestamp"`
	BlockTimestamp string `json:"block_timestamp"`
	Hash           string `json:"hash"`
	ParentHash     string `json:"parent_hash"`
}

// Accesses the Block struct and returns the block number.
func (b *Block) BlockNumber(chain, timestamp string) int {
	block := b.RetrieveBlock(chain, timestamp)

	return block.Block
}

// Queries Moralis API to return Block struct. Takes "chain" and "timestamp" variable.
func (b *Block) RetrieveBlock(chain string, timestamp string) Block {
	var blockchain string
	var url string

	if chain == "eth" || chain == "ethereum" || chain == "Ethereum" || chain == "ETH" || chain == "Eth" {
		blockchain = "eth"
	} else if chain == "polygon" || chain == "matic" || chain == "Polygon" || chain == "MATIC" {
		blockchain = "polygon"
	} else if chain == "arbitrum" || chain == "Arbitrum" || chain == "arb" {
		blockchain = "arbitrum"
	} else if chain == "bsc" || chain == "binance" || chain == "binance smart chain" || chain == "bnb chain" || chain == "bnb" || chain == "BNB" || chain == "Binance Smart Chain" || chain == "BSC" {
		blockchain = "bsc"
	} else if chain == "ftm" || chain == "fantom" || chain == "FTM" || chain == "Fantom" {
		blockchain = "fantom"
	} else if chain == "cro" || chain == "CRO" || chain == "cronos" || chain == "Cronos" {
		blockchain = "cronos"
	} else if chain == "avax" || chain == "avalanche" || chain == "AVAX" {
		blockchain = "avalanche"
	} else {
		fmt.Println("\nBlockchain not supported. Please use on the supported chains:\nEthereum\nArbitrum\nPolygon\nBinance Smart Chain\nFantom\nAvalanche")
		time.Sleep(time.Second * 3)
	}

	url = fmt.Sprintf("https://deep-index.moralis.io/api/v2/dateToBlock?chain=%v&date=%v", blockchain, b.TimestampToUnix(timestamp))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Error creating GET request to retrieve block: %v.", err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", initialisers.APIKEY)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("Error performing GET request to retrieve block: %v.", err)

	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error parsing retrieve block response body: %v.", err)
	}

	var block Block

	err = json.Unmarshal(body, &block)
	if err != nil {
		fmt.Printf("Error unmarshalling block body: %v", err)
	}

	return block
}

// Used to convert "timestamp" variable to unix format. Takes "timestamp" in "31/12/2022 23:00:00" format.
func (b *Block) TimestampToUnix(timestamp string) string {
	utc, err := time.Parse("2006-01-02 15:04:05 UTC", timestamp+" UTC")
	if err != nil {
		fmt.Printf("Error parsing UTC timestamp: %v", err)
	}

	unix := utc.Unix()

	return strconv.Itoa(int(unix))
}
