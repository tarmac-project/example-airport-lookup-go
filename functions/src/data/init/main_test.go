package main

import (
	"errors"
	"strings"
	"testing"

	"github.com/tarmac-project/sdk/sql"
)

type fakeFunctionClient struct {
	name  string
	input []byte
	rsp   []byte
	err   error
}

func (f *fakeFunctionClient) Call(name string, input []byte) ([]byte, error) {
	f.name = name
	f.input = append([]byte(nil), input...)
	return f.rsp, f.err
}

type fakeSQLClient struct {
	queries []string
	err     error
}

func (f *fakeSQLClient) Exec(query string) (sql.ExecResult, error) {
	f.queries = append(f.queries, query)
	return sql.ExecResult{}, f.err
}

func (*fakeSQLClient) Query(string) (sql.QueryResult, error) { return sql.QueryResult{}, nil }
func (*fakeSQLClient) Close() error                          { return nil }

type fakeLogger struct{}

func (*fakeLogger) Info(string)  {}
func (*fakeLogger) Warn(string)  {}
func (*fakeLogger) Error(string) {}
func (*fakeLogger) Debug(string) {}
func (*fakeLogger) Trace(string) {}

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		sqlErr     error
		seedRsp    []byte
		seedErr    error
		wantErrSub string
	}{
		{name: "initializes and seeds", seedRsp: []byte("{\"successful_upsert\":20}")},
		{name: "create table fails", sqlErr: errors.New("database unavailable"), wantErrSub: "failed to create table"},
		{name: "seed fails", seedErr: errors.New("seed unavailable"), wantErrSub: "failed to seed airport data"},
		{name: "empty seed response fails", wantErrSub: "empty summary payload"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			sqlClient := &fakeSQLClient{err: tc.sqlErr}
			functionClient := &fakeFunctionClient{rsp: tc.seedRsp, err: tc.seedErr}
			function := &Function{
				logging:  &fakeLogger{},
				function: functionClient,
				sql:      sqlClient,
			}

			_, err := function.Handler(nil)
			if tc.wantErrSub == "" && err != nil {
				t.Fatalf("Handler() unexpected error: %v", err)
			}
			if tc.wantErrSub != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErrSub)) {
				t.Fatalf("Handler() error = %v, want substring %q", err, tc.wantErrSub)
			}
			if len(sqlClient.queries) == 0 || !strings.Contains(sqlClient.queries[0], "CREATE TABLE IF NOT EXISTS airports") {
				t.Fatalf("expected schema creation query, got %v", sqlClient.queries)
			}
			if tc.sqlErr == nil && functionClient.name != "seed" {
				t.Fatalf("function call = %q, want seed", functionClient.name)
			}
		})
	}
}
