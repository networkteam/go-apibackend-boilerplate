package mail_test

import (
	"bytes"
	"net/mail"
	"testing"

	"github.com/stretchr/testify/require"

	pkg_mail "myvendor.mytld/myproject/backend/mail"
)

func requireParseMessage(t *testing.T, msg *pkg_mail.Message) *mail.Message {
	t.Helper()

	var buf bytes.Buffer
	_, err := msg.WriteTo(&buf)
	require.NoError(t, err, "Writing mail message to buffer")

	parsedMsg, err := mail.ReadMessage(&buf)
	require.NoError(t, err, "Reading mail message from buffer")

	return parsedMsg
}
