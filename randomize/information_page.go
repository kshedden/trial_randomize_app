package randomize

import (
	"log"
	"net/http"
)

// InformationPage displays a page of information about this application.
func InformationPage(w http.ResponseWriter, r *http.Request) {

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

	if err := tmpl.ExecuteTemplate(w, "information_page.html", tvals); err != nil {
		log.Printf("Execute template faile in informationPage: %v", err)
	}
}
