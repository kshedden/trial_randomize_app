package randomize

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cloud.google.com/go/firestore"
)

const (
	projectID = "trial-randomize"
)

var (
	tmpl *template.Template
)

// SetTemplates allows the templates to be set from the main function of the
// web application.
func SetTemplates(t *template.Template) {
	tmpl = t
}

// userEmail returns the email of the current user.
// TODO This should be made more robust by checking the certificate, see:
//   https://cloud.google.com/go/getting-started/authenticate-users-with-iap
func userEmail(r *http.Request) string {

	assertion := r.Header.Get("X-Goog-IAP-JWT-Assertion")
	if assertion == "" {
		log.Printf("No Cloud IAP header found.")
		return ""
	}

	e := r.Header.Get("X-Goog-Authenticated-User-Email")
	e = strings.Replace(e, "accounts.google.com:", "", 1)

	return e
}

// checkAccess returns true if and only if the currently
// logged in user has access to the project with the given key.
func checkAccess(key string, r *http.Request) bool {
	email := userEmail(r)
	toks := splitKey(key)
	return toks[0] == email
}

// DataRecord stores one record of raw data.
type DataRecord struct {

	// SubjectId identifies the subject
	SubjectId string

	// AssignedTime is the time at which the treatment group assignment was made
	AssignedTime time.Time

	// AssignedGroup is the group label that the subject is assigned to
	AssignedGroup string

	// CurrentGroup
	CurrentGroup string

	// Includes is true if the subject has not been removed from the study
	Included bool

	// Data is the raw data for this subject
	Data []string

	// Assigner is the id of the person who last assigned this person to a group
	Assigner string
}

// Project stores all information about one project.
type Project struct {

	// Owner is the Google id of the project owner
	Owner string

	// The date and time that the project was created
	Created time.Time

	// Name contains the name of the project as selected by its creator
	Name string

	// The names of the groups
	GroupNames []string

	// Variables is a slice containing all variables being balanced
	Variables []Variable

	// Assignments contains the number of subjects currently
	// assigned to each group.
	Assignments []int

	// CellTotals contains the total cell counts
	CellTotals []float64

	// Bias controls the level of determinism in the group assignments
	Bias int

	// Comments is a list of comments for the subject.
	Comments []*Comment

	// The date and time of the last assignment
	Modified time.Time

	// If true, store the individual-level data, otherwise only store aggregates
	StoreRawData bool

	// The individual-level data, if stored
	RawData []*DataRecord

	// RemovedSubjects is a slice containing the ids of all subjects
	// who have been removed from the study
	RemovedSubjects []string

	// Open is true if the project is currently open for enrollment
	Open bool

	// SamplingRates contains the sampling rates for each treatment group.
	// The default sampleing rates are 1 for each group.
	SamplingRates []float64
}

// NumAssignments returns the total number of current treatment group assignments.
func (proj *Project) NumAssignments() int {
	t := 0
	for _, n := range proj.Assignments {
		t += n
	}
	return t
}

// GetData returns the number of enrolled subjects with the given
// level of the given variable that are assigned to the given group.
func (proj *Project) GetData(variable, level, group int) float64 {

	p := len(proj.Variables)
	q := len(proj.GroupNames)
	r := len(proj.CellTotals) / (p * q)

	ii := q*r*variable + q*level + group
	return proj.CellTotals[ii]
}

// SetData sets the number of enrolled subjects with the given level
// of the given variable that are assigned to the given group.
func (proj *Project) SetData(variable, level, group int, x float64) {

	p := len(proj.Variables)
	q := len(proj.GroupNames)
	r := len(proj.CellTotals) / (p * q)

	ii := q*r*variable + q*level + group
	proj.CellTotals[ii] = x
}

