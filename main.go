package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kollalabs/sdk-go/kc"
)

type WrappedJSON struct {
	Data interface{} `json:"data"`
}

type People struct {
	Employees []Employee `json:"data"`
}

type Employee struct {
	EmployeeID string `json:"employeeId"`
	Status     Status `json:"status"`
	Name       string `json:"name"`
	Start      string `json:"start"`
	End        string `json:"end"`
	Created    string `json:"created"`
	Amount     Amount `json:"amount"`
	Notes      Notes  `json:"notes"`
}

type Status struct {
	LastChanged         string `json:"lastChanged"`
	LastChangedByUserId string `json:"lastChangedByUserId"`
	Status              string `json:"status"`
}

type Amount struct {
	Unit   string `json:"unit"`
	Amount string `json:"amount"`
}

type Notes struct {
	Manager  string `json:"manager"`
	Employee string `json:"employee"`
}

func main() {
	oldItems := getItems()
	if oldItems == nil {
		fmt.Println("Error getting old items.")
		return
	}

	apiKey, mondayConnector, kollaCustomerID, boardID, groupID, bambooConnector, bambooCustomerID, companyDomain := getVars()
	if apiKey == "" {
		fmt.Println("Error getting env vars.")
		return
	}

	kolla, err := kc.New(apiKey)
	if err != nil {
		fmt.Println("Error connecting to Kolla.")
		panic(err)
	}

	creds := getCreds(kolla, mondayConnector, kollaCustomerID)
	if creds == nil {
		fmt.Println("Error getting credentials.")
		return
	}
	mondayKey := creds.Token

	url := "https://api.monday.com/v2"

	success := deleteItems(oldItems, url, mondayKey)
	if !success {
		fmt.Println("Error deleting old items.")
		return
	}

	query := "query { users { account { id show_timeline_weekends tier slug plan { period }}}} "
	payloadBytes := getPayload(query)
	if payloadBytes == nil {
		fmt.Println("Error getting payload.")
		return
	}

	req := getPostRequest(url, payloadBytes)
	if req == nil {
		fmt.Println("Error getting post request.")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mondayKey)

	resp := doRequest(req)
	if resp == nil {
		fmt.Println("Error getting response.")
		return
	}
	defer resp.Body.Close()

	responseJSON := getResponse(resp)
	if responseJSON == nil {
		fmt.Println("Error getting response JSON.")
		return
	}
	success = turnPretty(responseJSON)
	if !success {
		fmt.Println("Error in turning JSON pretty.")
		return
	}

	query = "query { boards (ids: " + boardID + ") { name state id groups { title id } columns { type } }}"
	payloadBytes = getPayload(query)
	if payloadBytes == nil {
		fmt.Println("Error getting payload.")
		return
	}

	req = getPostRequest(url, payloadBytes)
	if req == nil {
		fmt.Println("Error getting post request.")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mondayKey)

	resp = doRequest(req)
	if resp == nil {
		fmt.Println("Error getting response.")
		return
	}
	defer resp.Body.Close()

	responseJSON = getResponse(resp)
	if responseJSON == nil {
		fmt.Println("Error getting response JSON.")
		return
	}
	success = turnPretty(responseJSON)
	if !success {
		fmt.Println("Error in turning JSON pretty.")
		return
	}

	var items []string
	/*for i := 0; i < 3; i++ {
		name := "test"
		id := "1"
		start := "2023-09-09"
		end := "2023-09-09"
		status := "approved"
		created := "2023-09-09"
		items = createTestEmployee(name, id, start, end, status, created, boardID, groupID, url, mondayKey, items)
	}*/

	items = bamboo(kolla, bambooConnector, bambooCustomerID, companyDomain, boardID, groupID, mondayKey, url, items)
	if items == nil {
		fmt.Println("Error getting items.")
		return
	}

	success = addItems(items)
	if !success {
		fmt.Println("Error adding items.")
		return
	}
}

