package randomize

import (
	"fmt"
	"testing"
	"time"
)

// arange returns the numerical range (maximum minus minimum)
// for the values in the given array.
func arange(x []float64) float64 {

	var mn, mx, n float64
	for i, u := range x {
		n += u
		if i == 0 || u < mn {
			mn = u
		}
		if i == 0 || u > mx {
			mx = u
		}
	}

	return (mx - mn) / n
}

// fmax returns the maximum value of the provided array.
func fmax(x []float64) float64 {

	var v float64
	for i, y := range x {
		if i == 0 || y > v {
			v = y
		}
	}
	return v
}

// Returns the maximum difference in assignment counts
// between any two arms, for people in specific levels
// of specific variables.
func maxrange(proj *Project) []float64 {

	nvar := len(proj.Variables)
	ngrp := len(proj.GroupNames)

	var rv []float64
	for v := 0; v < nvar; v++ {
		nlev := len(proj.Variables[v].Levels)
		for l := 0; l < nlev; l++ {
			var xl []float64
			for g := 0; g < ngrp; g++ {

				// Number of people with level l of variable v
				// who are assigned to group g.
				x := proj.GetData(v, l, g)
				xl = append(xl, x)
			}
			rv = append(rv, arange(xl))
		}
	}

	return rv
}

func checkAssignment(bias int) float64 {

	// Randomize nn subjects
	nn := 200

	// Check balance after this many subjects have been assigned
	firstCheck := 50

	// Print information when the relative balance is worse than
	// this value.
	rmax := 0.1

	va1 := Variable{
		Name:   "BMI",
		Levels: []string{"low", "high"},
		Weight: 1,
	}

	va2 := Variable{
		Name:   "Age",
		Levels: []string{"<20", "20-50", "50+"},
		Weight: 1,
	}

	proj := &Project{
		GroupNames:    []string{"A", "B", "C"},
		Variables:     []Variable{va1, va2},
		CellTotals:    make([]float64, 18),
		Assignments:   make([]int, 3),
		SamplingRates: []float64{1, 1, 1},
		Bias:          bias,
	}

	// Randomize nn subjects
	mx := 0.0
	for i := 0; i < nn; i++ {

		// Create the covariates for one subject
		mpv := make(map[string]string)
		switch i%3 {
		   case 0:
				mpv["BMI"] = "low"
				mpv["Age"] = "<20"
			case 1:
				mpv["BMI"] = "high"
				mpv["Age"] = "20-50"
			case 2:
				mpv["BMI"] = "low"
				mpv["Age"] = "50+"
			default:
				panic("!!")
		}

		time.Sleep(3)
		_, err := proj.doAssignment(mpv, fmt.Sprintf("%d", i), "user")
		if err != nil {
			panic(err)
		}

		// The balance is never good in the first few steps so skip.
		if i < firstCheck {
			continue
		}

		r := fmax(maxrange(proj))
		if i == firstCheck || r > mx {
			mx = r

			// Only print results when the balance is poor
			if mx > rmax {
				fmt.Printf("i=%d\n", i)
				fmt.Printf("%+v\n", maxrange(proj))
				fmt.Printf("relative range=%v\n", r)
				proj.PrintData()
			}
		}
	}

	return mx
}

func TestAssignments(t *testing.T) {

	bias := 10

	for j := 0; j < 10; j++ {
		checkAssignment(bias)
	}
}
