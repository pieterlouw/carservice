package app

import (
	"bytes"
	"fmt"
	htmpl "html/template"
	"log"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
)

// Source code used here is based on the same principles as seen in:  https://github.com/sourcegraph/thesrc

var (
	// TemplateDir is the directory containing the html/template template files.
	TemplateDir = filepath.Join(defaultBase("github.com/pieterlouw/carservice/app"), "templates")
)

// loadTemplates parse and load all HTML templates
func loadTemplates() {
	err := parseHTMLTemplates([][]string{
		{"bookings/add_form.html", "bookings/common.html", "common.html", "layout.html"},
		{"bookings/view.html", "bookings/common.html", "common.html", "layout.html"},
		{"bookings/list.html", "bookings/common.html", "common.html", "layout.html"},

		{"dashboard.html", "common.html", "layout.html"},
		{"error.html", "common.html", "layout.html"},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func renderTemplate(w http.ResponseWriter, r *http.Request, name string, status int, data interface{}) error {
	w.WriteHeader(status)
	if ct := w.Header().Get("content-type"); ct == "" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	t := templates[name]
	if t == nil {
		return fmt.Errorf("Template %s not found", name)
	}

	// Write to a buffer to properly catch errors and avoid partial output written to the http.ResponseWriter
	var buf bytes.Buffer
	err := t.Execute(&buf, data)
	if err != nil {
		return err
	}
	_, err = buf.WriteTo(w)
	return err
}

var templates = map[string]*htmpl.Template{}

func parseHTMLTemplates(sets [][]string) error {
	for _, set := range sets {
		t := htmpl.New("")
		/*t.Funcs(htmpl.FuncMap{
			"itoa": strconv.Itoa,
		})*/

		_, err := t.ParseFiles(joinTemplateDir(TemplateDir, set)...)
		if err != nil {
			return fmt.Errorf("template %v: %s", set, err)
		}

		t = t.Lookup("ROOT")
		if t == nil {
			return fmt.Errorf("ROOT template not found in %v", set)
		}
		templates[set[0]] = t
	}
	return nil
}

func joinTemplateDir(base string, files []string) []string {
	result := make([]string, len(files))
	for i := range files {
		result[i] = filepath.Join(base, files[i])
	}
	return result
}

func urlDomain(urlStr string) string {
	url, err := url.Parse(urlStr)
	if err != nil {
		return "invalid URL"
	}
	return strings.TrimPrefix(url.Host, "www.")
}
