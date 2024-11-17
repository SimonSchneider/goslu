package srvu

import (
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log"
	"os"
	"sync"
)

func getFS(embeddedFS fs.FS, dir string, watch bool) (fs.FS, error) {
	if watch {
		if _, err := os.Stat(dir); err == nil {
			return os.DirFS(dir), nil
		}
	}
	f, err := fs.Sub(embeddedFS, dir)
	if err != nil {

		return nil, fmt.Errorf("failed to get static files: %v", err)
	}
	return f, nil
}

type TemplateProvider interface {
	Lookup(name string) *template.Template
	ExecuteTemplate(w io.Writer, name string, data interface{}) error
}

type TemplateProviderFunc func() *template.Template

func (t TemplateProviderFunc) Lookup(name string) *template.Template {
	return t().Lookup(name)
}

func (t TemplateProviderFunc) ExecuteTemplate(w io.Writer, name string, data interface{}) error {
	return t().ExecuteTemplate(w, name, data)
}

func GetPublicAndTemplates(embeddedFS fs.FS, dir string, watch bool, public string, tmplPatterns ...string) (fs.FS, TemplateProvider, error) {
	f, err := getFS(embeddedFS, dir, watch)
	if err != nil {
		return nil, nil, err
	}
	publicFS, err := fs.Sub(f, public)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get public files: %v", err)
	}
	getTmpl := func() *template.Template {
		tmpl, err := template.ParseFS(f, tmplPatterns...)
		if err != nil {
			log.Fatal(fmt.Errorf("failed to parse templates: %v", err))
		}
		return tmpl
	}
	if watch {
		return publicFS, TemplateProviderFunc(getTmpl), nil
	}
	return publicFS, TemplateProviderFunc(sync.OnceValue(getTmpl)), nil
}
