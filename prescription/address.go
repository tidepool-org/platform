package prescription

import (
	"regexp"

	"github.com/tidepool-org/platform/structure"
)

const (
	CountryCodeUS = "US"

	StateAlabama                     = "AL"
	StateAlaska                      = "AK"
	StateAmericanSamoa               = "AS"
	StateArizona                     = "AZ"
	StateArkansas                    = "AR"
	StateCalifornia                  = "CA"
	StateColorado                    = "CO"
	StateConnecticut                 = "CT"
	StateDelaware                    = "DE"
	StateDistrictOfColumbia          = "DC"
	StateFederatedStatesOfMicronesia = "FM"
	StateFlorida                     = "FL"
	StateGeorgia                     = "GA"
	StateGuam                        = "GU"
	StateHawaii                      = "HI"
	StateIdaho                       = "ID"
	StateIllinois                    = "IL"
	StateIndiana                     = "IN"
	StateIowa                        = "IA"
	StateKansas                      = "KS"
	StateKentucky                    = "KY"
	StateLouisiana                   = "LA"
	StateMaine                       = "ME"
	StateMarshallIslands             = "MH"
	StateMaryland                    = "MD"
	StateMassachusetts               = "MA"
	StateMichigan                    = "MI"
	StateMinnesota                   = "MN"
	StateMississippi                 = "MS"
	StateMissouri                    = "MO"
	StateMontana                     = "MT"
	StateNebraska                    = "NE"
	StateNevada                      = "NV"
	StateNewHampshire                = "NH"
	StateNewJersey                   = "NJ"
	StateNewMexico                   = "NM"
	StateNewYork                     = "NY"
	StateNorthCarolina               = "NC"
	StateNorthDakota                 = "ND"
	StateNorthernMarianaIslands      = "MP"
	StateOhio                        = "OH"
	StateOklahoma                    = "OK"
	StateOregon                      = "OR"
	StatePalau                       = "PW"
	StatePennsylvania                = "PA"
	StatePuertoRico                  = "PR"
	StateRhodeIsland                 = "RI"
	StateSouthCarolina               = "SC"
	StateSouthDakota                 = "SD"
	StateTennessee                   = "TN"
	StateTexas                       = "TX"
	StateUtah                        = "UT"
	StateVermont                     = "VT"
	StateVirginIslands               = "VI"
	StateVirginia                    = "VA"
	StateWashington                  = "WA"
	StateWestVirginia                = "WV"
	StateWisconsin                   = "WI"
	StateWyoming                     = "WY"
)

type Address struct {
	Line1      string `json:"line1,omitempty" bson:"line1,omitempty"`
	Line2      string `json:"line2,omitempty" bson:"line2,omitempty"`
	City       string `json:"city,omitempty" bson:"city,omitempty"`
	State      string `json:"state,omitempty" bson:"state,omitempty"`
	PostalCode string `json:"postalCode,omitempty" bson:"postalCode,omitempty"`
	Country    string `json:"country,omitempty" bson:"country,omitempty"`
}

func (a *Address) Validate(validator structure.Validator) {
	if a.State != "" {
		validator.String("state", &a.State).OneOf(USStates()...)
	}
	if a.PostalCode != "" {
		validator.String("postalCode", &a.PostalCode).Matches(postalCodeExpression)
	}
	if a.Country != "" {
		validator.String("country", &a.Country).OneOf(ValidCountryCodes()...)
	}
}

func (a *Address) ValidateAllRequired(validator structure.Validator) {
	validator.String("line1", &a.Line1).NotEmpty()
	validator.String("city", &a.City).NotEmpty()
	validator.String("state", &a.State).NotEmpty()
	validator.String("postalCode", &a.PostalCode).NotEmpty()
	validator.String("country", &a.Country).NotEmpty()
}

var postalCodeExpression = regexp.MustCompile("^[0-9a]{5}(-[0-9]{4})?")

func USStates() []string {
	return []string{
		StateAlabama,
		StateAlaska,
		StateAmericanSamoa,
		StateArizona,
		StateArkansas,
		StateCalifornia,
		StateColorado,
		StateConnecticut,
		StateDelaware,
		StateDistrictOfColumbia,
		StateFederatedStatesOfMicronesia,
		StateFlorida,
		StateGeorgia,
		StateGuam,
		StateHawaii,
		StateIdaho,
		StateIllinois,
		StateIndiana,
		StateIowa,
		StateKansas,
		StateKentucky,
		StateLouisiana,
		StateMaine,
		StateMarshallIslands,
		StateMaryland,
		StateMassachusetts,
		StateMichigan,
		StateMinnesota,
		StateMississippi,
		StateMissouri,
		StateMontana,
		StateNebraska,
		StateNevada,
		StateNewHampshire,
		StateNewJersey,
		StateNewMexico,
		StateNewYork,
		StateNorthCarolina,
		StateNorthDakota,
		StateNorthernMarianaIslands,
		StateOhio,
		StateOklahoma,
		StateOregon,
		StatePalau,
		StatePennsylvania,
		StatePuertoRico,
		StateRhodeIsland,
		StateSouthCarolina,
		StateSouthDakota,
		StateTennessee,
		StateTexas,
		StateUtah,
		StateVermont,
		StateVirginIslands,
		StateVirginia,
		StateWashington,
		StateWestVirginia,
		StateWisconsin,
		StateWyoming,
	}
}

func ValidCountryCodes() []string {
	return []string{
		CountryCodeUS,
	}
}
