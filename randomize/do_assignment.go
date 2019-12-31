package randomize

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

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
	N := len(proj.GroupNames)
	qmin := 1 / float64(N)
	qmax := 2 / float64(N-1)
	qq := qmin + float64(proj.Bias-1)*(qmax-qmin)/9.0
	prob := make([]float64, N)
	for j := range prob {
		prob[j] = qq - 2*(float64(N)*qq-1)*float64(j+1)/float64(N*(N+1))
	}

	// The cumulative Pocock Simon probabilities.
	cumprob := make([]float64, N)
	copy(cumprob, prob)
	for j := 1; j < len(cumprob); j++ {
		cumprob[j] += cumprob[j-1]
	}

	// A random value distributed according to the Pocock Simon
	// probabilities.
	ur := rgen.Float64()
	jr := 0
	for ii, x := range cumprob {
		if x > ur {
			jr = ii
			break
		}
	}

	// Get all values tied with the selected value.
	var ties []int
	for i, x := range potentialScores {
		if x == sortedScores[jr] {
			ties = append(ties, i)
		}
	}

	// Assign to this group.
	ii := ties[rgen.Intn(len(ties))]

	// Update the project.
	proj.Assignments[ii]++
	for j := 0; j < numvar; j++ {

		VA := proj.Variables[j]
		x := mpv[VA.Name]

		kk := -1
		for k, v := range VA.Levels {
			if x == v {
				kk = k
				break
			}
		}
		if kk == -1 {
			return "", fmt.Errorf("Invalid state in Do_assignment")
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

// Score calculates the contribution to the overall score for a given variable
// `va` if we put a subject with data value `x` into group `grp`.
// `counts` contains the current cell counts for each level x group
// combination for this variable, `va` contains variable information.
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
			nc := proj.GetData(k, j, i)
			if i == grp {
				nc++
			}
			nc /= proj.SamplingRates[i]
			if nc < mn {
				mn = nc
			}
			if nc > mx {
				mx = nc
			}
		}

		scoreChange += mx - mn
	}

	return scoreChange
}