// ProjectView is a printable version of Project.
type ProjectView struct {

	// Owner identifies the owner of the project
	Owner string

	// CreatedDate is the printable form of the date when the project was created
	CreatedDate string

	// CreatedTime is the printable form of the time when the project was created
	CreatedTime string

	// GroupNames contains the names of the treatment groups
	GroupNames string

	// Name is the name of the project
	Name string

	// Variables contains printable versions of the variables
	Variables []VariableView

	// Key is the database key used to store the project
	Key string

	// Bias is a printable form of the bias/determinism constant for the project
	Bias string

	// Comments contains all project comments
	Comments []*Comment

	// ModifiedDate is a printable version of the date when the project was last modified
	ModifiedDate string

	// ModifiedTime is a printable version of the time when the project was last modified
	ModifiedTime string

	// StoreRawData indicates whether the raw data are stored for this project
	StoreRawData bool

	// RemovedSubjects contains the names of the subjects who have been removed from the study
	RemovedSubjects []string

	// Open is true if the project is currently open for assignments
	Open bool

	// SamplingRates is a printable version of the project sampling rates
	SamplingRates string

	// The project that this view was derived from
	Project *Project
}

// Variable contains information about one variable that will be used
// as part of the treatment assignment.
type Variable struct {

	// Name identifies the variable
	Name string

	// Levels are the distinct values that the variable can have
	Levels []string

	// Weight is a numeric weight for this variable
	Weight float64
}

// VariableView is a printable version of a variable.
type VariableView struct {
	Name   string
	Levels string
	Index  int
	Weight string
}

// Comment stores a single comment.
type Comment struct {

	// Commenter identifies the person who made the comment
	Commenter string

	// DateTime records when the comment was made
	DateTime time.Time

	// Date is the printable form of the date when the comment was made
	Date string

	// Time is the printable form of the time when the comment was made
	Time string

	// Comment containst the comment, broken into lines of text
	Comment []string
}

// cleanSplit splits a string into tokens delimited by a given
// separator.  If S equals the empty string, this function returns an
// empty list, rather than a list containing an empty string as its
// sole element.  Leading and trailing whitespace is removed from each
// element of the returned list.
func cleanSplit(s string, sep string) []string {

	if len(s) == 0 {
		return []string{}
	}

	parts := strings.Split(s, sep)

	for i, v := range parts {
		parts[i] = strings.TrimSpace(v)
	}

	return parts
}

// getProjectfromKey
func getProjectFromKey(pkey string) (*Project, error) {

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	ds, err := client.Doc("Project/" + pkey).Get(ctx)
	if err != nil {
		return nil, err
	}

	var proj Project
	if err := ds.DataTo(&proj); err != nil {
		return nil, err
	}

	return &proj, nil
}

// storeProject
func storeProject(ctx context.Context, proj *Project, pkey string) error {

	ctx = context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.Doc("Project/"+pkey).Set(ctx, proj)

	return err
}

func makeKey(owner, name string) string {
	return owner + "::" + name
}

func splitKey(key string) []string {
	return strings.Split(key, "::")
}

// formatProject returns a ProjectView object corresponding to the
// given Project and Key object.
func formatProject(proj *Project) *ProjectView {

	t := proj.Created
	loc, _ := time.LoadLocation("America/New_York")
	t = t.In(loc)

	fp := ProjectView{
		Owner:           proj.Owner,
		Name:            proj.Name,
		Comments:        proj.Comments,
		Bias:            fmt.Sprintf("%d", proj.Bias),
		CreatedDate:     t.Format("2006-1-2"),
		CreatedTime:     t.Format("3:04pm"),
		GroupNames:      strings.Join(proj.GroupNames, ","),
		Variables:       make([]VariableView, len(proj.Variables)),
		RemovedSubjects: proj.RemovedSubjects,
		Open:            proj.Open,
		Key:             makeKey(proj.Owner, proj.Name),
		Project:         proj,
	}

	rateStr := make([]string, len(proj.SamplingRates))
	for i, x := range proj.SamplingRates {
		rateStr[i] = fmt.Sprintf("%.0f", x)
	}
	fp.SamplingRates = strings.Join(rateStr, ",")

	for i, pv := range proj.Variables {
		fp.Variables[i] = formatVariable(pv)
	}

	t = proj.Modified
	if !t.IsZero() {
		loc, _ = time.LoadLocation("America/New_York")
		t = t.In(loc)
		fp.ModifiedDate = t.Format("2006-1-2")
		fp.ModifiedTime = t.Format("3:04pm")
	}

	return &fp
}

// formatProjects...
func formatProjects(projects []*Project) []*ProjectView {

	var fp []*ProjectView

	for _, proj := range projects {
		fp = append(fp, formatProject(proj))
	}

	return fp
}

