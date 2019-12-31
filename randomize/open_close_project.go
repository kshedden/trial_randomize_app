package randomize

import (
	"fmt"
	"log"
	"net/http"
)

// OpenCloseProject is step 1 of changing the open/close status of a project.
func OpenCloseProject(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(pkey, r); !ok {
		msg := "You do not have access to this page."
		rmsg := "Return"
		messagePage(w, r, msg, rmsg, "/")
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Printf("OpenClose_project [1]: %v", err)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can open or close a project for enrollment."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		ProjectName string
		GroupNames  []string
		Open        bool
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		Pkey:        pkey,
		ProjectName: proj.Name,
		GroupNames:  proj.GroupNames,
		Open:        proj.Open,
	}

	if err := tmpl.ExecuteTemplate(w, "openclose_project.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// OpenCloseCompleted is the second step of changing the open/close status of a project.
func OpenCloseCompleted(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if ok := checkAccess(pkey, r); !ok {
		msg := "You do not have access to this page."
		rmsg := "Return"
		messagePage(w, r, msg, rmsg, "/")
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can open or close a project for enrollment."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	status := r.FormValue("open")

	if status == "open" {
		msg := fmt.Sprintf("The project \"%s\" is now open for enrollment.", proj.Name)
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		proj.Open = true
	} else {
		msg := fmt.Sprintf("The project \"%s\" is now closed for enrollment.", proj.Name)
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		proj.Open = false
	}

	err = storeProject(ctx, proj, pkey)
	if err != nil {
		msg := "Error, the project was not stored."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
	}
}
