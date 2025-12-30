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
	"reflect"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql"
	"golang.org/x/crypto/bcrypt"

	"myvendor.mytld/myproject/backend/api"
	"myvendor.mytld/myproject/backend/api/graph/admin"
	"myvendor.mytld/myproject/backend/api/graph/public"
	api_handler "myvendor.mytld/myproject/backend/api/handler"
	http_api "myvendor.mytld/myproject/backend/api/http"
	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/mail"
	"myvendor.mytld/myproject/backend/mail/fixture"
	"myvendor.mytld/myproject/backend/test"
	"myvendor.mytld/myproject/backend/test/auth"
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
	if reflect.DeepEqual(deps.Config, domain.Config{}) {
		deps.Config = domain.DefaultConfig()
	}
	// Always use a reduced hash cost for tests
	deps.Config.HashCost = bcrypt.MinCost
	deps.Config.JWTSecret = auth.FixedJWTSecret

	if deps.TimeSource == nil {
		deps.TimeSource = test.FixedTime()
	}

	if deps.Mailer == nil {
		sender := fixture.NewSender()
		deps.Mailer = mail.NewMailer(sender, mail.DefaultConfig(deps.Config))
	}
}

func HandlePublic(t *testing.T, deps api.ResolverDependencies, req *http.Request, dst interface{}) *httptest.ResponseRecorder {
	t.Helper()

	SetTestDependencies(t, &deps)

	apiHandlerConfig := api_handler.Config{
		DisableRecover: true,
	}
	publicExecutableSchema := public.BuildExecutableSchema(deps, apiHandlerConfig)
	return handleSchema(t, deps, publicExecutableSchema, req, dst)
}

func HandleAdmin(t *testing.T, deps api.ResolverDependencies, req *http.Request, dst interface{}) *httptest.ResponseRecorder {
	t.Helper()

	SetTestDependencies(t, &deps)

	apiHandlerConfig := api_handler.Config{
		DisableRecover: true,
	}
	publicExecutableSchema := admin.BuildExecutableSchema(deps, apiHandlerConfig)
	return handleSchema(t, deps, publicExecutableSchema, req, dst)
}

func handleSchema(t *testing.T, deps api.ResolverDependencies, executableSchema graphql.ExecutableSchema, req *http.Request, dst interface{}) *httptest.ResponseRecorder {
	t.Helper()

	apiHandlerConfig := api_handler.Config{
		DisableRecover: true,
	}
	graphqlHandler := api_handler.NewGraphqlHandler(deps, apiHandlerConfig, executableSchema)
	srv := http_api.MiddlewareStackWithAuth(deps, graphqlHandler)

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	body := rec.Body.Bytes()
	err := json.Unmarshal(body, dst)
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

type GraphqlQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

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
