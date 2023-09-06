package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/harrisandtrotter/proof-of-balance/blocks"
	"github.com/harrisandtrotter/proof-of-balance/initialisers"
	"github.com/harrisandtrotter/proof-of-balance/models"
)

var block blocks.Block

func Setup() {
	router := fiber.New()

	router.Post("/balances", GetBalance)

	log.Fatal(router.Listen(":8000"))
}

func GetBalance(c *fiber.Ctx) error {
	c.Accepts("application/json")

	var body map[string]string

	// parse request body into request struct
	if err := c.BodyParser(&body); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "error with request. Please contact support (devops@harrisandtrotter.co.uk)",
		})
	}

	// assign request body values to request variable
	request := models.Request{
		Address:   body["address"],
		Chain:     body["chain"],
		Date:      body["date"],
		Timestamp: body["timestamp"],
	}

	// chain for moralis API
	chain, err := models.DetermineChain(request.Chain)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// block number based on chain and timestamp
	blockNo := block.BlockNumber(chain, request.Date+" "+request.Timestamp)

	// relevant info to be returned to user
	asset, url, name, err := models.ReturnNativeInfo(chain)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// get native balance
	nativeBalanceResp, err := getNativeBalance(request.Address, chain, blockNo)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	// convert type string to float64
	balanceStr, err := strconv.ParseFloat(nativeBalanceResp.Balance, 64)

	// convert from wei to ether
	balance := balanceStr / math.Pow10(18)

	var tokenBalanceResponse *models.TokenBalance

	tokenBalanceResp, err := getTokenBalance(request.Address, chain, blockNo)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	for _, value := range tokenBalanceResp {
		tokenBalanceResponse = &value

		tokenStr, err := strconv.ParseFloat(tokenBalanceResponse.Balance, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		tokenBalance := tokenStr / math.Pow10(tokenBalanceResponse.Decimals)

		tokenBalanceStr := strconv.FormatFloat(tokenBalance, 'f', 6, 64)

		tokenBalanceResponse.Balance = tokenBalanceStr
	}

	return c.JSON(fiber.Map{
		"addresss":                     request.Address,
		"chain":                        chain,
		"block_number":                 blockNo,
		"asset":                        asset,
		"native_checker_url":           url,
		"token_name":                   name,
		"balance":                      balance,
		"erc20_token_name":             tokenBalanceResponse.Name,
		"erc20_token_symbol":           tokenBalanceResponse.Symbol,
		"erc20_token_contract_address": tokenBalanceResponse.TokenAddress,
		"erc20_token_balance":          tokenBalanceResponse.Balance,
	})

}

// Get ERC20 token balance
func getTokenBalance(address, chain string, block int) ([]models.TokenBalance, error) {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%v/erc20?chain=%v&to_block=%v", address, chain, block)

	resp, err := get(url)
	if err != nil {
		return []models.TokenBalance{}, err
	}

	var response *[]models.TokenBalance

	err = json.Unmarshal(resp, &response)
	if err != nil {
		return []models.TokenBalance{}, err
	}

	return *response, nil
}

// Get native token balance
func getNativeBalance(address, chain string, block int) (models.NativeBalance, error) {
	url := fmt.Sprintf("https://deep-index.moralis.io/api/v2/%v/balance?chain=%v&to_block=%v", address, chain, block)

	resp, err := get(url)
	if err != nil {
		return models.NativeBalance{}, err
	}

	var response *models.NativeBalance

	err = json.Unmarshal(resp, &response)
	if err != nil {
		return models.NativeBalance{}, err
	}

	return *response, nil
}

func get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("X-API-Key", initialisers.APIKEY)
	req.Header.Add("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