func getVars() (string, string, string, string, string, string, string, string) {
	var apiKey, mondayConnector, kollaCustomerID, boardID, groupID, bambooConnector, bambooCustomerID, companyDomain string

	file, err := os.Open("env-vars.txt")
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return "", "", "", "", "", "", "", ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "=")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if key == "APIKEY" {
				apiKey = value
			} else if key == "MONDAYCONNECTOR" {
				mondayConnector = value
			} else if key == "KOLLACUSTOMERID" {
				kollaCustomerID = value
			} else if key == "BOARDID" {
				boardID = value
			} else if key == "GROUPID" {
				groupID = value
			} else if key == "BAMBOOCONNECTOR" {
				bambooConnector = value
				fmt.Println(key, bambooConnector)
			} else if key == "BAMBOOCUSTOMERID" {
				bambooCustomerID = value
				fmt.Println(key, bambooCustomerID)
			} else if key == "COMPANYDOMAIN" {
				companyDomain = value
				fmt.Println(key, companyDomain)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the file:", err)
		return "", "", "", "", "", "", "", ""
	}

	return apiKey, mondayConnector, kollaCustomerID, boardID, groupID, bambooConnector, bambooCustomerID, companyDomain
}

func getCreds(kolla *kc.Client, connector string, customerID string) *kc.Credentials {
	ctx := context.Background()
	creds, err := kolla.Credentials(ctx, connector, customerID)
	if err != nil {
		fmt.Println("Error getting credentials.")
		return nil
	}
	return creds
}

func getPayload(query string) []byte {
	payload := map[string]interface{}{
		"query": query,
	}
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return nil
	}
	return payloadBytes
}

func getPostRequest(url string, payloadBytes []byte) *http.Request {
	req, err := http.NewRequest("POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return req
}

func doRequest(req *http.Request) *http.Response {
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil
	}
	return resp
}

func getResponse(resp *http.Response) map[string]interface{} {
	var responseJSON map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&responseJSON)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}
	return responseJSON
}

func turnPretty(responseJSON map[string]interface{}) bool {
	prettyJSON, err := json.MarshalIndent(responseJSON, "", "  ")
	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return false
	}
	fmt.Println(string(prettyJSON))
	return true
}

/*
func createTestEmployee(name string, id string, start string, end string, status string, created string, boardID string, groupID string, url string, mondayKey string, items []string) []string {
	fmt.Println(name, "\t", id, "\t", status, "\t", start, "\t", end, "\t", created)

	column_values := `"{\"text\":\"` + id + `\",
								\"status\":\"` + status + `\",
								\"date4\":\"` + start + `\",
								\"date\":\"` + end + `\",
								\"created1\":\"` + created + `\"}"`

	query := `mutation {
				create_item
					(
						board_id: ` + boardID + `,
						group_id: "` + groupID + `",
						item_name: "` + name + `",
						column_values: ` + column_values + `
					)
					{
						id
					}
				}`
	payloadBytes := getPayload(query)

	req := getPostRequest(url, payloadBytes)

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", mondayKey)

	resp := doRequest(req)

	responseJSON := getResponse(resp)
	itemID := responseJSON["data"].(map[string]interface{})["create_item"].(map[string]interface{})["id"].(string)
	items = append(items, itemID)

	turnPretty(responseJSON)
	return items
}
*/

func addItems(items []string) bool {
	filePath := "item-ids.txt"

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return false
	}
	defer file.Close()

	for _, itemID := range items {
		_, err = file.WriteString(itemID + "\n")
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return false
		}
	}
	return true
}

func getItems() []string {
	filePath := "item-ids.txt"
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return nil
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading file:", err)
		return nil
	}

	return lines
}

