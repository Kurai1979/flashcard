package flashcard

import (
	"fmt"
	"strings"
)

type Rating int16

const (
	RatingAgain Rating = 1
	RatingHard  Rating = 2
	RatingGood  Rating = 3
	RatingEasy  Rating = 4
)

func (r Rating) String() string {
	switch r {
	case RatingAgain:
		return "Again"
	case RatingHard:
		return "Hard"
	case RatingGood:
		return "Good"
	case RatingEasy:
		return "Easy"
	}

	return "Unknown"
}

func ParseRating(s string) (Rating, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "again", "1":
		return RatingAgain, nil
	case "hard", "2":
		return RatingHard, nil
	case "good", "3":
		return RatingGood, nil
	case "easy", "4":
		return RatingEasy, nil
	}
	return 0, fmt.Errorf("invalid rating: %q", s)
}

func (r Rating) Valid() bool {
	return r >= RatingAgain && r <= RatingEasy
}
