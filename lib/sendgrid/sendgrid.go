package sendgrid

import (
	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// ClientSendGrid ....
type ClientSendGrid struct {
	client *sendgrid.Client
}

// NewSendGridClient return new database action
func NewSendGridClient() *ClientSendGrid {
	return &ClientSendGrid{client: nil}
}

// InitializeSendGridClient ....
func (sgc *ClientSendGrid) InitializeSendGridClient() error {
	sgAPIKey := os.Getenv("SENDGRID_API_KEY")
	if sgAPIKey == "" {
		sgAPIKey = "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"
	}
	client := sendgrid.NewSendClient(sgAPIKey)
	sgc.client = client
	return nil
}

// SendPasswordResetEmail ......
func (sgc *ClientSendGrid) SendJoinEmail(
	email string,
	verifyURL string, username string) error {
	from := mail.NewEmail("Zicops Admin", "noreply@zicops.com")
	subject := "You are invited to join Zicops"
	to := mail.NewEmail(username, email)
	plainTextContent := "Follow the link to reset your password: " + verifyURL
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	_, err := sgc.client.Send(message)
	if err != nil {
		return err
	}
	return nil
}
