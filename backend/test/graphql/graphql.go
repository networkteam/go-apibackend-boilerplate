package graphql

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"myvendor.mytld/myproject/backend/api"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	http_api "myvendor.mytld/myproject/backend/api/http"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/mail"
	"myvendor.mytld/myproject/backend/mail/fixture"
	"myvendor.mytld/myproject/backend/test"
)

func NewRequest(t *testing.T, query GraphqlQuery) *http.Request {
	t.Helper()

	data, err := json.Marshal(query)
	if err != nil {
		t.Fatalf("could not marshal GraphQL query: %v", err)
	}

	//nolint:noctx
	req, err := http.NewRequest(http.MethodPost, "http://localhost/query", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("could not build GraphQL request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func SetTestDependencies(t *testing.T, deps *api.ResolverDependencies) {
	t.Helper()

	// Use default config if config is zero value
	if deps.Config == (domain.Config{}) {
		deps.Config = domain.DefaultConfig()
	}
	// Always use a reduced hash cost for tests
	deps.Config.HashCost = bcrypt.MinCost

	if deps.TimeSource == nil {
		deps.TimeSource = test.FixedTime()
	}

	if deps.Mailer == nil {
		sender := fixture.NewSender()
		deps.Mailer = mail.NewMailer(sender, mail.DefaultConfig(domain.DefaultConfig()))
	}
}

func Handle(t *testing.T, deps api.ResolverDependencies, req *http.Request, dst interface{}) *httptest.ResponseRecorder {
	t.Helper()

	SetTestDependencies(t, &deps)

	graphqlHandler := api_handler.NewGraphqlHandler(deps, api_handler.Config{
		DisableRecover: true,
	})
	srv := http_api.MiddlewareStackWithAuth(deps, graphqlHandler)

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	err := json.Unmarshal(rec.Body.Bytes(), dst)
	if err != nil {
		t.Fatalf("could not decode response JSON: %v", err)
	}

	return rec
}

func NewMultipartRequest(t *testing.T, body bytes.Buffer, query GraphqlQuery, files map[string]MultipartFileInfo) *http.Request {
	t.Helper()

	multipartWriter := multipart.NewWriter(&body)

	// Add operations from query
	formField, err := multipartWriter.CreateFormField("operations")
	if err != nil {
		t.Fatalf("could not create multipart form field: %v", err)
	}
	enc := json.NewEncoder(formField)
	if err = enc.Encode(query); err != nil {
		t.Fatalf("could not marshal GraphQL operations: %v", err)
	}

	// Add map from files to variables
	formField, err = multipartWriter.CreateFormField("map")
	if err != nil {
		t.Fatalf("could not create multipart form field: %v", err)
	}
	enc = json.NewEncoder(formField)

	fileMap := make(map[string][]string)
	for name, fileInfo := range files {
		fileMap[name] = fileInfo.Variables
	}

	if err = enc.Encode(fileMap); err != nil {
		t.Fatalf("could not marshal GraphQL map: %v", err)
	}

	// Add form files
	for name, fileInfo := range files {
		formField, err = multipartWriter.CreateFormFile(name, fileInfo.Name)
		if err != nil {
			t.Fatalf("could not create multipart form file: %v", err)
		}

		if fileInfo.Filename != "" {
			func() {
				data, err := os.ReadFile(fileInfo.Filename)
				if err != nil {
					t.Fatalf("could not read fixture file: %v", err)
				}
				_, err = formField.Write(data)
				if err != nil {
					t.Fatalf("could not write fixture file into multipart request: %v", err)
				}
			}()
		} else if fileInfo.Reader != nil {
			_, err = io.Copy(formField, fileInfo.Reader)
			if err != nil {
				t.Fatalf("could not read fixture file into multipart request: %v", err)
			}
		} else {
			t.Fatalf("no reader or filename given for multipart file %q", name)
		}
	}

	if err = multipartWriter.Close(); err != nil {
		t.Fatalf("could not close multipart writer: %v", err)
	}

	//nolint:noctx
	req, err := http.NewRequest(http.MethodPost, "http://localhost/query", &body)
	if err != nil {
		t.Fatalf("could not build GraphQL request: %v", err)
	}
	req.Header.Set("Content-Type", multipartWriter.FormDataContentType())

	return req
}

type MultipartFileInfo struct {
	Name      string
	Variables []string
	Reader    io.Reader
	Filename  string
}

//nolint:revive // Better readability if we repeat Graphql
type GraphqlQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

//nolint:revive // Better readability if we repeat Graphql
type GraphqlErrors struct {
	Errors []GraphqlError `json:"errors"`
}

func (e GraphqlErrors) String() string {
	var sb strings.Builder
	for i, err := range e.Errors {
		err.writeTo(&sb)
		if i < len(e.Errors)-1 {
			sb.WriteRune('\n')
		}
	}
	return sb.String()
}

//nolint:revive // Better readability if we repeat Graphql
type GraphqlError struct {
	Message    string                 `json:"message"`
	Path       []any                  `json:"path"`
	Extensions GraphqlErrorExtensions `json:"extensions"`
}

func (e GraphqlError) String() string {
	var b strings.Builder
	e.writeTo(&b)
	return b.String()
}

func (e GraphqlError) writeTo(w io.Writer) {
	if len(e.Path) > 0 {
		_, _ = fmt.Fprintf(w, "%v", e.Path)
	} else {
		_, _ = fmt.Fprint(w, "<empty path>")
	}
	if e.Message != "" {
		_, _ = fmt.Fprintf(w, " %s", e.Message)
	}
	var extensions []string
	if e.Extensions.Field != "" {
		extensions = append(extensions, fmt.Sprintf("field: %q", e.Extensions.Field))
	}
	if e.Extensions.Type != "" {
		extensions = append(extensions, fmt.Sprintf("type: %q", e.Extensions.Type))
	}
	if e.Extensions.Code != "" {
		extensions = append(extensions, fmt.Sprintf("code: %q", e.Extensions.Code))
	}
	if len(extensions) > 0 {
		_, _ = fmt.Fprintf(w, " (%s)", strings.Join(extensions, ", "))
	}
}

//nolint:revive // Better readability if we repeat Graphql
type GraphqlErrorExtensions struct {
	Field string `json:"field"`
	Type  string `json:"type"`
	Code  string `json:"code"`
}

type FieldsError struct {
	Errors []struct {
		Path      []string
		Code      string
		Arguments []string
	}
}
