package price

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

const (
	MoralisAPI = "https://deep-index.moralis.io/api/v2"
)

type Price struct {
	NativePrice struct {
		Value    string `json:"value"`
		Decimals int    `json:"decimals"`
		Name     string `json:"name"`
		Symbol   string `json:"symbol"`
	} `json:"nativePrice"`
	UsdPrice        float64 `json:"usdPrice"`
	ExchangeAddress string  `json:"exchangeAddress"`
	ExchangeName    string  `json:"exchangeName"`
}

type Error struct {
	Message string `json:"message"`
}

type File struct {
	Name    string
	Address string
}

// The GetPrice function returns the price for the specified asset. It takes the chain id (as specified by the CG API), the token contract address,
// the quote currency required and the from and to timestamp range.
func (p *Price) GetPrice(address, chain string, block int) float64 {

	url := fmt.Sprintf("%v/erc20/%v/price?chain=%v&to_block=%v", MoralisAPI, address, chain, block)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}

	req.Header.Add("Accept", "application/json")
	req.Header.Add("X-API-Key", "JODRWjX4czsUtirHaEXxEY81jznW3gGQf1GExtDVP30Mao7HUWbvLSPrx2VNaaKE")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data Price
	var price float64
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	fmt.Println(url)

	msg := data.CheckError(body)
	if strings.Contains(msg, "No pools found with enough liquidity, to calculate the price") {
		fmt.Printf("\nCoin is not found. Coin is most likely spam!\nToken address: %v\n\n", address)
		price = 0
	} else {
		price = data.UsdPrice
		fmt.Printf("Exchange: %v\n", data.ExchangeName)
	}

	return price
}

func (p *Price) CheckError(body []byte) string {
	var errorMessage Error

	err := json.Unmarshal(body, &errorMessage)
	if err != nil {
		panic(err)
	}

	return errorMessage.Message
}

func (p *Price) ReturnPrice() []Price {
	var price []Price

	return price
}