// boolYesNo takes a bool and returns a string "Yes" (if true) or "No" (if false),
// accordingly.
func boolYesNo(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// formatVariable returns a VariableView object corresponding to the
// given Variable object.
func formatVariable(va Variable) VariableView {

	return VariableView{
		Name:   va.Name,
		Levels: strings.Join(va.Levels, ","),
		Weight: fmt.Sprintf("%.0f", va.Weight),
	}
}

// getSharedUsers returns the user id's for for users who are
// shared for the given project.
func getSharedUsers(ctx context.Context, projectName string) (map[string]bool, error) {

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	doc, err := client.Doc("SharingByProject/" + projectName).Get(ctx)
	if err != nil {
		return nil, err
	}

	sbp := make(map[string]bool)
	if doc.Exists() {
		if err := doc.DataTo(&sbp); err != nil {
			return nil, err
		}
	}

	return sbp, nil
}

// addSharing adds all the given users to be shared for the given
// project.
func addSharing(projectName string, userNames []string) error {

	ctx := context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return err
	}
	defer client.Close()

	if len(userNames) == 0 {
		return nil
	}

	sbp := make(map[string]bool)
	doc, err := client.Doc("SharingByProject/" + projectName).Get(ctx)
	if status.Code(err) == codes.NotFound {
		// OK
	} else if err != nil {
		log.Printf("addSharing [1]: %v", err)
		return err
	} else {
		if err := doc.DataTo(&sbp); err != nil {
			log.Printf("addSharing [2]: %v", err)
			return err
		}
	}

	for _, u := range userNames {
		sbp[u] = true
	}

	// Store the update sharing by project information
	if _, err = client.Doc("SharingByProject/"+projectName).Set(ctx, &sbp); err != nil {
		log.Printf("addSharing [3]: %v", err)
		return err
	}

	// Update SharingByUser
	for _, uname := range userNames {

		sbu := make(map[string]bool)
		na := "SharingByUser/" + strings.ToLower(uname)
		doc, err := client.Doc(na).Get(ctx)
		if status.Code(err) == codes.NotFound {
			// OK
		} else if err != nil {
			log.Printf("addSharing [4]: %v", err)
			return err
		} else {
			if err := doc.DataTo(&sbu); err != nil {
				log.Printf("addSharing [5]: %v", err)
				return err
			}
		}

		sbu[projectName] = true

		if _, err := client.Doc(na).Set(ctx, &sbu); err != nil {
			log.Printf("addSharing [6]: %v", err)
			return err
		}
	}

	return nil
}

// removeSharing removes the given users from the access list for the given project.
func removeSharing(ctx context.Context, projectName string, userNames []string) error {

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	// Update SharingByProject.
	sbp := make(map[string]string)
	doc, err := client.Doc("SharingByProject/" + projectName).Get(ctx)
	if status.Code(err) == codes.NotFound {
		// OK
	} else if err != nil {
		return err
	} else {
		if err := doc.DataTo(&sbp); err != nil {
			return err
		}
		for _, u := range userNames {
			delete(sbp, u)
		}
	}

	if _, err = client.Doc("SharingByProject/"+projectName).Set(ctx, sbp); err != nil {
		return err
	}

	// Update SharingByUser
	for _, name := range userNames {

		sbu := make(map[string]bool)
		na := "SharingByUser/" + strings.ToLower(name)
		doc, err := client.Doc(na).Get(ctx)
		if status.Code(err) == codes.NotFound {
			// OK
		} else if err != nil {
			return err
		} else {
			if err := doc.DataTo(&sbu); err != nil {
				return err
			}
			delete(sbu, projectName)
		}

		if _, err := client.Doc(na).Set(ctx, &sbu); err != nil {
			return err
		}
	}

	return nil
}

