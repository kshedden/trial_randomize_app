package randomize

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/context"
)

// EditSharing is page 1 for changing the sharing settings
func EditSharing(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	ctx := context.Background()
	user := userEmail(r)
	pkey := r.FormValue("pkey")
	shr := splitKey(pkey)
	owner := shr[0]
	projectName := shr[1]

	if strings.ToLower(owner) != strings.ToLower(user) {
		msg := "Only the owner of a project can manage sharing."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		return
	}

	sharedUsers, err := getSharedUsers(ctx, pkey)
	if err != nil {
		log.Printf("editSharing failed to retrieve sharing: %v %v", projectName, owner)
	}

	var sul []string
	for k := range sharedUsers {
		sul = append(sul, k)
	}

	tvals := struct {
		User           string
		LoggedIn       bool
		SharedUsers    []string
		AnySharedUsers bool
		ProjectName    string
		Pkey           string
	}{
		User:           user,
		LoggedIn:       user != "",
		SharedUsers:    sul,
		AnySharedUsers: len(sul) > 0,
		ProjectName:    projectName,
		Pkey:           pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "edit_sharing.html", tvals); err != nil {
		log.Printf("editSharing failed to execute template: %v", err)
	}
}

// EditSharingConfirm is page 2 for editing the sharing settings
func EditSharingConfirm(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		Serve404(w)
		return
	}

	user := userEmail(r)
	pkey := r.FormValue("pkey")

	spkey := splitKey(pkey)
	projectName := spkey[1]

	ap := r.FormValue("additional_people")
	addUsers := cleanSplit(ap, ",")
	for k, x := range addUsers {
		addUsers[k] = strings.ToLower(x)
	}

	err := addSharing(pkey, addUsers)
	if err != nil {
		msg := "Database error: unable to update sharing information."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		log.Printf("editSharingConfirm [1]: %v", err)
		return
	}

	removeUsers := r.Form["remove_users"]
	err = removeSharing(pkey, removeUsers)
	if err != nil {
		msg := "Database error: unable to update sharing information."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, "/dashboard")
		log.Printf("editSharingConfirm [2]: %v", err)
		return
	}

	tvals := struct {
		User        string
		LoggedIn    bool
		ProjectName string
		Pkey        string
	}{
		User:        user,
		LoggedIn:    user != "",
		ProjectName: projectName,
		Pkey:        pkey,
	}

	if err := tmpl.ExecuteTemplate(w, "edit_sharing_confirm.html", tvals); err != nil {
		log.Printf("Failed to execute template: %v", err)
	}
}
