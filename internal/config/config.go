package config

import "text/template"

var (
	StaticPath     = "./../../web/static/"
	IndexFile      = "index.html"
	VoteFile       = "vote.html"
	TermosFile     = "termos-uso-privacidade.html"
	ResultFile     = "result.html"
	ResultTemplate *template.Template
)