// getProjects returns all projects owned by the given user.
// Optionally also include projects that are shared with the user.
func getProjects(ctx context.Context, user string, includeShared bool) ([]*Project, error) {

	user = strings.ToLower(user)

	ctx = context.Background()
	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Printf("GetProjects[1]: %v", err)
		return nil, err
	}
	defer client.Close()

	docs := client.Collection("Project")
	q := docs.Where("Owner", "==", user).OrderBy("Created", firestore.Desc).Limit(100).Documents(ctx)

	var projlist []*Project
	adocs, err := q.GetAll()
	if err != nil {
		log.Printf("GetProjects[2]: %v", err)
		return nil, err
	}
	log.Printf("Got %d projects", len(adocs))

	for _, doc := range adocs {
		var proj Project
		if err := doc.DataTo(&proj); err != nil {
			log.Printf("GetProjects[3]: %v", err)
			return nil, err
		}
		projlist = append(projlist, &proj)
	}

	if !includeShared {
		return projlist, nil
	}

	// Get project ids that are shared with this user
	var sbu map[string]string
	doc, err := client.Doc("SharingByUser/" + user).Get(ctx)
	if status.Code(err) == codes.NotFound {
		// No projects shared with this user
		return projlist, nil
	} else if err != nil {
		log.Printf("GetProjects[2]: %v", err)
		return nil, err
	}

	if err := doc.DataTo(&sbu); err != nil {
		log.Printf("GetProjects[3]: %v", err)
		return nil, err
	}

	// Get the shared projects
	for _, spv := range sbu {

		doc, err := client.Doc("Project/" + spv).Get(ctx)
		if status.Code(err) == codes.NotFound {
			log.Printf("getProjects[4]: %v", err)
			continue
		} else if err != nil {
			log.Printf("getProjects[5]: %v", err)
			return nil, err
		}

		var proj Project
		if err := doc.DataTo(&proj); err != nil {
			log.Printf("getProjects [6]: %v\n%v", spv, err)
			return nil, err
		}

		projlist = append(projlist, &proj)
	}

	return projlist, nil
}

// Serve404 is used when the GET/POST method is mismatched to the handler.
func Serve404(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "Not Found")
}

// ServeError displays an error page.  We should avoid using this when possible.
func ServeError(ctx context.Context, w http.ResponseWriter, err error) {

	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = io.WriteString(w, "Internal Server Error")
	log.Printf("ServeError [1]: %v", err)
}

// messagePage presents a simple message page and presents the user
// with a link that leads to a followup page.
func messagePage(w http.ResponseWriter, r *http.Request, msg string, rmsg string, returnURL string) {

	useremail := userEmail(r)

	tvals := struct {
		User      string
		LoggedIn  bool
		Msg       string
		ReturnURL string
		ReturnMsg string
	}{
		User:      useremail,
		LoggedIn:  useremail != "",
		Msg:       msg,
		ReturnURL: returnURL,
		ReturnMsg: rmsg,
	}

	if err := tmpl.ExecuteTemplate(w, "message.html", tvals); err != nil {
		log.Printf("messagePage failed to execute template: %v", err)
	}
}

// getIndex returns the position of `val` within `vec`.
func getIndex(vec []string, val string) int {

	for i, x := range vec {
		if val == x {
			return i
		}
	}
	return -1
}

// removeFromAggregate updates the aggregate statistics (count per
// treatment arm for each level of each variable) for the given data
// record.
func removeFromAggregate(rec *DataRecord, proj *Project) {

	grpIx := getIndex(proj.GroupNames, rec.CurrentGroup)

	// Update the overall assignment totals
	proj.Assignments[grpIx]--

	// Update the within-variable assignment totals
	for j, va := range proj.Variables {
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				x := proj.GetData(j, k, grpIx)
				proj.SetData(j, k, grpIx, x-1)
			}
		}
	}
}

// addToAggregate updates the aggregate statistics (count per
// treatment arm for each level of each variable) for the given data
// record.
func addToAggregate(rec *DataRecord,
	proj *Project) {

	grpIx := getIndex(proj.GroupNames, rec.CurrentGroup)

	// Update the overall assignment totals
	proj.Assignments[grpIx]++

	// Update the within-variable assignment totals
	for j, va := range proj.Variables {
		for k, lev := range va.Levels {
			if rec.Data[j] == lev {
				x := proj.GetData(j, k, grpIx)
				proj.SetData(j, k, grpIx, x+1)
			}
		}
	}
}

// PrintData prints the cell totals to standard output.
func (proj *Project) PrintData() {

	ngrp := len(proj.GroupNames)

	for k, va := range proj.Variables {
		fmt.Printf("%s\n", va.Name)
		for j := range va.Levels {
			for g := 0; g < ngrp; g++ {
				fmt.Printf("%4.0f", proj.GetData(k, j, g))
			}
			fmt.Printf("\n")
		}
		fmt.Printf("\n")
	}
}
