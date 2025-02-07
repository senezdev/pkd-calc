package calc_test

import (
	"strconv"
	"testing"

	"pkd-bot/calc"
)

func TestCalcTwoBoost(t *testing.T) {
	testCases := []struct {
		input []string
	}{
		{
			input: []string{
				"Around Pillars",
				"Fortress",
				"Blocks",
				"Ice",
				"Tightrope",
				"Sandpit",
				"Tower Tightrope",
				"Fences",
				"Finish Room",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res, boostRooms, err := calc.CalcTwoBoost(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			t.Log(res)
			t.Log(boostRooms)
			t.Fail()
		})
	}
}

func TestCalcThreeBoost(t *testing.T) {
	testCases := []struct {
		input []string
	}{
		{
			input: []string{
				"Around Pillars",
				"Fortress",
				"Blocks",
				"Ice",
				"Tightrope",
				"Sandpit",
				"Tower Tightrope",
				"Fences",
				"Finish Room",
			},
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			res, boostRooms, err := calc.CalcThreeBoost(tc.input)
			if err != nil {
				t.Fatal(err)
			}

			t.Log(res)
			t.Log(boostRooms)
			t.Fail()
		})
	}
}
