package config

import "text/template"

var (
	StaticPath     = "./../../web/static/"
	IndexFile      = StaticPath + "index.html"
	VoteFile       = StaticPath + "vote.html"
	TermosFile     = StaticPath + "termos-uso-privacidade.html"
	ResultFile     = StaticPath + "result.html"
	InfoFile       = StaticPath + "info.html"
	AdminLogin     = StaticPath + "admin.html"
	Dashboard      = StaticPath + "dashboard.html"
	ResultTemplate *template.Template
)
