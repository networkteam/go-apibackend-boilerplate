package mail

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/friendsofgo/errors"
)

//go:embed templates/*.txt
//go:embed templates/*.html
var templatesFS embed.FS

// ExecuteTemplates executes both text and HTML templates with the given name and data.
// Returns both text and HTML bodies for multipart emails.
func ExecuteTemplates(name string, data any) (textBody, htmlBody string, err error) {
	textBody, err = ExecuteTextTemplate(name, data)
	if err != nil {
		return "", "", errors.Wrap(err, "executing text template")
	}

	htmlBody, err = ExecuteHTMLTemplate(name, data)
	if err != nil {
		return "", "", errors.Wrap(err, "executing html template")
	}

	return textBody, htmlBody, nil
}

// ExecuteTextTemplate executes a text template and returns the body.
func ExecuteTextTemplate(name string, data any) (body string, err error) {
	templates, err := template.
		New("").
		Funcs(sprig.TxtFuncMap()).
		ParseFS(templatesFS, "templates/*.txt")
	if err != nil {
		return "", errors.Wrap(err, "parsing text templates")
	}

	templateName := fmt.Sprintf("%s.txt", name)

	var buffer bytes.Buffer
	err = templates.ExecuteTemplate(&buffer, templateName, data)
	if err != nil {
		return "", errors.Wrap(err, "executing text template")
	}
	messageText := buffer.String()

	// Convert CRLF to LF for consistent line endings
	messageText = strings.ReplaceAll(messageText, "\r\n", "\n")

	return messageText, nil
}

// ExecuteHTMLTemplate executes an HTML template and returns the body.
func ExecuteHTMLTemplate(name string, data any) (body string, err error) {
	tmpl, err := template.
		New("").
		Funcs(sprig.HtmlFuncMap()).
		ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return "", errors.Wrap(err, "parsing html templates")
	}

	templateName := fmt.Sprintf("%s.html", name)

	var buffer bytes.Buffer
	err = tmpl.ExecuteTemplate(&buffer, templateName, data)
	if err != nil {
		return "", errors.Wrap(err, "executing html template")
	}

	return buffer.String(), nil
}
