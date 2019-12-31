package randomize

import (
	"log"
	"net/http"
	"strings"
)

// ProjectDashboard gets the project name from the user.
func ProjectDashboard(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(pkey, r); !ok {
		return
	}

	splkey := strings.Split(pkey, "::")
	owner := splkey[0]

	proj, _ := getProjectFromKey(pkey)
	projView := formatProject(proj)

	susers, _ := getSharedUsers(ctx, pkey)

	var sul []string
	for k := range susers {
		sul = append(sul, k)
	}

	tvals := struct {
		User            string
		LoggedIn        bool
		ProjView        *ProjectView
		NumGroups       int
		Sharing         string
		SharedUsers     []string
		Pkey            string
		ShowEditSharing bool
		Owner           string
		StoreRawData    string
		Open            string
		AnyVars         bool
	}{
		User:            useremail,
		LoggedIn:        useremail != "",
		ProjView:        projView,
		NumGroups:       len(proj.GroupNames),
		AnyVars:         len(proj.Variables) > 0,
		Pkey:            pkey,
		Sharing:         "Nobody",
		SharedUsers:     sul,
		ShowEditSharing: owner == useremail,
		Owner:           owner,
		StoreRawData:    boolYesNo(proj.StoreRawData),
		Open:            boolYesNo(projView.Open),
	}

	if len(susers) > 0 {
		tvals.Sharing = strings.Join(sul, ", ")
	}

	if err := tmpl.ExecuteTemplate(w, "project_dashboard.html", tvals); err != nil {
		log.Printf("projectDashbord failed to execute template: %v", err)
	}
}
