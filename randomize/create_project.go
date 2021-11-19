package randomize

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CreateProjectStep1 gets the project name from the user.
func CreateProjectStep1(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)

	tvals := struct {
		User     string
		LoggedIn bool
	}{
		User:     useremail,
		LoggedIn: useremail != "",
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step1.html", tvals); err != nil {
		log.Printf("createProjectStep1 failed to execute template: %v", err)
	}
}

// CreateProjectStep2 asks if the subject-level data are to be logged.
func CreateProjectStep2(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)
	projectName := r.FormValue("project_name")

	// Check if the project name has already been used.
	pkey := useremail + "::" + projectName
	_, err = client.Doc("Project/" + pkey).Get(ctx)
	if status.Code(err) == codes.NotFound {
		// OK
	} else if err != nil {
		msg := "Database error"
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	} else if err == nil {
		msg := fmt.Sprintf("A project named '%s' belonging to user %s already exists.", projectName, useremail)
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}

	tvals := struct {
		User     string
		LoggedIn bool
		Name     string
		Pkey     string
	}{
		User:     useremail,
		LoggedIn: useremail != "",
		Name:     r.FormValue("project_name"),
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step2.html", tvals); err != nil {
		log.Printf("createProjectStep2 failed to execute template: %v", err)
	}
}

// CreateProjectStep3 gets the number of treatment groups.
func CreateProjectStep3(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)

	tvals := struct {
		User         string
		LoggedIn     bool
		Name         string
		Pkey         string
		StoreRawData bool
	}{
		User:         useremail,
		LoggedIn:     useremail != "",
		Name:         r.FormValue("project_name"),
		StoreRawData: r.FormValue("store_rawdata") == "yes",
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step3.html", tvals); err != nil {
		log.Printf("createProjectStep3 failed to execute template: %v", err)
	}
}

// CreateProjectStep4
func CreateProjectStep4(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))

	// Group numbers (they don't have names yet)
	groupix := make([]int, numgroups)
	for i := 0; i < numgroups; i++ {
		groupix[i] = i + 1
	}

	tvals := struct {
		User         string
		LoggedIn     bool
		Name         string
		Pkey         string
		StoreRawData bool
		NumGroups    int
		IX           []int
	}{
		User:         useremail,
		LoggedIn:     useremail != "",
		Name:         r.FormValue("project_name"),
		StoreRawData: r.FormValue("store_rawdata") == "true",
		IX:           groupix,
		NumGroups:    numgroups,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step4.html", tvals); err != nil {
		log.Printf("createProjectStep4 failed to execute template: %v", err)
	}
}

// CreateProjectStep5
func CreateProjectStep5(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))

	// Indices for the groups
	groupix := make([]int, numgroups)
	for i := 0; i < numgroups; i++ {
		groupix[i] = i
	}

	// Get the group names from the previous page
	GroupNames := make([]string, numgroups)
	for i := 0; i < numgroups; i++ {
		GroupNames[i] = r.FormValue(fmt.Sprintf("name%d", i+1))
	}

	tvals := struct {
		User           string
		LoggedIn       bool
		Name           string
		Pkey           string
		GroupNames     string
		GroupNames_arr []string
		StoreRawData   bool
		NumGroups      int
		IX             []int
	}{
		User:           useremail,
		LoggedIn:       useremail != "",
		Name:           r.FormValue("project_name"),
		GroupNames:     strings.Join(GroupNames, ","),
		GroupNames_arr: GroupNames,
		NumGroups:      len(GroupNames),
		StoreRawData:   r.FormValue("store_rawdata") == "true",
		IX:             groupix,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step5.html", tvals); err != nil {
		log.Printf("createProjectStep5 failed to execute template: %v", err)
	}
}

