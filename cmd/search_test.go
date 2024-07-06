package cmd

import (
	"testing"
	"time"
)

func Test_minAgeToBornBefore(t *testing.T) {
	tests := map[string]struct {
		minAge   int
		today    string
		expected string
	}{
		"born before 18 years ago":     {minAge: 18, today: "2021-01-01", expected: "2003-01-01"},
		"leap year today":              {minAge: 18, today: "2020-02-29", expected: "2002-03-01"},
		"person born on 29th February": {minAge: 18, today: "2022-03-01", expected: "2004-02-29"},
	}
	for name, test := range tests {
		expected, err := time.Parse("2006-01-02", test.expected)
		if err != nil {
			t.Fatalf("Unable to parse expected time %v", err)
		}
		today, err := time.Parse("2006-01-02", test.today)
		if err != nil {
			t.Fatalf("Unable to parse today time %v", err)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := minAgeToBornBefore(test.minAge, today)
			if actual != expected {
				t.Errorf("Expected %s, got %s", test.expected, actual)
			}
		})
	}
}

func Test_maxAgeToBornAfter(t *testing.T) {
	tests := map[string]struct {
		maxAge   int
		today    string
		expected string
	}{
		"max age 18": {maxAge: 18, today: "2021-01-01", expected: "2002-01-02"},
	}
	for name, test := range tests {
		expected, err := time.Parse("2006-01-02", test.expected)
		if err != nil {
			t.Fatalf("Unable to parse expected time %v", err)
		}
		today, err := time.Parse("2006-01-02", test.today)
		if err != nil {
			t.Fatalf("Unable to parse today time %v", err)
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			actual := maxAgeToBornAfter(test.maxAge, today)
			if actual != expected {
				t.Errorf("Expected %s, got %s", test.expected, actual)
			}
		})
	}

}
