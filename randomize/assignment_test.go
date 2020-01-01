package randomize

import (
	"fmt"
	"testing"
	"time"
)

func xxTestAssignments2(t *testing.T) {

	va1 := Variable{
		Name:   "BMI",
		Levels: []string{"low", "high"},
		Weight: 1,
	}

	proj := &Project{
		GroupNames:    []string{"treatment", "control"},
		Variables:     []Variable{va1},
		CellTotals:    make([]float64, 4),
		Assignments:   make([]int, 2),
		SamplingRates: []float64{1, 3},
		Bias:          5,
	}

	// Randomize 100 subjects
	for i := 0; i < 100; i++ {
		mpv := map[string]string{"BMI": "low"}
		if i%2 == 1 {
			mpv["BMI"] = "high"
		}
		time.Sleep(3)
		_, err := proj.doAssignment(mpv, fmt.Sprintf("%d", i), "user")
		if err != nil {
			panic(err)
		}
	}
}

func arange(x []float64) float64 {

	var mn, mx float64
	for i, u := range x {
		if i == 0 || u < mn {
			mn = u
		}
		if i == 0 || u > mx {
			mx = u
		}
	}

	return mx - mn
}

func fmax(x []float64) float64 {

	var v float64
	for i, y := range x {
		if i == 0 || y > v {
			v = y
		}
	}
	return v
}

func maxrange(proj *Project) []float64 {

	nvar := len(proj.Variables)
	ngrp := len(proj.GroupNames)
	nlev := len(proj.Variables[0].Levels)

	var rv []float64
	for v := 0; v < nvar; v++ {
		for l := 0; l < nlev; l++ {
			var xl []float64
			for g := 0; g < ngrp; g++ {
				x := proj.GetData(v, l, g)
				xl = append(xl, x)
			}
			rv = append(rv, arange(xl))
		}
	}

	return rv
}

func checkAssignment3(bias int) float64 {

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

	// Randomize 100 subjects
	mx := 0.0
	for i := 0; i < 100; i++ {
		mpv := map[string]string{"BMI": "low", "Age": "<20"}
		if i%3 == 1 {
			mpv["BMI"] = "high"
			mpv["Age"] = "20-50"
		} else if i%3 == 2 {
			mpv["Age"] = "50+"
		}
		time.Sleep(3)
		_, err := proj.doAssignment(mpv, fmt.Sprintf("%d", i), "user")
		if err != nil {
			panic(err)
		}

		if i < 10 {
			continue
		}

		m := fmax(maxrange(proj))
		r := m / float64(i+1)
		if i == 10 || r > mx {
			mx = r

			if mx > 0.2 {
				fmt.Printf("relative range=%v\n", mx)
				fmt.Printf("range=%v\n", m)
				proj.PrintData()
			}
		}
	}

	return mx
}

func TestAssignments3(t *testing.T) {

	bias := 5

	for j := 0; j < 10; j++ {
		_ := checkAssignment3(bias)
	}
}