// CreateProjectStep6
func CreateProjectStep6(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)

	numgroups, err := strconv.Atoi(r.FormValue("numgroups"))
	if err != nil {
		msg := "Unable to parse project, the number of groups must be a number."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		log.Printf("createProjectStep6: %v", err)
	}

	// Get the sampling rates from the previous page
	groupNamesArr := cleanSplit(r.FormValue("group_names"), ",")
	samplingRates := make([]string, numgroups)
	for i := 0; i < numgroups; i++ {

		samplingRates[i] = r.FormValue(fmt.Sprintf("rate%s", groupNamesArr[i]))

		x, err := strconv.ParseFloat(samplingRates[i], 64)
		if err != nil {
			log.Printf("createProjectStep6: %v", err)
		}

		if err != nil || x <= 0 {
			msg := "The sampling rates must be positive numbers."
			rmsg := "Return to dashboard"
			messagePage(w, r, msg, rmsg, "/dashboard")
			return
		}
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		GroupNames    string
		StoreRawData  bool
		SamplingRates string
		NumGroups     int
	}{
		User:          useremail,
		LoggedIn:      useremail != "",
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: strings.Join(samplingRates, ","),
		NumGroups:     numgroups,
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step6.html", tvals); err != nil {
		log.Printf("createProjectStep6 failed to execute template: %v", err)
	}
}

// CreateProjectStep7
func CreateProjectStep7(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	if err := r.ParseForm(); err != nil {
		log.Printf("createProjectStep7: %v", err)
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)

	numgroups, _ := strconv.Atoi(r.FormValue("numgroups"))
	numvar, _ := strconv.Atoi(r.FormValue("numvar"))

	groupix := make([]int, numvar)
	for i := 0; i < numvar; i++ {
		groupix[i] = i + 1
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		IX            []int
		GroupNames    string
		StoreRawData  bool
		NumGroups     int
		NumVar        int
		AnyVars       bool
		SamplingRates string
	}{
		User:          useremail,
		LoggedIn:      useremail != "",
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		IX:            groupix,
		NumGroups:     numgroups,
		NumVar:        numvar,
		AnyVars:       numvar > 0,
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: r.FormValue("rates"),
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step7.html", tvals); err != nil {
		log.Printf("createProjectStep7 failed to execute template: %v", err)
	}
}

func processVariableInfo(r *http.Request, numvar int) (string, bool) {

	variables := make([]string, numvar)

	for i := 0; i < numvar; i++ {
		vec := make([]string, 4)

		vname := fmt.Sprintf("name%d", i+1)
		vec[0] = strings.TrimSpace(r.FormValue(vname))
		if len(vec[0]) == 0 {
			return "", false
		}

		vname = fmt.Sprintf("levels%d", i+1)
		vec[1] = r.FormValue(vname)
		levels := cleanSplit(vec[1], ",")
		if len(levels) < 2 {
			return "", false
		}
		for _, x := range levels {
			if len(x) == 0 {
				return "", false
			}
		}

		vec[2] = r.FormValue(fmt.Sprintf("weight%d", i+1))
		vec[3] = r.FormValue(fmt.Sprintf("func%d", i+1))
		variables[i] = strings.Join(vec, ";")
	}

	return strings.Join(variables, ":"), true
}

// CreateProjectStep8
func CreateProjectStep8(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()
	useremail := userEmail(r)

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	numgroups, err := strconv.Atoi(r.FormValue("numgroups"))
	if err != nil {
		log.Printf("createrProjectStep8: %v", err)
	}

	numvar, err := strconv.Atoi(r.FormValue("numvar"))
	if err != nil {
		log.Printf("createrProjectStep8: %v", err)
	}

	variables, ok := processVariableInfo(r, numvar)
	if !ok {
		validationErrorStep8(w, r)
		return
	}

	ix := make([]int, numvar)
	for i := 0; i < numvar; i++ {
		ix[i] = i + 1
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		Pkey          string
		IX            []int
		GroupNames    string
		StoreRawData  bool
		NumGroups     int
		Numvar        int
		Variables     string
		SamplingRates string
	}{
		User:          useremail,
		LoggedIn:      useremail != "",
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		IX:            ix,
		NumGroups:     numgroups,
		Numvar:        numvar,
		Variables:     variables,
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: r.FormValue("rates"),
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step8.html", tvals); err != nil {
		log.Printf("createProjectStep8 failed to execute template: %v", err)
	}
}

