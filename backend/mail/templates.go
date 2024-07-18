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
var templatesFS embed.FS

func executeTemplate(name string, data any) (subject, body string, err error) {
	templates, err := template.
		New("").
		Funcs(sprig.TxtFuncMap()).
		ParseFS(templatesFS, "templates/*.txt")
	if err != nil {
		return "", "", errors.Wrap(err, "parsing templates")
	}

	templateName := fmt.Sprintf("%s.txt", name)

	var buffer bytes.Buffer
	err = templates.ExecuteTemplate(&buffer, templateName, data)
	if err != nil {
		return "", "", errors.Wrap(err, "executing template")
	}
	messageText := buffer.String()

	messageParts := strings.SplitN(messageText, "\n\n", 2)

	return messageParts[0], messageParts[1], nil
}
