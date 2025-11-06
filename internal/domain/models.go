package domain

type Airport struct {
	SiteNumber    string `json:"site_number"`
	FacilityName  string `json:"facility_name"`
	Faa           string `json:"faa"`
	Icao          string `json:"icao"`
	StateCode     string `json:"state_code"`
	StateFull     string `json:"state_full"`
	County        string `json:"county"`
	City          string `json:"city"`
	OwnershipType string `json:"ownership_type"`
	UseType       string `json:"use_type"`
	Manager       string `json:"manager"`
	ManagerPhone  string `json:"manager_phone"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	AirportStatus string `json:"airport_status"`
	Weather       string `json:"weather"`
}

// Helper for parsing Weather API JSON
type WeatherResponse struct {
	Current struct {
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}
