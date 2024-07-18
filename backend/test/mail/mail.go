package mail

import (
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func RequireParseMailMessage(t *testing.T, content string) *mail.Message {
	t.Helper()

	parsedMsg, err := mail.ReadMessage(strings.NewReader(content))
	require.NoError(t, err, "Reading mail message from buffer")

	return parsedMsg
}

func AssertMailMessageHeaderEquals(t *testing.T, msg *mail.Message, headerName string, expectedValue string) {
	t.Helper()

	rawValue := msg.Header.Get(headerName)
	dec := new(mime.WordDecoder)
	subject, err := dec.DecodeHeader(rawValue)
	require.NoError(t, err)

	assert.Equal(t, subject, expectedValue)
}

func AssertMailMessageBodyContains(t *testing.T, msg *mail.Message, substr string) {
	t.Helper()

	var (
		err  error
		body []byte
	)

	switch msg.Header.Get("Content-Transfer-Encoding") {
	case "quoted-printable":
		r := quotedprintable.NewReader(msg.Body)
		body, err = io.ReadAll(r)
		require.NoError(t, err)
	default:
		t.Errorf("Unsupported Content-Transfer-Encoding: %q", msg.Header.Get("Content-Transfer-Encoding"))
	}

	assert.Contains(t, string(body), substr)
}

// AssertMailMessageHasFileAttachment asserts that a mail message has a certain file attachment
//
// See https://github.com/kirabou/parseMIMEemail.go for parsing multipart messages.
func AssertMailMessageHasFileAttachment(t *testing.T, msg *mail.Message, expectedFileName string) {
	t.Helper()

	mediaType, params, err := mime.ParseMediaType(msg.Header.Get("Content-Type"))
	if err != nil {
		t.Fatalf("parsing media type: %v", err)
	}

	if !strings.HasPrefix(mediaType, "multipart/") {
		t.Fatalf("not a multipart media type: %s", mediaType)
	}

	reader := multipart.NewReader(msg.Body, params["boundary"])
	if reader == nil {
		t.Fatal("no multipart reader")
		return
	}

	var fileNames []string

	for {
		newPart, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			t.Fatalf("reading next part: %v", err)
		}

		if newPart.FileName() == expectedFileName {
			// We found the file we were looking for
			return
		}

		fileNames = append(fileNames, newPart.FileName())
	}

	t.Fatalf("no file attachment for %s found, got attachments: %v", expectedFileName, fileNames)
}
