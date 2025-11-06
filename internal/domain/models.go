package domain

type Airport struct {
	SiteNumber    string `json:"site_number"`
	FacilityName  string `json:"facility_name"`
	Faa           string `json:"faa_ident"`
	Icao          string `json:"icao_ident"`
	StateCode     string `json:"state"`
	StateFull     string `json:"state_full"`
	County        string `json:"county"`
	City          string `json:"city"`
	OwnershipType string `json:"ownership"`
	UseType       string `json:"use"`
	Manager       string `json:"manager"`
	ManagerPhone  string `json:"manager_phone"`
	Latitude      string `json:"latitude"`
	Longitude     string `json:"longitude"`
	AirportStatus string `json:"status"`
	Weather       string `json:"weather"`
}

// WeatherResponse represents the structure of data returned by the weather API.
type WeatherResponse struct {
	Current struct {
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
}
