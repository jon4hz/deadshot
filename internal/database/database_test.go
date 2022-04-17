package database

import "testing"

func init() {
	if err := InitDB(); err != nil {
		panic(err)
	}
}

func TestGetAllNetwork(t *testing.T) {
	networks, err := FetchAllNetworks(true)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(len(networks))
	t.Log(networks)
}

func TestFindCustomEndpoint(t *testing.T) {
	var endpoint Endpoint
	result := findCustomEndpoint(&endpoint, "matic")
	if err := result.Error; err != nil {
		t.Error(err)
	}
	t.Log(result.RowsAffected)
	t.Log(endpoint.URL)
}
