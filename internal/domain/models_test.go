package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAirportJSONMarshalUnmarshal(t *testing.T) {
	// Sample Airport data
	expectedAirport := Airport{
		SiteNumber:    "12345",
		FacilityName:  "Test Airport",
		Faa:           "TST",
		Icao:          "KTST",
		StateCode:     "CA",
		StateFull:     "California",
		County:        "Test County",
		City:          "Test City",
		OwnershipType: "Public",
		UseType:       "Public Use",
		Manager:       "Test Manager",
		ManagerPhone:  "123-456-7890",
		Latitude:      "34.0522",
		Longitude:     "-118.2437",
		AirportStatus: "Open",
		Weather:       "Clear",
	}

	// Test Marshal (encoding, go -> data format)
	jsonBytes, err := json.Marshal(expectedAirport)
	assert.NoError(t, err, "Should marshal Airport without error")

	expectedJSON := `{"site_number":"12345","facility_name":"Test Airport","faa_ident":"TST","icao_ident":"KTST","state":"CA","state_full":"California","county":"Test County","city":"Test City","ownership":"Public","use":"Public Use","manager":"Test Manager","manager_phone":"123-456-7890","latitude":"34.0522","longitude":"-118.2437","status":"Open","weather":"Clear"}`
	assert.JSONEq(t, expectedJSON, string(jsonBytes), "Marshaled JSON should match expected")

	// Test Unmarshal (decoding, data format -> go)
	var actualAirport Airport
	err = json.Unmarshal(jsonBytes, &actualAirport)
	assert.NoError(t, err, "Should unmarshal Airport without error")

	assert.Equal(t, expectedAirport, actualAirport, "Unmarshaled Airport should match original")
}

func TestWeatherResponseJSONMarshalUnmarshal(t *testing.T) {
	// Sample WeatherResponse data
	expectedWeather := WeatherResponse{
		Current: struct {
			Condition struct {
				Text string `json:"text"`
			} `json:"condition"`
		}{
			Condition: struct {
				Text string `json:"text"`
			}{
				Text: "Sunny",
			},
		},
	}

	// Test Marshal (encoding, go -> data format)
	jsonBytes, err := json.Marshal(expectedWeather)
	assert.NoError(t, err, "Should marshal WeatherResponse without error")

	expectedJSON := `{"current":{"condition":{"text":"Sunny"}}}`
	assert.JSONEq(t, expectedJSON, string(jsonBytes), "Marshaled JSON should match expected")

	// Test Unmarshal (decoding, data format -> go)
	var actualWeather WeatherResponse
	err = json.Unmarshal(jsonBytes, &actualWeather)
	assert.NoError(t, err, "Should unmarshal WeatherResponse without error")

	assert.Equal(t, expectedWeather.Current.Condition.Text, actualWeather.Current.Condition.Text, "Unmarshaled WeatherResponse condition text should match")
}
