package initialisers

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

var APIKEY string

// Efficiently load environment variables
func LoadEnvironment() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("error loading environment variables. message: %v", err)
	}

}

func LoadAPIKey() {
	APIKEY = os.Getenv("MORALIS_API_KEY")
}
