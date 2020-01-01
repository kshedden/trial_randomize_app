package randomize

import (
	"log"
	"net/http"
	"strings"
	"time"
)

// ViewComments
func ViewComments(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if !checkAccess(pkey, r) {
		return
	}

	proj, _ := getProjectFromKey(pkey)
	projv := formatProject(proj)

	for _, c := range projv.Comments {
		c.Date = c.DateTime.Format("2006-1-2")
		c.Time = c.DateTime.Format("3:04pm")
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Proj        *Project
		ProjView    *ProjectView
		Pkey        string
		AnyComments bool
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		Proj:        proj,
		ProjView:    projv,
		AnyComments: len(proj.Comments) > 0,
		Pkey:        pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "view_comments.html", tvals); err != nil {
		log.Printf("ViewComments: %v", err)
	}
}

// AddComment
func AddComment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if !checkAccess(pkey, r) {
		return
	}

	proj, _ := getProjectFromKey(pkey)
	fproj := formatProject(proj)

	tvals := struct {
		User     string
		LoggedIn bool
		PR       *Project
		PV       *ProjectView
		Pkey     string
	}{
		User:     useremail,
		LoggedIn: useremail != "",
		PR:       proj,
		PV:       fproj,
		Pkey:     pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "add_comment.html", tvals); err != nil {
		log.Printf("addComment: %v", err)
	}
}

// ConfirmAddComment
func ConfirmAddComment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	if !checkAccess(pkey, r) {
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		msg := "Datastore error, unable to add comment."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		log.Printf("confirmAddComment [1]: %v", err)
		return
	}

	commentText := r.FormValue("comment_text")
	commentText = strings.TrimSpace(commentText)
	commentLines := strings.Split(commentText, "\n")

	if len(commentText) == 0 {
		msg := "No comment was entered."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	loc, _ := time.LoadLocation("America/New_York")
	t := time.Now().In(loc)
	comment := &Comment{
		Commenter: useremail,
		DateTime:  time.Now(),
		Date:      t.Format("2006-1-2"),
		Time:      t.Format("3:04pm"),
		Comment:   commentLines,
	}
	proj.Comments = append(proj.Comments, comment)

	err = storeProject(ctx, proj, pkey)
	if err != nil {
		msg := "Error, your project was not saved."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	msg := "Your comment has been added to the project."
	rmsg := "Return to project"
	messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
}
