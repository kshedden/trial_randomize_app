package randomize

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ViewCompleteData displays the complete data in raw text form.
func ViewCompleteData(w http.ResponseWriter, r *http.Request) {

	if r.Method != "GET" {
		Serve404(w)
		return
	}

	pkey := r.FormValue("pkey")

	if !checkAccess(pkey, r) {
		return
	}

	proj, _ := getProjectFromKey(pkey)
	if !proj.StoreRawData {
		msg := "Complete data are not stored for this project."
		rmsg := "Return to dashboard"
		messagePage(w, r, msg, rmsg, fmt.Sprintf("/project_dashboard?pkey=%s", pkey))
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// Header line
	_, _ = io.WriteString(w, "Subject id,Assignment date,Assignment time,")
	_, _ = io.WriteString(w, "Assigned group,Final group,Included,Assigner")
	for _, va := range proj.Variables {
		_, _ = io.WriteString(w, ",")
		_, _ = io.WriteString(w, va.Name)
	}
	_, _ = io.WriteString(w, "\n")

	for _, rec := range proj.RawData {
		_, _ = io.WriteString(w, rec.SubjectId)
		_, _ = io.WriteString(w, ",")
		t := rec.AssignedTime
		loc, _ := time.LoadLocation("America/New_York")
		t = t.In(loc)
		_, _ = io.WriteString(w, t.Format("2006-1-2"))
		_, _ = io.WriteString(w, ",")
		_, _ = io.WriteString(w, t.Format("3:04 PM EST"))
		_, _ = io.WriteString(w, ",")
		_, _ = io.WriteString(w, rec.AssignedGroup)
		_, _ = io.WriteString(w, ",")
		_, _ = io.WriteString(w, rec.CurrentGroup)
		_, _ = io.WriteString(w, ",")
		if rec.Included {
			_, _ = io.WriteString(w, "Yes,")
		} else {
			_, _ = io.WriteString(w, "No,")
		}
		_, _ = io.WriteString(w, rec.Assigner+",")
		_, _ = io.WriteString(w, strings.Join(rec.Data, ","))
		_, _ = io.WriteString(w, "\n")
	}
}
