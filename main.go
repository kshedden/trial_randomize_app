package main

import (
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/kshedden/trial_randomize_app/randomize"
)

func init() {

	log.Printf("running init function")

	http.HandleFunc("/", randomize.InformationPage)
	http.HandleFunc("/dashboard", randomize.Dashboard)

	// Project creation pages
	http.HandleFunc("/create_project_step1", randomize.CreateProjectStep1)
	http.HandleFunc("/create_project_step2", randomize.CreateProjectStep2)
	http.HandleFunc("/create_project_step3", randomize.CreateProjectStep3)
	http.HandleFunc("/create_project_step4", randomize.CreateProjectStep4)
	http.HandleFunc("/create_project_step5", randomize.CreateProjectStep5)
	http.HandleFunc("/create_project_step6", randomize.CreateProjectStep6)
	http.HandleFunc("/create_project_step7", randomize.CreateProjectStep7)
	http.HandleFunc("/create_project_step8", randomize.CreateProjectStep8)
	http.HandleFunc("/create_project_step9", randomize.CreateProjectStep9)

	// Copy project pages
	http.HandleFunc("/copy_project", randomize.CopyProject)
	http.HandleFunc("/copy_project_completed", randomize.CopyProjectCompleted)

	// Import/export pages
	http.HandleFunc("/export_project", randomize.ExportProject)
	http.HandleFunc("/import_project_step1", randomize.ImportProjectStep1)
	http.HandleFunc("/import_project_step2", randomize.ImportProjectStep2)

	// Project deletion pages
	http.HandleFunc("/delete_project_step1", randomize.DeleteProjectStep1)
	http.HandleFunc("/delete_project_step2", randomize.DeleteProjectStep2)
	http.HandleFunc("/delete_project_step3", randomize.DeleteProjectStep3)

	http.HandleFunc("/project_dashboard", randomize.ProjectDashboard)
	http.HandleFunc("/edit_sharing", randomize.EditSharing)
	http.HandleFunc("/edit_sharing_confirm", randomize.EditSharingConfirm)

	// Treatment assignment pages
	http.HandleFunc("/assign_treatment_input", randomize.AssignTreatmentInput)
	http.HandleFunc("/assign_treatment_confirm", randomize.AssignTreatmentConfirm)
	http.HandleFunc("/assign_treatment", randomize.AssignTreatment)

	http.HandleFunc("/view_statistics", randomize.ViewStatistics)
	http.HandleFunc("/view_comments", randomize.ViewComments)
	http.HandleFunc("/add_comment", randomize.AddComment)
	http.HandleFunc("/confirm_add_comment", randomize.ConfirmAddComment)
	http.HandleFunc("/view_complete_data", randomize.ViewCompleteData)

	// Remove subject pages
	http.HandleFunc("/remove_subject", randomize.RemoveSubject)
	http.HandleFunc("/remove_subject_confirm", randomize.RemoveSubjectConfirm)
	http.HandleFunc("/remove_subject_completed", randomize.RemoveSubjectCompleted)

	// Edit assignment pages
	http.HandleFunc("/edit_assignment", randomize.EditAssignment)
	http.HandleFunc("/edit_assignment_confirm", randomize.EditAssignmentConfirm)
	http.HandleFunc("/edit_assignment_completed", randomize.EditAssignmentCompleted)

	// Close or open project for enrollment pages
	http.HandleFunc("/openclose_project", randomize.OpenCloseProject)
	http.HandleFunc("/openclose_completed", randomize.OpenCloseCompleted)
}

/*
// checkAccess determines whether the given user has permission to
// access the given project.
func checkAccess(ctx context.Context, user *user.User, pkey string, w *http.ResponseWriter, r *http.Request) bool {

	userName := strings.ToLower(user.String())

	keyparts := strings.Split(pkey, "::")
	owner := keyparts[0]

	// A user can always access his or her own projects.
	if userName == strings.ToLower(owner) {
		return true
	}

	// Otherwise, check if the project is shared with the user.
	key := datastore.NewKey(ctx, "SharingByUser", userName, 0, nil)
	var sbuser SharingByUser
	err := datastore.Get(ctx, key, &sbuser)
	if err == datastore.ErrNoSuchEntity {
		checkAccessFailed(ctx, nil, w, r, user)
		return false
	} else if err != nil {
		checkAccessFailed(ctx, &err, w, r, user)
		return false
	}
	L := cleanSplit(sbuser.Projects, ",")
	for _, x := range L {
		if pkey == x {
			return true
		}
	}
	checkAccessFailed(ctx, nil, w, r, user)
	return false
}

// checkAccessFailed displays an error message when a project cannot be accessed.
func checkAccessFailed(ctx context.Context, err *error, w *http.ResponseWriter, r *http.Request, user *user.User) {

	if err != nil {
		msg := "A datastore error occurred.  Ask the administrator to check the log for error details."
		log.Errorf(ctx, "check_access_failed: %v", err)
		rmsg := "Return to dashboard"
		messagePage(*w, r, user, msg, rmsg, "/dashboard")
		return
	}
	msg := "You don't have access to this project."
	rmsg := "Return to dashboard"
	log.Infof(ctx, fmt.Sprintf("Failed access: %v\n", user))
	messagePage(*w, r, user, msg, rmsg, "/dashboard")
}
*/

type handler func(http.ResponseWriter, *http.Request)

func main() {

	tmpl := template.Must(template.ParseGlob("html_templates/*.html"))
	randomize.SetTemplates(tmpl)

	port := os.Getenv("PORT")
	if port == "" {
		log.Printf("Defaulting to port %s", port)
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
