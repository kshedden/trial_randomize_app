package randomize

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

func cumsum(x []float64) []float64 {
	y := make([]float64, len(x))
	copy(y, x)
	for j := 1; j < len(x); j++ {
		y[j] += y[j-1]
	}
	return y
}

func sample(rgen *rand.Rand, cumprob []float64) int {
	ur := rgen.Float64()
	jr := 0
	for ii, x := range cumprob {
		if x > ur {
			jr = ii
			break
		}
	}

	return jr
}

func genPocockSimon(n, bias int) []float64 {
	qmin := 1 / float64(n)
	qmax := 2 / float64(n-1)
	qq := qmin + float64(bias-1)*(qmax-qmin)/9.0
	prob := make([]float64, n)
	nf := float64(n)
	for j := range prob {
		prob[j] = qq - 2*(nf*qq-1)*float64(j+1)/(nf*(nf+1))
	}
	return prob
}

// doAssignment
func (proj *Project) doAssignment(mpv map[string]string, subjectId string, userId string) (string, error) {

	// Set the seed to a random time.  Not sure if this is needed,
	// but since each assignment runs as a new instance we might
	// be getting the same "random numbers" every time if we don't
	// do this.
	source := rand.NewSource(time.Now().UnixNano())
	rgen := rand.New(source)

	numgroups := len(proj.GroupNames)
	numvar := len(proj.Variables)

	// Calculate the scores if assigning the new subject
	// to each possible group.
	potentialScores := make([]float64, numgroups)
	for i := 0; i < numgroups; i++ {

		// The score is a weighted linear combination over the
		// variables.
		for j, va := range proj.Variables {
			x := mpv[va.Name]
			score := proj.Score(x, i, j)
			potentialScores[i] += va.Weight * score
		}
	}

	// Get a sorted copy of the scores.
	sortedScores := make([]float64, len(potentialScores))
	copy(sortedScores, potentialScores)
	sort.Float64s(sortedScores)

	// Construct the Pocock/Simon probabilities.
	prob := genPocockSimon(len(proj.GroupNames), proj.Bias)

	// The cumulative Pocock Simon probabilities.
	cumprob := cumsum(prob)

	// A random value distributed according to the Pocock Simon
	// probabilities.
	jr := sample(rgen, cumprob)

	// Get all groups whose score is tied with the score of the selected value.
	var ties []int
	for i, x := range potentialScores {
		if x == sortedScores[jr] {
			ties = append(ties, i)
		}
	}

	// Assign to this group.
	ii := ties[rgen.Intn(len(ties))]

	// Update the cell totals.
	proj.Assignments[ii]++
	for j := 0; j < numvar; j++ {

		va := proj.Variables[j]
		x, ok := mpv[va.Name]
		if !ok {
			msg := fmt.Sprintf("Variable '%s' not found", va.Name)
			return "", fmt.Errorf(msg)
		}

		kk := -1
		for k, v := range va.Levels {
			if x == v {
				kk = k
				break
			}
		}
		if kk == -1 {
			return "", fmt.Errorf("Invalid state in DoAssignment")
		}

		z := proj.GetData(j, kk, ii)
		proj.SetData(j, kk, ii, z+1)
	}

	// Update the stored data
	if proj.StoreRawData {

		data := make([]string, len(proj.Variables))
		for j, v := range proj.Variables {
			data[j] = mpv[v.Name]
		}

		rec := DataRecord{
			SubjectId:     subjectId,
			AssignedTime:  time.Now(),
			AssignedGroup: proj.GroupNames[ii],
			CurrentGroup:  proj.GroupNames[ii],
			Included:      true,
			Data:          data,
			Assigner:      userId,
		}

		proj.RawData = append(proj.RawData, &rec)
	}

	return proj.GroupNames[ii], nil
}

// Score calculates the contribution to the overall score if we assign
// a subject with level `x` for the kth variable into group `grp`.
func (proj *Project) Score(x string, grp, k int) float64 {

	numGroups := len(proj.GroupNames)
	va := proj.Variables[k]

	scoreChange := 0.0
	for j := range va.Levels {

		if x != va.Levels[j] {
			continue
		}

		// Get the count for each group if we were to assign
		// this unit to group `grp`.
		var mn, mx float64
		for i := 0; i < numGroups; i++ {

			// The current count for variable k, level j, group i.
			nc := proj.GetData(k, j, i)

			// Add 1 if we are assigning the current subject to this
			// group.
			if i == grp {
				nc++
			}

			nc /= proj.SamplingRates[i]
			if i == 0 || nc < mn {
				mn = nc
			}
			if i == 0 || nc > mx {
				mx = nc
			}
		}

		scoreChange += mx - mn
	}

	return scoreChange
}
