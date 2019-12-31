package randomize

import (
	"fmt"
	"log"
	"net/http"
)

// ViewStatistics displays the current numbers of people assigned to each
// treatment arm, within each level of each variable.
func ViewStatistics(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(pkey, r); !ok {
		return
	}

	var err error
	proj, err := getProjectFromKey(pkey)
	if err != nil {
		log.Printf("View_statistics [1]: %v", err)
		msg := "Datastore error: unable to view statistics."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		log.Printf("View_statistics [1]: %v", err)
		return
	}
	projectView := formatProject(proj)

	// Treatment assignment.
	txAsgn := make([][]string, len(proj.GroupNames))
	for k, v := range proj.GroupNames {
		txAsgn[k] = []string{v, fmt.Sprintf("%d", proj.Assignments[k])}
	}

	numGroups := len(proj.GroupNames)

	m := 0
	for _, v := range proj.Variables {
		m += len(v.Levels)
	}

	// Balance statistics
	balStat := make([][]string, m)
	jj := 0
	for j, v := range proj.Variables {
		numLevels := len(v.Levels)
		for k := 0; k < numLevels; k++ {
			fstat := make([]string, 1+numGroups)
			fstat[0] = v.Name + "=" + v.Levels[k]
			for q := 0; q < numGroups; q++ {
				u := proj.GetData(j, k, q)
				fstat[q+1] = fmt.Sprintf("%.0f", u)
			}
			balStat[jj] = fstat
			jj++
		}
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Project     *Project
		AnyVars     bool
		ProjectView *ProjectView
		TxAsgn      [][]string
		BalStat     [][]string
		Pkey        string
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		Project:     proj,
		AnyVars:     len(proj.Variables) > 0,
		ProjectView: projectView,
		TxAsgn:      txAsgn,
		Pkey:        pkey,
		BalStat:     balStat,
	}

	if err := tmpl.ExecuteTemplate(w, "view_statistics.html", tvals); err != nil {
		log.Printf("viewStatistics failed to execute template: %v", err)
	}
}
