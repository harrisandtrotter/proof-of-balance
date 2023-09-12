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
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/harrisandtrotter/proof-of-balance/backend/blocks"
	"github.com/harrisandtrotter/proof-of-balance/backend/initialisers"
	"github.com/harrisandtrotter/proof-of-balance/backend/models"
)

var block blocks.Block

func Setup() {
	router := fiber.New()

	router.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE",
	}))

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
			"error determining chain": err.Error(),
		})
	}

	// block number based on chain and timestamp
	blockNo := block.BlockNumber(chain, request.Date+" "+request.Timestamp)

	// relevant info to be returned to user
	asset, url, name, tokenUrl, err := models.ReturnInfo(chain)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error with native information": err.Error(),
		})
	}

	// get native balance
	nativeBalanceResp, err := getNativeBalance(request.Address, chain, blockNo)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error with json": err.Error(),
		})
	}

	// convert type string to float64
	balanceStr, err := strconv.ParseFloat(nativeBalanceResp.Balance, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error with native type conversion": err.Error(),
		})
	}

	// convert from wei to ether
	balance := balanceStr / math.Pow10(18)

	tokenBalanceResp, err := getTokenBalance(request.Address, chain, blockNo)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error with erc20 token balances": err.Error(),
		})
	}
	var response []models.ClientResponse

	response = append(response, models.ClientResponse{
		Address:      request.Address,
		Chain:        chain,
		BlockNumber:  blockNo,
		Asset:        asset,
		AssetName:    name,
		AssetAddress: "N/A",
		Balance:      balance,
		CheckerUrl:   url,
		PossibleSpam: false,
	})

	for _, value := range tokenBalanceResp {

		tokenStr, err := strconv.ParseFloat(value.Balance, 64)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error with token type conversion": err.Error(),
			})
		}

		tokenBalance := tokenStr / math.Pow10(value.Decimals)

		response = append(response, models.ClientResponse{
			Address:      request.Address,
			Chain:        chain,
			BlockNumber:  blockNo,
			Asset:        value.Symbol,
			AssetName:    value.Name,
			AssetAddress: value.TokenAddress,
			Balance:      tokenBalance,
			CheckerUrl:   tokenUrl,
			PossibleSpam: value.PossibleSpam,
		})

	}

	return c.JSON(response)

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
