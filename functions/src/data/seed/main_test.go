package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/tarmac-project/sdk/sql"
)

type fakeSQLClient struct {
	calls  int
	failAt int
}

func (f *fakeSQLClient) Exec(string) (sql.ExecResult, error) {
	f.calls++
	if f.calls == f.failAt {
		return sql.ExecResult{}, errors.New("upsert failed")
	}
	return sql.ExecResult{RowsAffected: 1}, nil
}

func (*fakeSQLClient) Query(string) (sql.QueryResult, error) { return sql.QueryResult{}, nil }
func (*fakeSQLClient) Close() error                          { return nil }

type fakeLogger struct{}

func (*fakeLogger) Info(string)  {}
func (*fakeLogger) Warn(string)  {}
func (*fakeLogger) Error(string) {}
func (*fakeLogger) Debug(string) {}
func (*fakeLogger) Trace(string) {}

func TestSeedEscapeSQL(t *testing.T) {
	input := "Chicago O'Hare"
	got := escapeSQL(input)

	if got != "Chicago O''Hare" {
		t.Fatalf("expected doubled single quote escaping, got %q", got)
	}
}

func TestBuildSeedUpsertQuery(t *testing.T) {
	query := buildSeedUpsertQuery(airportSeed{
		LocalCode:    "ORD",
		Name:         "Chicago O'Hare International Airport",
		Type:         "large_airport",
		TypeEmoji:    "🛫",
		Continent:    "NA",
		ISOCountry:   "US",
		ISORegion:    "US-IL",
		Municipality: "Chicago",
		Emoji:        "🇺🇸",
		Status:       "open",
	})

	if !strings.Contains(query, "INSERT INTO airports") {
		t.Fatalf("expected insert statement, got %q", query)
	}
	if !strings.Contains(query, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("expected upsert statement, got %q", query)
	}
	if !strings.Contains(query, "Chicago O''Hare International Airport") {
		t.Fatalf("expected escaped airport name, got %q", query)
	}
}

func TestMarshalSeedSummary(t *testing.T) {
	payload := marshalSeedSummary(seedSummary{
		SeededAirports: 20,
		SuccessUpsert:  20,
		FailedUpsert:   0,
	})

	got := string(payload)
	for _, expected := range []string{
		`"seeded_airports":20`,
		`"successful_upsert":20`,
		`"failed_upsert":0`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected summary to contain %q, got %q", expected, got)
		}
	}
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		failAt     int
		wantErrSub string
		wantFailed string
	}{
		{name: "seeds every airport", wantFailed: "\"failed_upsert\":0"},
		{name: "reports partial failure", failAt: 3, wantErrSub: "1 failed upserts", wantFailed: "\"failed_upsert\":1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sqlClient := &fakeSQLClient{failAt: tc.failAt}
			function := &Function{logging: &fakeLogger{}, sql: sqlClient}

			summary, err := function.Handler(nil)
			if tc.wantErrSub == "" && err != nil {
				t.Fatalf("Handler() unexpected error: %v", err)
			}
			if tc.wantErrSub != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErrSub)) {
				t.Fatalf("Handler() error = %v, want substring %q", err, tc.wantErrSub)
			}
			if sqlClient.calls != len(primaryUSAirports) {
				t.Fatalf("Exec calls = %d, want %d", sqlClient.calls, len(primaryUSAirports))
			}
			if !strings.Contains(string(summary), tc.wantFailed) {
				t.Fatalf("summary = %s, want substring %q", summary, tc.wantFailed)
			}
		})
	}
}
