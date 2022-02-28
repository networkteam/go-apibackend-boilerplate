package mail

import "gopkg.in/gomail.v2"

type Sender interface {
	Send(message *gomail.Message) error
}