// CreateProjectStep9 creates the project using all supplied
// information, and stores the project in the database.
func CreateProjectStep9(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	ctx := r.Context()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		panic(err)
	}

	if err := r.ParseForm(); err != nil {
		ServeError(ctx, w, err)
		return
	}

	useremail := userEmail(r)

	numvar, err := strconv.Atoi(r.FormValue("numvar"))
	if err != nil {
		log.Printf("createProjectStep9: %v", err)
	}

	GroupNames := r.FormValue("group_names")
	projectName := r.FormValue("project_name")
	variables := r.FormValue("variables")
	VL := cleanSplit(variables, ":")

	bias, err := strconv.Atoi(r.FormValue("bias"))
	if err != nil {
		log.Printf("createProjectStep9: %v", err)
	}

	// Parse and validate the variable information.
	vars := make([]Variable, numvar)
	for i, vl := range VL {
		vx := cleanSplit(vl, ";")
		var va Variable
		va.Name = vx[0]
		va.Levels = cleanSplit(vx[1], ",")

		va.Weight, err = strconv.ParseFloat(vx[2], 64)
		if err != nil {
			log.Printf("createProjectStep9: %v", err)
		}

		vars[i] = va
	}

	gn := cleanSplit(GroupNames, ",")

	proj := Project{
		Owner:        useremail,
		Created:      time.Now(),
		Name:         projectName,
		Variables:    vars,
		Bias:         bias,
		GroupNames:   gn,
		Assignments:  make([]int, len(gn)),
		StoreRawData: r.FormValue("store_rawdata") == "true",
		Open:         true,
	}

	// Convert the rates to numbers
	rates := r.FormValue("rates")
	ratesArr := cleanSplit(rates, ",")
	ratesNum := make([]float64, len(ratesArr))
	for i, x := range ratesArr {
		ratesNum[i], err = strconv.ParseFloat(x, 64)
		if err != nil {
			log.Printf("createProjectStep9: %v", err)
		}
	}
	proj.SamplingRates = ratesNum

	// Set up the data.
	{
		// Maximum number of levels
		n := 1
		for _, va := range proj.Variables {
			if len(va.Levels) > n {
				n = len(va.Levels)
			}
		}

		m := n * len(proj.GroupNames) * len(proj.Variables)
		log.Printf("allocated data slice of length %d", m)
		proj.CellTotals = make([]float64, m)
	}

	pkey := makeKey(useremail, projectName)
	if _, err := client.Doc("Project/"+pkey).Set(ctx, proj); err != nil {
		msg := "A database error occurred, the project was not created."
		log.Printf("Create_project_step9: %v", err)
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}

	// Remove any stale SharingByProject entities so that this project starts out
	// with no sharing
	if _, err := client.Doc("SharingByProject/" + pkey).Delete(ctx); err != nil {
		log.Printf("Create_project_step9 [3]: %v", err)
	}

	tvals := struct {
		User     string
		LoggedIn bool
	}{
		User:     useremail,
		LoggedIn: useremail != "",
	}

	if err := tmpl.ExecuteTemplate(w, "create_project_step9.html", tvals); err != nil {
		log.Printf("createProjectStep9 failed to execute template: %v", err)
	}
}

func validationErrorStep8(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	useremail := userEmail(r)

	if err := r.ParseForm(); err != nil {
		ctx := r.Context()
		ServeError(ctx, w, err)
		return
	}

	numgroups, err := strconv.Atoi(r.FormValue("numgroups"))
	if err != nil {
		log.Printf("validationErrorStep8: %v", err)
	}

	numvar, err := strconv.Atoi(r.FormValue("numvar"))
	if err != nil {
		log.Printf("validationErrorStep8: %v", err)
	}

	tvals := struct {
		User          string
		LoggedIn      bool
		Name          string
		NumGroups     int
		Pkey          string
		GroupNames    string
		StoreRawData  bool
		Numvar        int
		SamplingRates string
	}{
		User:          useremail,
		LoggedIn:      useremail != "",
		Name:          r.FormValue("project_name"),
		GroupNames:    r.FormValue("group_names"),
		NumGroups:     numgroups,
		Numvar:        numvar,
		StoreRawData:  r.FormValue("store_rawdata") == "true",
		SamplingRates: r.FormValue("rates"),
	}

	if err := tmpl.ExecuteTemplate(w, "validation_error_step8.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}
