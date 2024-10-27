package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
)

// Define a struct for transactions
type Transaction struct {
	Hash  string `json:"hash"`
	From  string `json:"from"`
	To    string `json:"to"`
	Value string `json:"value"`
	Time  string `json:"timeStamp"`
}

var ethClient *ethclient.Client

func init() {
	// Initialize Ethereum Client
	var err error
	ethClient, err = ethclient.Dial("https://mainnet.infura.io/v3/48b725421b1b4693975c19019dcaa8fd")
	if err != nil {
		panic(err)
	}
}

// RenderHTML renders the HTML template
func RenderHTML(c *gin.Context, data interface{}) {
	tmpl, err := template.ParseFiles("index.html")
	if err != nil {
		c.String(http.StatusInternalServerError, "Failed to load template: %v", err)
		return
	}
	c.Header("Content-Type", "text/html")
	tmpl.Execute(c.Writer, data)
}

// GetBalance retrieves the Ethereum balance for a given address
func GetBalance(address string) string {
	addr := common.HexToAddress(address)
	balance, err := ethClient.BalanceAt(context.Background(), addr, nil)
	if err != nil {
		return "Error fetching balance"
	}
	return balance.String() // Returns balance in wei
}

// GetTransactions retrieves the latest transactions for a given address from Etherscan
func GetTransactions(address string) ([]Transaction, error) {
	apiKey := "EZEYWNY8TZFDT4XTNVMRJFU857SXVYAY1P" // Replace with your Etherscan API Key
	url := fmt.Sprintf("https://api.etherscan.io/api?module=account&action=txlist&address=%s&startblock=0&endblock=99999999&sort=desc&apikey=%s", address, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Status  string        `json:"status"`
		Message string        `json:"message"`
		Data    []Transaction `json:"result"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "1" {
		return nil, fmt.Errorf("etherscan API error: %s", result.Message)
	}

	return result.Data, nil
}

// DashboardHandler handles the dashboard route
func DashboardHandler(c *gin.Context) {
	address := c.Query("address")
	balance := ""
	transactions := []Transaction{}

	if address != "" {
		balance = GetBalance(address)
		trans, err := GetTransactions(address)
		if err == nil {
			transactions = trans
		}
	}

	RenderHTML(c, gin.H{"address": address, "balance": balance, "transactions": transactions})
}

func main() {
	router := gin.Default()

	// Serve the dashboard
	router.GET("/", DashboardHandler)

	router.Run(":8080") // Start the server
}
