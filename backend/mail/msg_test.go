package mail_test

import (
	"bytes"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/require"
	gomail "github.com/wneessen/go-mail"
)

func requireParseGomailMessage(t *testing.T, msg *gomail.Msg) *mail.Message {
	t.Helper()

	var buf bytes.Buffer
	_, err := msg.WriteTo(&buf)
	require.NoError(t, err, "Writing gomail message to buffer")

	parsedMsg, err := mail.ReadMessage(&buf)
	require.NoError(t, err, "Reading mail message from buffer")

	return parsedMsg
}
