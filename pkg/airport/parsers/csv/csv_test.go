package csv

import (
	"bytes"
	"reflect"
	"strings"
	"testing"

	"github.com/enescakir/emoji"
	"github.com/tarmac-project/example-airport-lookup-go/pkg/airport"
)

type RecordToAirportTestCase struct {
	name    string
	record  []string
	airport airport.Airport
	err     error
}

func TestRecordToAirport(t *testing.T) {
	tt := []RecordToAirportTestCase{
		{
			name: "Heliport",
			record: []string{
				"6523",
				"00A",
				"heliport",
				"Total RF Heliport",
				"40.070985",
				"-74.933689",
				"11",
				"NA",
				"US",
				"US-PA",
				"Bensalem",
				"no",
				"K00A",
				"",
				"00A",
				"https://www.penndot.pa.gov/TravelInPA/airports-pa/Pages/Total-RF-Heliport.aspx",
				"",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "heliport",
				TypeEmoji:    emoji.Helicopter.String(),
				Status:       "open",
			},
			err: nil,
		},
		{
			name: "Small airport",
			record: []string{
				"6523",
				"00A",
				"small_airport",
				"Total RF Heliport",
				"40.070985",
				"-74.933689",
				"11",
				"NA",
				"US",
				"US-PA",
				"Bensalem",
				"no",
				"K00A",
				"",
				"00A",
				"https://www.penndot.pa.gov/TravelInPA/airports-pa/Pages/Total-RF-Heliport.aspx",
				"",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "small_airport",
				TypeEmoji:    emoji.SmallAirplane.String(),
				Status:       "open",
			},
			err: nil,
		},
		{
			name: "Balloonport Type",
			record: []string{
				"6523", "00A", "balloonport", "Total RF Heliport", "40.070985", "-74.933689", "11", "NA", "US", "US-PA", "Bensalem", "no", "K00A", "", "00A", "https://www.example.com", "",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "balloonport",
				TypeEmoji:    emoji.Balloon.String(),
				Status:       "open",
			},
			err: nil,
		},
		{
			name: "Seaplane Base Type",
			record: []string{
				"6523", "00A", "seaplane_base", "Total RF Heliport", "40.070985", "-74.933689", "11", "NA", "US", "US-PA", "Bensalem", "no", "K00A", "", "00A", "https://www.example.com", "",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "seaplane_base",
				TypeEmoji:    emoji.Anchor.String(),
				Status:       "open",
			},
			err: nil,
		},
		{
			name: "Medium Airport Type",
			record: []string{
				"6523", "00A", "medium_airport", "Total RF Heliport", "40.070985", "-74.933689", "11", "NA", "US", "US-PA", "Bensalem", "no", "K00A", "", "00A", "https://www.example.com", "",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "medium_airport",
				TypeEmoji:    emoji.Airplane.String(),
				Status:       "open",
			},
			err: nil,
		},
		{
			name: "Large Airport Type",
			record: []string{
				"6523", "00A", "large_airport", "Total RF Heliport", "40.070985", "-74.933689", "11", "NA", "US", "US-PA", "Bensalem", "no", "K00A", "", "00A", "https://www.example.com", "",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "large_airport",
				TypeEmoji:    emoji.Airplane.String(),
				Status:       "open",
			},
			err: nil,
		},
		{
			name: "Closed airport",
			record: []string{
				"6523",
				"00A",
				"closed",
				"Total RF Heliport",
				"40.070985",
				"-74.933689",
				"11",
				"NA",
				"US",
				"US-PA",
				"Bensalem",
				"no",
				"K00A",
				"",
				"00A",
				"https://www.penndot.pa.gov/TravelInPA/airports-pa/Pages/Total-RF-Heliport.aspx",
				"",
			},
			airport: airport.Airport{
				Continent:    "NA",
				Emoji:        emoji.FlagForUnitedStates.String(),
				ISOCountry:   "US",
				ISORegion:    "US-PA",
				LocalCode:    "00A",
				Municipality: "Bensalem",
				Name:         "Total RF Heliport",
				Type:         "unknown",
				TypeEmoji:    emoji.Construction.String(),
				Status:       "closed",
			},
			err: nil,
		},
		{
			name: "CSV header",
			record: []string{
				"id",
				"ident",
				"type",
				"name",
				"latitude_deg",
				"longitude_deg",
				"elevation_ft",
				"continent",
				"iso_country",
				"iso_region",
				"municipality",
				"scheduled_service",
				"gps_code",
				"iata_code",
				"local_code",
				"home_link",
				"wikipedia_link",
				"keywords",
			},
			airport: airport.Airport{}, // This should result in an empty airport struct
			err:     ErrIsHeader,
		},
		{
			name: "Invalidly formatted airport record",
			record: []string{
				"6523",
				"00A",
				"heliport",
				"Total RF Heliport",
				"40.070985",
				"-74.933689",
				"11",
				"NA",
				"US",
				"US-PA",
				"Bensalem",
				"no",
				"K00A",
				"00A",
			}, // Missing fields
			airport: airport.Airport{}, // This should result in an empty airport struct
			err:     ErrNotEnoughFields,
		},
		{
			name:    "Empty CSV Row",
			record:  []string{},        // An empty row
			airport: airport.Airport{}, // This should result in an empty airport struct
			err:     ErrNotEnoughFields,
		},
		{
			name: "Minimum Required Fields",
			record: []string{
				"6523",
				"00A",
				"",
				"Total RF Heliport",
				"",
				"",
				"",
				"",
				"US",
				"",
				"",
				"",
				"",
				"",
				"00A",
				"",
				"",
				"",
				"",
			}, // Minimum required fields with empty values
			airport: airport.Airport{
				ISOCountry: "US",
				Emoji:      emoji.FlagForUnitedStates.String(),
				LocalCode:  "00A",
				Name:       "Total RF Heliport",
				Type:       "",
				TypeEmoji:  "❓",
				Status:     "unknown",
			},
			err: nil,
		},
	}

	for _, tc := range tt {
		t.Run("RecordToAirport: "+tc.name, func(t *testing.T) {
			results, err := RecordToAirport(tc.record)
			if err != tc.err {
				t.Fatalf("Expected error: %v, got error: %v", tc.err, err)
			}

			if !reflect.DeepEqual(results, tc.airport) {
				t.Errorf("Expected airport: %v, got airport: %v", tc.airport, results)
			}
		})
	}
}

