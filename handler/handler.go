package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"githup.com/GoLangTask/model"
)

type TransferRequest struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}

type API struct {
	Accounts *model.Accounts
}

func downloadFile(url, filepath string) error {
	// Send GET request to download the file
	response, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}
	defer response.Body.Close()

	// Create the file
	file, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	// Write the downloaded file to the local file
	_, err = io.Copy(file, response.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	return nil
}

func parseAccounts(filepath string) ([]model.Account, error) {
	// Read the JSON data from the file
	data, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read JSON file: %v", err)
	}

	// Parse the JSON data into account objects
	var accounts []model.Account
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON data: %v", err)
	}

	return accounts, nil
}
func ShowData() {
	fileURL := "https://git.io/Jm76h"
	filePath := "../accounts.json"

	err := downloadFile(fileURL, filePath)
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
		return
	}

	fmt.Printf("File downloaded successfully: %s\n", filePath)

	accounts, err := parseAccounts(filePath)
	if err != nil {
		fmt.Printf("Error parsing accounts: %v\n", err)
		return
	}

	fmt.Println("Accounts parsed successfully:")
	for _, account := range accounts {
		fmt.Printf("ID: %s, Name: %s, Balance: %s\n", account.ID, account.Name, account.Balance)
	}
}

// ListAccountsHandler handles the request to list all accounts.
func (api *API) ListAccountsHandler(w http.ResponseWriter, r *http.Request) {
	api.Accounts.RLock()
	defer api.Accounts.RUnlock()

	accountsData, err := json.Marshal(api.Accounts.Data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(accountsData)
}

// TransferHandler handles the request to transfer money between accounts.
func (api *API) TransferHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	var transferReq TransferRequest
	err = json.Unmarshal(body, &transferReq)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Error: %s", err.Error())
		return
	}

	api.Accounts.Lock()
	defer api.Accounts.Unlock()

	fromAccount, toAccount := api.findAccounts(transferReq.From, transferReq.To)
	if fromAccount == nil || toAccount == nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Invalid account(s) specified")
		return
	}

	amount := convertStringToFloat(transferReq.Amount)
	if amount <= 0 || amount > convertStringToFloat(fromAccount.Balance) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, "Error: Invalid transfer amount")
		return
	}

	fromAccount.Balance = convertFloatToString(convertStringToFloat(fromAccount.Balance) - amount)
	toAccount.Balance = convertFloatToString(convertStringToFloat(toAccount.Balance) + amount)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Transfer successful")
}

func (api *API) findAccounts(from, to string) (*model.Account, *model.Account) {
	for i := range api.Accounts.Data {
		if api.Accounts.Data[i].ID == from {
			for j := range api.Accounts.Data {
				if api.Accounts.Data[j].ID == to {
					return &api.Accounts.Data[i], &api.Accounts.Data[j]
				}
			}
		}
	}
	return nil, nil
}

func convertStringToFloat(s string) float64 {
	var f float64
	_, err := fmt.Sscanf(s, "%f", &f)
	if err != nil {
		log.Printf("Error converting string to float: %s", err.Error())
	}
	return f
}

func convertFloatToString(f float64) string {
	return fmt.Sprintf("%.2f", f)
}

func RunCode() {
	// Step 1: Parse the JSON file and ingest the accounts
	data, err := ioutil.ReadFile("../accounts.json")
	if err != nil {
		log.Fatal("Error reading accounts.json file:", err)
	}

	var accounts []model.Account
	err = json.Unmarshal(data, &accounts)
	if err != nil {
		log.Fatal("Error parsing accounts data:", err)
	}

	accountsData := &model.Accounts{
		Data: accounts,
	}

	// Step 2: Implement the RESTful API endpoints
	api := &API{
		Accounts: accountsData,
	}

	http.HandleFunc("/accounts", api.ListAccountsHandler)
	http.HandleFunc("/transfer", api.TransferHandler)

	// Step 3: Print console output message when all accounts have been ingested
	fmt.Println("Accounts ingestion completed. Ready to make transfers.")

	// Step 4: Run the HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
