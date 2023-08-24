package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/kollalabs/sdk-go/kc"
)

type WrappedJSON struct {
	Data interface{} `json:"data"`
}

type People struct {
	Employees []Employee `json:"data"`
}

type Employee struct {
	EmployeeID string    `json:"employeeId"`
	Status     Status    `json:"status"`
	Name       string    `json:"name"`
	Start      string    `json:"start"`
	End        string    `json:"end"`
	Created    string    `json:"created"`
	Type       []Types   `json:"type"`
	Amount     []Amounts `json:"amount"`
}

type Status struct {
	LastChanged         string `json:"lastChanged"`
	LastChangedByUserId string `json:"lastChangedByUserId"`
	Status              string `json:"status"`
}

type Types struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type Amounts struct {
	Unit   string `json:"unit"`
	Amount string `json:"amount"`
}

func main() {
	apiKey, mondayConnector, customerID, boardID, groupID, bambooConnector, companyDomain := getVars()
	fmt.Println(groupID, bambooConnector, companyDomain)

	kolla, err := kc.New(apiKey)
	if err != nil {
		fmt.Println("Error connecting to Kolla.")
		panic(err)
	}

	creds := getCreds(kolla, mondayConnector, customerID)
	mondayKey := creds.Token

	url := "https://api.monday.com/v2"

	// Getting monday.com account details.
	query := "query { users { account { id show_timeline_weekends tier slug plan { period }}}} "
	payloadBytes := getPayload(query)

	req := getPostRequest(url, payloadBytes)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mondayKey)

	resp := doRequest(req)
	defer resp.Body.Close()

	responseJSON := getResponse(resp)
	turnPretty(responseJSON)

	// Getting monday.com board.
	query = "query { boards (ids: " + boardID + ") { name state id groups { title id } columns { type } }}"
	payloadBytes = getPayload(query)

	req = getPostRequest(url, payloadBytes)

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mondayKey)

	resp = doRequest(req)

	responseJSON = getResponse(resp)
	turnPretty(responseJSON)

	createEmployee(boardID, groupID, url, mondayKey)

	query = "mutation { delete_item { id }}"

	payload := map[string]interface{}{
		"query": query,
	}

	payloadBytes, err = json.Marshal(payload)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req, err = http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", mondayKey)

	client := http.Client{}
	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	prettyResult, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(prettyResult))

	/*
		// Deleting group.
		query = `mutation { delete_group (board_id: ` + boardID + `, group_id: "` + groupID + `") { id deleted } }`

		data := map[string]interface{}{
			"query": query,
		}

		jsonData, err := json.Marshal(data)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		client = &http.Client{}
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", mondayKey)

		response, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer response.Body.Close()

		var responseData map[string]interface{}
		err = json.NewDecoder(response.Body).Decode(&responseData)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		prettyJSON, err = json.MarshalIndent(responseData, "", "  ")
		if err != nil {
			fmt.Println("Error formatting JSON:", err)
			return
		}

		fmt.Println(string(prettyJSON))
	*/

	/*
		// Connecting to bambooHR and getting time off requests.
		ctx = context.Background()
		creds, err = kolla.Credentials(ctx, bambooConnector, customerID)
		if err != nil {
			fmt.Println("Error getting credentials.")
			return
		}

		bambooKey := creds.Token
		today := time.Now()
		oneMonthFromToday := today.AddDate(0, 1, 0)
		start := today.Format("2006-01-02")
		end := oneMonthFromToday.Format("2006-01-02")

		url = "https://" + bambooKey + ":x@api.bamboohr.com/api/gateway.php/" + companyDomain + "/v1/time_off/requests/?start=" + start + "&end=" + end
		client = &http.Client{}
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Add("Accept", "application/json")

		response, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer response.Body.Close()

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		fmt.Println(&http.Client{})

		//wrappedJSON := WrappedJSON{Data: json.RawMessage(responseData)}

		prettyJSON, err = json.MarshalIndent(responseData, "", "    ")
		if err != nil {
			fmt.Println("Error formatting JSON:", err)
			return
		}

		fmt.Println(string(prettyJSON))

		// Write prettified JSON to a file
		err = ioutil.WriteFile("output.json", prettyJSON, 0644)
		if err != nil {
			fmt.Println("Error writing to file:", err)
			return
		}

		// Turning json into better object.
		var resObj People
		json.Unmarshal(prettyJSON, &resObj)

		for _, employee := range resObj.Employees {
			name := employee.Name
			id := employee.EmployeeID
			start := employee.Start
			end := employee.End
			status := employee.Status.Status
			created := employee.Created
			fmt.Println(name, "\t", id, "\t", status, "\t", start, "\t", end, "\t", created)

			column_values := `"{\"text\":\"` + id + `\",
								\"status\":\"` + status + `\",
								\"date4\":\"` + start + `\",
								\"date\":\"` + end + `\",
								\"created1\":\"` + created + `\"}"`

			// Updating an item.
			query = `mutation {
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

			url = "https://api.monday.com/v2"

			data := map[string]interface{}{
				"query": query,
			}

			jsonData2, err := json.Marshal(data)
			if err != nil {
				fmt.Println("Error marshaling JSON:", err)
				return
			}

			client = &http.Client{}
			req, err = http.NewRequest("POST", url, bytes.NewBuffer(jsonData2))
			if err != nil {
				fmt.Println("Error creating request:", err)
				return
			}

			req.Header.Add("Content-Type", "application/json")
			req.Header.Add("Authorization", mondayKey)

			response, err = client.Do(req)
			if err != nil {
				fmt.Println("Error sending request:", err)
				return
			}
			defer response.Body.Close()

			var responseData map[string]interface{}
			err = json.NewDecoder(response.Body).Decode(&responseData)
			if err != nil {
				fmt.Println("Error decoding JSON:", err)
				return
			}

				prettyJSON, err = json.MarshalIndent(responseData, "", "  ")
				if err != nil {
					fmt.Println("Error formatting JSON:", err)
					return
				}

				fmt.Println(string(prettyJSON))

		}
	*/

	/*
		// Deleting a board.
		query = "mutation { delete_board (board_id: 5024805591) { id }}"

		requestData := map[string]interface{}{
			"query": query,
		}

		requestJSON, err := json.Marshal(requestData)
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return
		}

		client = &http.Client{}
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(requestJSON))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Authorization", mondayKey)

		response, err := client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer response.Body.Close()

		responseData, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Println("Error reading response:", err)
			return
		}

		var formattedJSON bytes.Buffer
		err = json.Indent(&formattedJSON, responseData, "", "  ")
		if err != nil {
			fmt.Println("Error formatting JSON:", err)
			return
		}

		fmt.Println(string(formattedJSON.Bytes()))
	*/

	// Creating a board.
	/*
		query = "mutation { create_board (board_name: \"my board\", board_kind: public) { id }}"

		payload = map[string]interface{}{
			"query": query,
		}

		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			fmt.Println("Error marshaling JSON payload:", err)
			return
		}

		req, err = http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", mondayKey)

		client = &http.Client{}
		resp, err = client.Do(req)
		if err != nil {
			fmt.Println("Error sending request:", err)
			return
		}
		defer resp.Body.Close()

		var result2 map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&result2)
		if err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		prettyJSON, err = json.MarshalIndent(result2, "", "  ")
		if err != nil {
			fmt.Println("Error formatting JSON:", err)
			return
		}

		fmt.Println(string(prettyJSON))
	*/
}

