package graphql

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
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

	req, err := http.NewRequest(http.MethodPost, "http://localhost/query", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("could not build GraphQL request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req
}

func Handle(t *testing.T, deps api.ResolverDependencies, req *http.Request, dst interface{}) *httptest.ResponseRecorder {
	t.Helper()

	// Use default config if config is zero value
	if deps.Config == (domain.Config{}) {
		deps.Config = domain.DefaultConfig()
		// Use a reduced hash cost
		deps.Config.HashCost = bcrypt.MinCost
	}

	if deps.TimeSource == nil {
		deps.TimeSource = test.FixedTime()
	}

	if deps.Mailer == nil {
		sender := fixture.NewSender()
		deps.Mailer = mail.NewMailer(sender, mail.DefaultConfig(domain.DefaultConfig()))
	}

	graphqlHandler := api_handler.NewGraphqlHandler(deps, api_handler.HandlerConfig{
		DisableRecover: true,
	})
	srv := http_api.MiddlewareStack(deps, graphqlHandler)

	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("could not read response body: %v", err)
	}

	err = json.Unmarshal(body, dst)
	if err != nil {
		t.Fatalf("could not decode response JSON: %v", err)
	}

	return w
}

func NewMultipartRequest(t *testing.T, body bytes.Buffer, query GraphqlQuery, files map[string]MultipartFileInfo) *http.Request {
	t.Helper()

	mw := multipart.NewWriter(&body)

	// Add operations from query

	fw, err := mw.CreateFormField("operations")
	if err != nil {
		t.Fatalf("could not create multipart form field: %v", err)
	}
	enc := json.NewEncoder(fw)
	if err = enc.Encode(query); err != nil {
		t.Fatalf("could not marshal GraphQL operations: %v", err)
	}

	// Add map from files to variables

	fw, err = mw.CreateFormField("map")
	if err != nil {
		t.Fatalf("could not create multipart form field: %v", err)
	}
	enc = json.NewEncoder(fw)

	fileMap := make(map[string][]string)
	for name, fileInfo := range files {
		fileMap[name] = fileInfo.Variables
	}

	if err = enc.Encode(fileMap); err != nil {
		t.Fatalf("could not marshal GraphQL map: %v", err)
	}

	// Add form files

	for name, fileInfo := range files {
		fw, err = mw.CreateFormFile(name, fileInfo.Name)
		if err != nil {
			t.Fatalf("could not create multipart form file: %v", err)
		}

		if _, err = io.Copy(fw, fileInfo.Reader); err != nil {
			t.Fatalf("could not read fixture file into multipart request: %v", err)
		}
	}

	if err = mw.Close(); err != nil {
		t.Fatalf("could not close multipart writer: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, "http://localhost/query", &body)
	if err != nil {
		t.Fatalf("could not build GraphQL request: %v", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())

	return req
}

type MultipartFileInfo struct {
	Name      string
	Variables []string
	Reader    io.Reader
}

type GraphqlQuery struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

type GraphqlErrors struct {
	Errors []struct {
		Message    string        `json:"message"`
		Path       []interface{} `json:"path"`
		Extensions struct {
			Field string `json:"field"`
			Type  string `json:"type"`
			Code  string `json:"code"`
		} `json:"extensions"`
	} `json:"errors"`
}
