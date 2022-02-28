package mail_test

import (
	"bytes"
	"io"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/gomail.v2"
)

func assertMailMessageHeaderEquals(t *testing.T, msg *mail.Message, headerName string, expectedValue string) {
	t.Helper()

	rawValue := msg.Header.Get(headerName)
	dec := new(mime.WordDecoder)
	subject, err := dec.DecodeHeader(rawValue)
	require.NoError(t, err)

	assert.Equal(t, subject, expectedValue)
}

func requireParseGomailMessage(t *testing.T, err error, msg *gomail.Message) *mail.Message {
	t.Helper()

	var buf bytes.Buffer
	_, err = msg.WriteTo(&buf)
	require.NoError(t, err, "Writing gomail message to buffer")

	parsedMsg, err := mail.ReadMessage(&buf)
	require.NoError(t, err, "Reading mail message from buffer")

	return parsedMsg
}

// Assert that a mail message has a certain file attachment
//
// See https://github.com/kirabou/parseMIMEemail.go for parsing multipart messages.
func assertMailMessageHasFileAttachment(t *testing.T, msg *mail.Message, expectedFileName string) {
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
		new_part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("reading next part: %v", err)
		}

		if new_part.FileName() == expectedFileName {
			// We found the file we were looking for
			return
		}

		fileNames = append(fileNames, new_part.FileName())
	}

	t.Fatalf("no file attachment for %s found, got attachments: %v", expectedFileName, fileNames)
}
