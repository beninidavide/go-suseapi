package bugzilla_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/headzoo/surf"
)

const sample = `
<!DOCTYPE html PUBLIC "-//W3C//DTD HTML 4.01 Transitional//EN"
                      "http://www.w3.org/TR/html4/loose.dtd">
<html lang="en">
<head>
<title>Title</title>
</head>
<body>
<form name="changeform" id="changeform" method="post" action="process_bug.cgi">
 <input type="hidden" name="defined_groups" 
               value="foobaronly">

      <input type="checkbox" value="foobaronly"
             name="groups" id="group_10" checked="checked">
           <input type="hidden" name="defined_reporter_accessible" value="1">
          <input type="checkbox" value="1"
                 name="reporter_accessible" id="reporter_accessible">
          <label for="reporter_accessible">Reporter</label>
        </div>
        <div>
            <input type="hidden" name="defined_cclist_accessible" value="1">
          <input type="checkbox" value="1"
                 name="cclist_accessible" id="cclist_accessible">
          <label for="cclist_accessible">CC List</label>
</form>
</body>
`

const testMessage = "surfs error has been fixed, remove the workarounds in the code!"

func TestSurfCantHandleShortChecked(t *testing.T) {
	ts0 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/show_bug.cgi":
			io.WriteString(w, sample)
		case "/process_bug.cgi":
			r.ParseForm()
			cclist := r.Form.Get("cclist_accessible")
			reporter := r.Form.Get("reporter_accessible")
			groups := r.Form.Get("groups")

			if groups != "foobaronly" {
				t.Errorf("failed to parse groups: %q", groups)
			}

			if cclist != "" || reporter != "" {
				t.Error(testMessage)
			}
		}
	}))
	defer ts0.Close()

	bow := surf.NewBrowser()
	bow.Open(ts0.URL + "/show_bug.cgi")
	form, err := bow.Form("form[name=changeform]")
	if err != nil {
		t.Error("failed to find the form")
	}
	form.Submit()
}