func deleteItems(oldItems []string, url string, mondayKey string) bool {
	for _, item := range oldItems {
		query := "mutation { delete_item (item_id: " + item + ") { id }}"
		payloadBytes := getPayload(query)
		if payloadBytes == nil {
			fmt.Println("Error getting payload.")
			return false
		}

		req := getPostRequest(url, payloadBytes)
		if req == nil {
			fmt.Println("Error getting post request.")
			return false
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", mondayKey)

		resp := doRequest(req)
		if resp == nil {
			fmt.Println("Error getting response.")
			return false
		}
		defer resp.Body.Close()

		responseJSON := getResponse(resp)
		if responseJSON == nil {
			fmt.Println("Error getting JSON.")
			return false
		}
		success := turnPretty(responseJSON)
		if !success {
			fmt.Println("Error turning JSON pretty.")
			return false
		}
	}
	return true
}

func bamboo(kolla *kc.Client, bambooConnector string, customerID string, companyDomain string, boardID string, groupID string, mondayKey string, mondayURL string, items []string) []string {
	creds := getCreds(kolla, bambooConnector, customerID)

	bambooKey := creds.Token
	today := time.Now()
	oneMonthFromToday := today.AddDate(0, 1, 0)
	start := today.Format("2006-01-02")
	end := oneMonthFromToday.Format("2006-01-02")

	url := "https://" + bambooKey + ":x@api.bamboohr.com/api/gateway.php/" + companyDomain + "/v1/time_off/requests/?start=" + start + "&end=" + end
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil
	}

	req.Header.Add("Accept", "application/json")

	resp := doRequest(req)
	if resp == nil {
		fmt.Println("Error getting response.")
		return nil
	}
	defer resp.Body.Close()

	responseJSON, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return nil
	}
	wrappedJSON := WrappedJSON{Data: json.RawMessage(responseJSON)}

	prettyJSON, err := json.MarshalIndent(wrappedJSON, "", "    ")
	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return nil
	}

	var resObj People
	json.Unmarshal(prettyJSON, &resObj)

	for _, employee := range resObj.Employees {
		name := employee.Name
		id := employee.EmployeeID
		start := employee.Start
		end := employee.End
		status := employee.Status.Status
		created := employee.Created
		amount := employee.Amount.Amount
		unit := employee.Amount.Unit
		combinedAmount := amount + " " + unit
		var employeeNotes, managerNotes string
		if employee.Notes.Employee != "" && employee.Notes.Manager != "" {
			employeeNotes = employee.Notes.Employee
			managerNotes = employee.Notes.Manager
		} else if employee.Notes.Employee != "" && employee.Notes.Manager == "" {
			employeeNotes = employee.Notes.Employee
			managerNotes = ""
		} else if employee.Notes.Employee == "" && employee.Notes.Manager != "" {
			managerNotes = employee.Notes.Manager
			employeeNotes = ""
		}
		fmt.Println(name, "\t", id, "\t", status, "\t", start, "\t", end, "\t", created, "\t", combinedAmount, "\t", employeeNotes, "\t", managerNotes)

		column_values := `"{\"text\":\"` + id + `\",
							\"status\":\"` + status + `\",
							\"date4\":\"` + start + `\",
							\"date\":\"` + end + `\",
							\"created1\":\"` + created + `\",
							\"text3\":\"` + combinedAmount + `\",
							\"text2\":\"` + employeeNotes + `\",
							\"text38\":\"` + managerNotes + `\"}"`

		query := `mutation {
					create_item
						(
							board_id: ` + boardID + `,
							group_id: "` + groupID + `",
							item_name: "` + name + `",
							column_values: ` + column_values + `
						)
						{
							id
						}
					}`

		payloadBytes := getPayload(query)
		if payloadBytes == nil {
			fmt.Println("Error getting payload.")
			return nil
		}

		req := getPostRequest(mondayURL, payloadBytes)
		if req == nil {
			fmt.Println("Error getting request.")
			return nil
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", mondayKey)

		resp := doRequest(req)
		if resp == nil {
			fmt.Println("Error getting response.")
			return nil
		}
		defer resp.Body.Close()

		responseJSON := getResponse(resp)
		if responseJSON == nil {
			fmt.Println("Error getting JSON.")
			return nil
		}
		itemID := responseJSON["data"].(map[string]interface{})["create_item"].(map[string]interface{})["id"].(string)
		items = append(items, itemID)
		if items == nil {
			fmt.Println("Error appending items.")
			return nil
		}

		success := turnPretty(responseJSON)
		if !success {
			fmt.Println("Error turning JSON pretty.")
			return nil
		}
	}
	return items
}
