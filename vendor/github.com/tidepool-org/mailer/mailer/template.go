package mailer

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"html/template"
)

type EmailTemplate struct {
	body    *template.Template
	name    string
	subject *template.Template
}

func NewEmailTemplate(name string, subject string, body string) (*EmailTemplate, error) {
	if name == "" {
		return nil, errors.New("email template name cannot be empty")
	}
	if subject == "" {
		return nil, errors.New("email template subject cannot be empty")
	}
	if body == "" {
		return nil, errors.New("email template body cannot be empty")
	}
	precompiledSubject, err := template.New(fmt.Sprintf("%s_subject", name)).Parse(subject)
	if err != nil {
		return nil, err
	}
	precompiledBody, err := template.New(fmt.Sprintf("%s_body", name)).Parse(body)
	if err != nil {
		return nil, err
	}
	return &EmailTemplate{
		body:    precompiledBody,
		name:    name,
		subject: precompiledSubject,
	}, nil
}

func (e *EmailTemplate) RenderToEmail(params interface{}, email *Email) error {
	var subject bytes.Buffer
	var body bytes.Buffer

	if err := e.subject.Execute(&subject, params); err != nil {
		return err
	}
	if err := e.body.Execute(&body, params); err != nil {
		return err
	}

	email.Subject = subject.String()
	email.Body = body.String()
	return nil
}
