package randomize

import (
	"log"
	"net/http"
)

// Dashboard displays a list of projects for the current user.
func Dashboard(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		log.Printf("Dashboard method != GET")
		Serve404(w)
		return
	}

	ctx := r.Context()
	user := userEmail(r)
	log.Printf("Dashboard email=%s", user)

	projlist, err := getProjects(ctx, user, true)
	if err != nil {
		log.Printf("Dashboard: %v", err)
		msg := "A database error occurred, projects cannot be retrieved."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}
	log.Printf("Got %d projects for %s", len(projlist), user)

	tvals := struct {
		User        string
		LoggedIn    bool
		AnyProjects bool
		Projects    []*ProjectView
	}{
		User:        user,
		LoggedIn:    user != "",
		AnyProjects: len(projlist) > 0,
		Projects:    formatProjects(projlist),
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.html", tvals); err != nil {
		log.Printf("Dashboard failed to execute template: %v", err)
	}
}
