package smtp

import (
	"context"

	"github.com/friendsofgo/errors"
	gomail "github.com/wneessen/go-mail"

	"myvendor.mytld/myproject/backend/mail"
)

type sender struct {
	Dialer *gomail.Client
}

var _ mail.Sender = new(sender)

func NewSender(host string, port int, username, password, tlsPolicy string) (mail.Sender, error) {
	policy, err := getTLSPolicy(tlsPolicy)
	if err != nil {
		return nil, err
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
	}

	return &sender{
		Dialer: client,
	}, nil
}

func (m *sender) Send(ctx context.Context, message *gomail.Msg) error {
	err := m.Dialer.DialAndSendWithContext(ctx, message)
	if err != nil {
		return errors.Wrap(err, "sending message")
	}

	return nil
}

func getTLSPolicy(tlsPolicy string) (gomail.TLSPolicy, error) {
	switch tlsPolicy {
	case "opportunistic":
		return gomail.TLSOpportunistic, nil
	case "mandatory":
		return gomail.TLSMandatory, nil
	case "non":
		return gomail.NoTLS, nil
	}
	return -1, errors.New("invalid TLS policy")
}
