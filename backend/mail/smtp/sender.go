package smtp

import (
	"context"
	std_errors "errors"
	"log/slog"
	"mime"

	"github.com/friendsofgo/errors"
	"github.com/networkteam/slogutils"
	gomail "github.com/wneessen/go-mail"

	"myvendor.mytld/myproject/backend/mail"
)

type sender struct {
	Dialer *gomail.Client
	host   string
	port   int
}

var _ mail.Sender = new(sender)

func NewSender(host string, port int, username, password, tlsPolicy string) (mail.Sender, error) {
	policy, err := getTLSPolicy(tlsPolicy)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	client, err := gomail.NewClient(
		host,
		gomail.WithPort(port),
		gomail.WithTLSPolicy(policy),
	)
	if err != nil {
		return nil, errors.Wrap(err, "creating mail client")
	}

	if username != "" && password != "" {
		client.SetSMTPAuth(gomail.SMTPAuthLogin)
		client.SetUsername(username)
		client.SetPassword(password)
	} else {
		client.SetSMTPAuth(gomail.SMTPAuthNoAuth)
	}

	return &sender{
		Dialer: client,
		host:   host,
		port:   port,
	}, nil
}

func (m *sender) Send(ctx context.Context, message *gomail.Msg) error {
	if message.GetMessageID() == "" {
		message.SetMessageID()
	}

	err := m.Dialer.DialAndSendWithContext(ctx, message)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}

	logger := slogutils.FromContext(ctx).
		With(slog.Group("smtp", "host", m.host, "port", m.port))
	logger.InfoContext(ctx,
		"Sent email via SMTP",
		slog.Group("email",
			"subject", getAndDecodeSubject(message),
			"to", message.GetToString(),
			"messageID", message.GetMessageID(),
		),
	)

	return nil
}

func getAndDecodeSubject(message *gomail.Msg) string {
	var subject string
	subjectHeader := message.GetGenHeader(gomail.HeaderSubject)
	if len(subjectHeader) > 0 {
		dec := new(mime.WordDecoder)
		//nolint:errcheck // Just for logging
		subject, _ = dec.DecodeHeader(subjectHeader[0])
	}
	return subject
}

var errInvalidTLSPolicy = std_errors.New("invalid TLS policy")

func getTLSPolicy(tlsPolicy string) (gomail.TLSPolicy, error) {
	switch tlsPolicy {
	case "opportunistic":
		return gomail.TLSOpportunistic, nil
	case "mandatory":
		return gomail.TLSMandatory, nil
	case "non":
		return gomail.NoTLS, nil
	}
	return -1, errInvalidTLSPolicy
}
