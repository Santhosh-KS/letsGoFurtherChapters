package main

import (
	"html/template"
	"io/fs"
	"path/filepath"
	"snippetbox.glyphsmiths.com/internal/models"
	"snippetbox.glyphsmiths.com/ui"
	"time"
)

type templateData struct {
	CurrentYear     int
	Snippet         models.Snippet
	Snippets        []models.Snippet
	Form            any
	Flash           string
	IsAuthenticated bool
	CSRFToken       string
}

func newTemplateCache() (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}

	// pages, err := filepath.Glob("./ui/html/pages/*.templ")
	pages, err := fs.Glob(ui.Files, "html/pages/*.templ")
	if err != nil {
		return nil, err
	}

	for _, page := range pages {
		name := filepath.Base(page)

		patterns := []string{
			"html/base.templ",
			"html/partials/*.templ",
			page,
		}

		// Parse the base template file into a template set.
		// ts, err := template.New(name).Funcs(functions).ParseFiles("./ui/html/base.templ")
		ts, err := template.New(name).Funcs(functions).ParseFS(ui.Files, patterns...)
		if err != nil {
			return nil, err
		}

		/* // Call ParseGlob() *on this template set* to add any partials.
		ts, err = ts.ParseGlob("./ui/html/partials/*.templ")
		if err != nil {
			return nil, err
		}

		// Call ParseFiles() *on this template set* to add the  page template.
		ts, err = ts.ParseFiles(page)
		if err != nil {
			return nil, err
		} */

		// Add the template set to the map as normal...
		cache[name] = ts
	}
	return cache, nil
}

func humanDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("02 Jan 2006 at 15:04")
}

var functions = template.FuncMap{
	"humanDate": humanDate,
}
