package randomize

import (
	"log"
	"net/http"
)

func Dashboard(w http.ResponseWriter, r *http.Request) {

	log.Printf("Dashboard")

	if r.Method != "GET" {
		log.Printf("Dashboard method != GET")
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)
	log.Printf("Dashboard email=%s", useremail)

	projlist, err := getProjects(ctx, useremail, true)
	if err != nil {
		msg := "A database error occurred, projects cannot be retrieved."
		log.Printf("Dashboard: %v", err)
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}
	log.Printf("Got %d projects", len(projlist))

	tvals := struct {
		User        string
		LoggedIn    bool
		AnyProjects bool
		Projects    []*ProjectView
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		AnyProjects: len(projlist) > 0,
		Projects:    formatProjects(projlist),
	}

	if err := tmpl.ExecuteTemplate(w, "dashboard.html", tvals); err != nil {
		log.Printf("Dashboard failed to execute template: %v", err)
	}
}
