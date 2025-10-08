package reports

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const baseUrl = "https://botw-compendium.herokuapp.com/api/v3/compendium"

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type LozClient struct {
	baseUrl    string
	httpClient HttpClient
}

func NewClient(httpClient HttpClient) *LozClient {
	return &LozClient{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}

type Monster struct {
	Name            string   `json:"name"`
	Id              int      `json:"id"`
	Category        string   `json:"category"`
	Description     string   `json:"description"`
	Image           string   `json:"image"`
	CommonLocations []string `json:"common_locations"`
	Drops           []string `json:"drops"`
	Dlc             bool     `json:"dlc"`
}

type GetMonstersResponse struct {
	Data []Monster `json:"data"`
}

func (c *LozClient) GetMonsters() (*GetMonstersResponse, error) {
	req, err := http.NewRequest("GET", c.baseUrl+"/category/monsters", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	reqUrl := req.URL
	queryParams := req.URL.Query()
	queryParams.Set("game", "totk")
	reqUrl.RawQuery = queryParams.Encode()

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	var response *GetMonstersResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return response, nil
}
