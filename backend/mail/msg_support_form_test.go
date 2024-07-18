package mail_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"myvendor.mytld/myproject/backend/domain"
	"myvendor.mytld/myproject/backend/mail"
	test_mail "myvendor.mytld/myproject/backend/test/mail"
)

func TestSupportFormMsg_ToMessage(t *testing.T) {
	tt := []struct {
		name                   string
		msg                    mail.SupportFormMsg
		config                 mail.Config
		expectedHeaders        map[string]string
		expectedFileAttachment string
	}{
		{
			name: "without attachment",
			msg: mail.SupportFormMsg{
				SenderEmailAddress: "test@example.com",
				SenderName:         "Max Mustermann",
				OrganisationName:   "Acme Inc.",
				Subject:            "Testnachricht",
				Message:            "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod",
			},
			config: mail.DefaultConfig(domain.DefaultConfig()),
			expectedHeaders: map[string]string{
				"To":      "<app@example.com>",
				"Subject": "Neue Kontaktanfrage von Max Mustermann (Acme Inc.)",
			},
		},
		{
			name: "with attachment",
			msg: mail.SupportFormMsg{
				SenderEmailAddress: "test@example.com",
				SenderName:         "Max Mustermann",
				OrganisationName:   "Acme Inc.",
				Subject:            "Testnachricht",
				Message:            "Lorem ipsum dolor sit amet, consetetur sadipscing elitr, sed diam nonumy eirmod",
				FileName:           "test/my-screenshot.png",
				AttachedFile:       strings.NewReader("not actually an image..."),
			},
			config: mail.DefaultConfig(domain.DefaultConfig()),
			expectedHeaders: map[string]string{
				"To":      "<app@example.com>",
				"Subject": "Neue Kontaktanfrage von Max Mustermann (Acme Inc.)",
			},
			expectedFileAttachment: "my-screenshot.png",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			mailMsg, err := tc.msg.ToMessage(tc.config)
			require.NoError(t, err)

			parsedMsg := requireParseGomailMessage(t, mailMsg)

			for headerName, headerValue := range tc.expectedHeaders {
				test_mail.AssertMailMessageHeaderEquals(t, parsedMsg, headerName, headerValue)
			}

			if tc.expectedFileAttachment != "" {
				test_mail.AssertMailMessageHasFileAttachment(t, parsedMsg, tc.expectedFileAttachment)
			}
		})
	}
}
