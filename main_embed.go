//go:build embed

package main

import (
	"html/template"
	"net/http"

	"github.com/MhunterDev/img2ascii/source/www"
)

func getStaticFS() (tmpl *template.Template, staticFS http.FileSystem, err error) {
	tmpl, err = template.ParseFS(www.StaticFiles, "index.html")
	if err != nil {
		return nil, nil, err
	}
	return tmpl, http.FS(www.StaticFiles), nil
}
