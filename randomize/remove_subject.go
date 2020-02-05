package randomize

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
)

// RemoveSubject is the first step for removing a subject from a project.
func RemoveSubject(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")
	ctx := r.Context()
	susers, _ := getSharedUsers(ctx, pkey)

	if !checkAccess(pkey, susers, r) {
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		msg := "Database error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can change treatment group assignments."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Subjects cannot be removed for a project in which the subject-level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	tvals := struct {
		User               string
		LoggedIn           bool
		Pkey               string
		ProjectName        string
		AnyRemovedSubjects bool
		RemovedSubjects    string
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		Pkey:        pkey,
		ProjectName: proj.Name,
	}

	if len(proj.RemovedSubjects) > 0 {
		tvals.AnyRemovedSubjects = true
		tvals.RemovedSubjects = strings.Join(proj.RemovedSubjects, ", ")
	}

	if err := tmpl.ExecuteTemplate(w, "remove_subject.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// RemoveSubjectConfirm is the second step for removing a subject from a project.
func RemoveSubjectConfirm(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")
	subjectId := r.FormValue("subject_id")

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can remove treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	// Check if the subject has already been removed
	for _, s := range proj.RemovedSubjects {
		if s == subjectId {
			msg := fmt.Sprintf("Subject '%s' has already been removed from the study.", subjectId)
			rmsg := "Return to project"
			messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
			return
		}
	}

	// Check if the subject exists
	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			found = true
			break
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in the project.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		SubjectId   string
		ProjectName string
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		SubjectId:   subjectId,
		Pkey:        pkey,
		ProjectName: proj.Name,
	}

	if err := tmpl.ExecuteTemplate(w, "remove_subject_confirm.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// RemoveSubjectCompleted is the third step for removing a subject from a project.
func RemoveSubjectCompleted(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")
	ctx := context.Background()
	susers, _ := getSharedUsers(ctx, pkey)

	if !checkAccess(pkey, susers, r) {
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
		msg := "Only the project owner can remove treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Subjects cannot be removed for a project in which the subject level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	subjectId := r.FormValue("subject_id")
	found := false
	var removeRec *DataRecord
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			rec.Included = false
			removeRec = rec
			found = true
			break
		}
	}

	if !found {
		msg := fmt.Sprintf("Unable to remove subject '%s' from the project.", subjectId)
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	proj.RemovedSubjects = append(proj.RemovedSubjects, subjectId)

	comment := &Comment{
		Commenter: useremail,
		DateTime:  time.Now(),
		Comment:   []string{fmt.Sprintf("Subject '%s' removed from the project.", subjectId)},
	}
	proj.Comments = append(proj.Comments, comment)

	removeFromAggregate(removeRec, proj)

	if err := storeProject(ctx, proj, pkey); err != nil {
		msg := "Error, unable to save project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	msg := fmt.Sprintf("Subject '%s' has been removed from the study.", subjectId)
	rmsg := "Return to project dashboard"
	messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
