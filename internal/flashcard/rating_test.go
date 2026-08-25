package flashcard

import "testing"

// Every rating must parse back from its own display name, so a rename that
// touches String but not ParseRating (or vice versa) fails here.
func TestRatingRoundTrip(t *testing.T) {
	for r := RatingAgain; r <= RatingEasy; r++ {
		got, err := ParseRating(r.String())
		if err != nil {
			t.Errorf("ParseRating(%q): %v", r.String(), err)
			continue
		}
		if got != r {
			t.Errorf("ParseRating(%q) = %d, want %d", r.String(), got, r)
		}
	}
}

func TestParseRating(t *testing.T) {
	cases := []struct {
		in   string
		want Rating
	}{
		{"again", RatingAgain},
		{"Again", RatingAgain},
		{"AGAIN", RatingAgain},
		{"  again  ", RatingAgain},
		{"1", RatingAgain},
		{"hard", RatingHard},
		{"2", RatingHard},
		{"good", RatingGood},
		{"3", RatingGood},
		{"easy", RatingEasy},
		{"4", RatingEasy},
	}

	for _, c := range cases {
		got, err := ParseRating(c.in)
		if err != nil {
			t.Errorf("ParseRating(%q): %v", c.in, err)
			continue
		}
		if got != c.want {
			t.Errorf("ParseRating(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

func TestParseRating_Invalid(t *testing.T) {
	for _, in := range []string{"", "   ", "0", "5", "-1", "gud", "again hard"} {
		if _, err := ParseRating(in); err == nil {
			t.Errorf("ParseRating(%q): expected error", in)
		}
	}
}

func TestRatingValid(t *testing.T) {
	for r := RatingAgain; r <= RatingEasy; r++ {
		if !r.Valid() {
			t.Errorf("Rating(%d).Valid() = false, want true", r)
		}
	}

	for _, r := range []Rating{0, -1, 5, 100} {
		if r.Valid() {
			t.Errorf("Rating(%d).Valid() = true, want false", r)
		}
	}
}

func TestRatingString_Unknown(t *testing.T) {
	if got := Rating(99).String(); got != "Unknown" {
		t.Errorf("Rating(99).String() = %q, want %q", got, "Unknown")
	}
}
