package mysql

import (
	"reflect"
	"testing"
)

func TestSplitSQLStatements(t *testing.T) {
	raw := `-- comment
CREATE TABLE test_one (
    id BIGINT NOT NULL
);

CREATE TABLE test_two (
    id BIGINT NOT NULL
);
`

	got, err := splitSQLStatements(raw)
	if err != nil {
		t.Fatalf("splitSQLStatements returned error: %v", err)
	}

	want := []string{
		"CREATE TABLE test_one (\n    id BIGINT NOT NULL\n)",
		"CREATE TABLE test_two (\n    id BIGINT NOT NULL\n)",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("splitSQLStatements() = %#v, want %#v", got, want)
	}
}

func TestSplitSQLStatementsRejectsUnterminatedSQL(t *testing.T) {
	_, err := splitSQLStatements("CREATE TABLE broken (id BIGINT NOT NULL)\n")
	if err == nil {
		t.Fatal("expected unterminated SQL error")
	}
}