type ParseAirportTestCase struct {
	name    string
	raw     []byte
	records int
	pass    bool
}

func TestParseAirport(t *testing.T) {
	tt := []ParseAirportTestCase{
		{
			name:    "Empty file",
			raw:     []byte(""), // Empty file
			records: 0,
			pass:    true,
		},
		{
			name: "Single record",
			raw: []byte(
				"6523,00A,heliport,Total RF Heliport,40.070985,-74.933689,11,NA,US,US-PA,Bensalem,no,K00A,,00A,https://www.example.com,",
			),
			records: 1,
			pass:    true,
		},
		{
			name: "Multiple records",
			raw: []byte(
				"6523,00A,heliport,Total RF Heliport,40.070985,-74.933689,11,NA,US,US-PA,Bensalem,no,K00A,,00A,https://www.example.com,\n6524,00B,heliport,Total RF Heliport,40.070985,-74.933689,11,NA,US,US-PA,Bensalem,no,K00A,,00A,https://www.example.com,",
			),
			records: 2,
			pass:    true,
		},
		{
			name: "Header only",
			raw: []byte(
				"id,ident,type,name,latitude_deg,longitude_deg,elevation_ft,continent,iso_country,iso_region,municipality,scheduled_service,gps_code,iata_code,local_code,home_link,wikipedia_link,keywords",
			),
			records: 0,
			pass:    true,
		},
		{
			name: "Invalid record",
			raw: []byte(
				"6523,00A,heliport,Total RF Heliport,40.070985,-74.933689,11,NA,US,US-PA,Bensal",
			),
			records: 0,
			pass:    false,
		},
	}

	for _, tc := range tt {
		t.Run("ParseAirport: "+tc.name, func(t *testing.T) {
			// Create a new parser
			p, err := New(bytes.NewReader(tc.raw))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Parse the file
			results, err := p.Parse()
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Validate the records returned
			if len(results) != tc.records {
				t.Errorf("Expected %d records, got %d", tc.records, len(results))
			}
		})
	}
}

func TestParseReturnsCSVReadError(t *testing.T) {
	p, err := New(strings.NewReader("\"unterminated"))
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if _, err := p.Parse(); err == nil {
		t.Fatal("Parse() error = nil, want malformed CSV error")
	}
}

func FuzzParserParse(f *testing.F) {
	f.Add([]byte(""))
	f.Add([]byte("id,ident,type,name,latitude_deg,longitude_deg,elevation_ft,continent,iso_country,iso_region,municipality,scheduled_service,gps_code,iata_code,local_code,home_link,wikipedia_link,keywords\n"))
	f.Add([]byte("1,CIX1,small_airport,CI Fixture Airport,0,0,0,NA,US,US-AZ,Phoenix,no,KCIX,,CIX1,,,\n"))
	f.Add([]byte("\"unterminated"))

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			t.Skip()
		}

		p, err := New(bytes.NewReader(data))
		if err != nil {
			t.Fatalf("New() unexpected error: %v", err)
		}
		_, _ = p.Parse()
	})
}
