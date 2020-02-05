package randomize

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"
)

// EditAssignment is step 1 of the process of editing (changing) a treatment assignment.
func EditAssignment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")
	ctx := context.Background()
	susers, _ := getSharedUsers(ctx, pkey)

	if !checkAccess(pkey, susers, r) {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		msg := "Datastore error: unable to retrieve project."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Printf("Edit_assignment_confirm [1]: %v", err)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if proj.NumAssignments() == 0 {
		msg := "There are no assignments to edit."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Group assignments cannot be edited for a project in which the subject level data is not stored"
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
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		Pkey:        pkey,
		ProjectName: proj.Name,
		GroupNames:  proj.GroupNames,
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment.html", tvals); err != nil {
		log.Printf("editAssignment failed to execute template: %v", err)
	}
}

// EditAssignmentConfirm is step 2 of the process of editing (changing) a treatment assignment.
func EditAssignmentConfirm(w http.ResponseWriter, r *http.Request) {

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
		log.Printf("Edit_assignment_confirm [1]: %v", err)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Assignments cannot be edited in a project in which the subject level data is not stored"
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	subjectId := r.FormValue("SubjectId")

	tvals := struct {
		User             string
		LoggedIn         bool
		Pkey             string
		ProjectName      string
		CurrentGroupName string
		NewGroupName     string
		SubjectId        string
	}{
		User:         useremail,
		LoggedIn:     useremail != "",
		Pkey:         pkey,
		ProjectName:  proj.Name,
		NewGroupName: r.FormValue("NewGroupName"),
		SubjectId:    subjectId,
	}

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			tvals.CurrentGroupName = rec.CurrentGroup
			found = true
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if tvals.CurrentGroupName == tvals.NewGroupName {
		msg := fmt.Sprintf("You have requested to change the treatment group of subject '%s' to '%s', but the subject is already in this treatment group.", subjectId, tvals.NewGroupName)
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if err := tmpl.ExecuteTemplate(w, "edit_assignment_confirm.html", tvals); err != nil {
		log.Printf("editAssignmentConfirm failed to execute template: %v", err)
	}
}

// EditAssignmentCompleted is step 3 of the process of editing (changing) a treatment assignment.
func EditAssignmentCompleted(w http.ResponseWriter, r *http.Request) {

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
		log.Printf("Edit_assignment_completed [1]: %v", err)
		return
	}

	if proj.Owner != useremail {
		msg := "Only the project owner can edit treatment group assignments that have already been made."
		rmsg := "Return to project dashboard"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	if !proj.StoreRawData {
		msg := "Group assignments cannot be edited in a project in which the subject level data is not stored."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	newGroupName := r.FormValue("new_group_name")
	subjectId := r.FormValue("subject_id")

	found := false
	for _, rec := range proj.RawData {
		if rec.SubjectId == subjectId {
			removeFromAggregate(rec, proj)
			oldGroupName := rec.CurrentGroup
			rec.CurrentGroup = newGroupName
			addToAggregate(rec, proj)

			comment := &Comment{
				Commenter: useremail,
				DateTime:  time.Now(),
				Comment: []string{
					fmt.Sprintf("Group assignment for subject '%s' changed from '%s' to '%s'",
						subjectId, oldGroupName, newGroupName)},
			}
			proj.Comments = append(proj.Comments, comment)

			found = true
		}
	}
	if !found {
		msg := fmt.Sprintf("There is no subject with id '%s' in this project, the assignment was not changed.", subjectId)
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	err = storeProject(ctx, proj, pkey)
	if err != nil {
		msg := "Database error, your project was not saved."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	msg := "The assignment has been changed."
	rmsg := "Return to project"
	messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