func getVars() (string, string, string, string, string, string, string) {
	var apiKey, mondayConnector, customerID, boardID, groupID, bambooConnector, companyDomain string

	file, err := os.Open("env-vars.txt")
	if err != nil {
		fmt.Println("Error opening the file:", err)
		return "", "", "", "", "", "", ""
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
			} else if key == "CUSTOMERID" {
				customerID = value
			} else if key == "BOARDID" {
				boardID = value
			} else if key == "GROUPID" {
				groupID = value
			} else if key == "BAMBOOCONNECTOR" {
				bambooConnector = value
			} else if key == "COMPANYDOMAIN" {
				companyDomain = value
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading the file:", err)
		return "", "", "", "", "", "", ""
	}

	return apiKey, mondayConnector, customerID, boardID, groupID, bambooConnector, companyDomain
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

func turnPretty(responseJSON map[string]interface{}) {
	prettyJSON, err := json.MarshalIndent(responseJSON, "", "  ")
	if err != nil {
		fmt.Println("Error formatting JSON:", err)
		return
	}
	fmt.Println(string(prettyJSON))
}

func createEmployee(boardID string, groupID string, url string, mondayKey string) {
	name := "test"
	id := "1"
	start := "2023-09-09"
	end := "2023-09-09"
	status := "approved"
	created := "2023-09-09"
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
	fmt.Println(responseJSON)
	fmt.Println(responseJSON["account_id"])

	turnPretty(responseJSON)
}
