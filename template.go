package main

import (
	"bytes"
	"html/template"

	"github.com/sirupsen/logrus"
)

type TemplateWriter struct {
	tmpl *template.Template
}

func NewTemplateWriter(t string) (*TemplateWriter, error) {
	tmpl, err := template.New("tmpl").Parse(t)
	if err != nil {
		return nil, err
	}

	return &TemplateWriter{
		tmpl: tmpl,
	}, nil
}

func (t *TemplateWriter) Write(v interface{}) string {
	var tpl bytes.Buffer
	err := t.tmpl.Execute(&tpl, v)
	if err != nil {
		logrus.WithError(err).Error("error writing reply...")
	}

	return tpl.String()
}
