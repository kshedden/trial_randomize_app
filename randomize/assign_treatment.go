package randomize

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

// AssignTreatmentInput is step 1 of assigning a treatment group to a subject.
func AssignTreatmentInput(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	pkey := r.FormValue("pkey")

	if !checkAccess(pkey, r) {
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		log.Printf("Assign_treatment_input: %v", err)
		msg := "A database error occurred, the project could not be loaded."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}

	if !proj.Open {
		msg := "This project is currently not open for new enrollments.  The project owner can change this by following the \"Open/close enrollment\" link on the project dashboard."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return
	}

	fproj := formatProject(proj)

	tvals := struct {
		User      string
		LoggedIn  bool
		PR        *Project
		PV        *ProjectView
		NumGroups int
		Fields    string
		Pkey      string
	}{
		User:      useremail,
		LoggedIn:  useremail != "",
		PR:        proj,
		PV:        fproj,
		NumGroups: len(proj.GroupNames),
		Pkey:      pkey,
	}

	S := make([]string, len(proj.Variables))
	for i, v := range proj.Variables {
		S[i] = v.Name
	}
	tvals.Fields = strings.Join(S, ",")

	if err := tmpl.ExecuteTemplate(w, "assign_treatment_input.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

func checkBeforeAssigning(proj *Project, pkey string, subjectId string, w http.ResponseWriter, r *http.Request) bool {

	if !proj.Open {
		msg := "This project is currently not open for new enrollments.  The project owner can change this by following the \"Open/close enrollment\" link on the project dashboard."
		rmsg := "Return to project"
		messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
		return false
	}

	// Check the subject id
	if proj.StoreRawData {

		if len(subjectId) == 0 {
			msg := fmt.Sprintf("The subject id may not be blank.")
			rmsg := "Return to project"
			messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
			return false
		}

		for _, rec := range proj.RawData {
			if subjectId == rec.SubjectId {
				msg := fmt.Sprintf("Subject '%s' has already been assigned to a treatment group.  Please use a different subject id.", subjectId)
				rmsg := "Return to project"
				messagePage(w, r, msg, rmsg, "/project_dashboard?pkey="+pkey)
				return false
			}
		}
	}

	return true
}

// AssignTreatmentConfirm is step 2 of assigning a treatment group to a subject.
func AssignTreatmentConfirm(w http.ResponseWriter, r *http.Request) {

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

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	subjectId := r.FormValue("subject_id")
	subjectId = strings.TrimSpace(subjectId)

	project, err := getProjectFromKey(pkey)
	if err != nil {
		log.Printf("Assign_treatment_confirm: %v", err)
		msg := "A database error occurred, the project could not be loaded."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}

	ok := checkBeforeAssigning(project, pkey, subjectId, w, r)
	if !ok {
		return
	}

	projView := formatProject(project)

	Fields := strings.Split(r.FormValue("fields"), ",")
	FV := make([][]string, len(Fields)+1)
	Values := make([]string, len(Fields))

	FV[0] = []string{"Subject id", subjectId}
	for i, v := range Fields {
		x := r.FormValue(v)
		FV[i+1] = []string{v, x}
		Values[i] = x
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		Pkey        string
		Project     *Project
		ProjectView *ProjectView
		NumGroups   int
		Fields      string
		FV          [][]string
		Values      string
		SubjectId   string
		AnyVars     bool
	}{
		User:        useremail,
		LoggedIn:    useremail != "",
		Pkey:        pkey,
		Project:     project,
		ProjectView: projView,
		NumGroups:   len(project.GroupNames),
		Fields:      strings.Join(Fields, ","),
		FV:          FV,
		Values:      strings.Join(Values, ","),
		SubjectId:   subjectId,
		AnyVars:     len(project.Variables) > 0,
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment_confirm.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}

// AssignTreatment is step 3 of assigning a treatment group to a subject.
func AssignTreatment(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)
	pkey := r.FormValue("pkey")

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	if !checkAccess(pkey, r) {
		return
	}

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	proj, err := getProjectFromKey(pkey)
	if err != nil {
		log.Printf("Assign_treatment %v", err)
		msg := "A database error occurred, the project could not be loaded."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}
	log.Printf("pkey=%v", pkey)
	log.Printf("proj=%+v\n", proj)

	subjectId := r.FormValue("subject_id")

	// Check this a second time in case someone lands on this page
	// without going through the previous checks
	// (e.g. inappropriate use of back button on browser).
	ok := checkBeforeAssigning(proj, pkey, subjectId, w, r)
	if !ok {
		return
	}

	pview := formatProject(proj)

	fields := strings.Split(r.FormValue("fields"), ",")
	values := strings.Split(r.FormValue("values"), ",")

	// mpv maps variable names to values for the unit that is about
	// to be randomized to a treatment group.
	mpv := make(map[string]string)
	for i, x := range fields {
		mpv[x] = values[i]
	}

	ax, err := proj.doAssignment(mpv, subjectId, useremail)
	if err != nil {
		log.Printf("%v", err)
	}

	proj.Modified = time.Now()

	// Update the project in the database.
	if _, err := client.Doc("Project/"+pkey).Set(ctx, proj); err != nil {
		log.Printf("Assign_treatment: %v", err)
		msg := "A database error occurred, the project could not be updated."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}

	tvals := struct {
		User      string
		LoggedIn  bool
		Project   *Project
		ProjView  *ProjectView
		NumGroups int
		Ax        string
		Pkey      string
	}{
		User:      useremail,
		LoggedIn:  useremail != "",
		Ax:        ax,
		Project:   proj,
		ProjView:  pview,
		NumGroups: len(proj.GroupNames),
		Pkey:      pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "assign_treatment.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}
