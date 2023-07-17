package main

import (
	"testing"
	"time"
)

// TestHumanDate1 checks that the HumanDate template function returns a date in the expected format
func TestHumanDate1(t *testing.T) {
	const expected = "17 Dec 2020 at 10:00"
	// Initialize a new time.Time object and pass it to the humanDate function.
	tm := time.Date(2020, 12, 17, 10, 0, 0, 0, time.UTC)
	hd := humanDate(tm)

	// Check that the output from the humanDate function is in the format we
	// expect. If it isn't what we expect, use the t.Errorf() function to
	// indicate that the test has failed and log the expected and actual
	// values.
	if hd != expected {
		t.Errorf("want %q; got %q", expected, hd)
	}
}

// TestHumanDate checks that HumanDate (added to template function map) displays
// dates correctly.  It is a refactoring of the above TestHumanDate1 to be a
// table-driven test so we can easily test lots of different time.Time values
func TestHumanDate(t *testing.T) {
	tests := map[string]struct {
		tm   time.Time
		want string
	}{
		"2006": {
			tm:   time.Date(2006, 2, 1, 15, 4, 5, 0, time.UTC),
			want: "01 Feb 2006 at 15:04",
		},
		"1999": {
			tm:   time.Date(1999, 12, 17, 10, 0, 0, 0, time.UTC),
			want: "17 Dec 1999 at 10:00",
		},
		"2099": {
			tm:   time.Date(2099, 12, 17, 10, 0, 0, 0, time.UTC),
			want: "17 Dec 2099 at 10:00",
		},
		"Empty": {
			tm:   time.Time{},
			want: "",
		},
		"CET": {
			tm:   time.Date(2020, 12, 17, 10, 0, 0, 0, time.FixedZone("CET", 1*60*60)),
			want: "17 Dec 2020 at 09:00",
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			got := humanDate(tt.tm)

			if got != tt.want {
				t.Errorf("want %q; got %q", tt.want, got)
			}
		})
	}
}
