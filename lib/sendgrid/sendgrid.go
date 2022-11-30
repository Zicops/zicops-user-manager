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

// SendJoinEmail ......
func (sgc *ClientSendGrid) SendJoinEmail(
	email string,
	verifyURL string, username string) error {
	mailSetup := mail.NewV3Mail()
	from := mail.NewEmail("Zicops Admin", "no_reply@zicops.com")
	to := mail.NewEmail(username, email)
	mailSetup.SetFrom(from)
	mailSetup.SetTemplateID("d-fec6618de32244f092ee0d45b405501f")
	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("b_url", verifyURL)
	mailSetup.AddPersonalizations(p)
	request := sendgrid.GetRequest("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", "/v3/mail/send", "https://api.sendgrid.com")
	request.Method = "POST"
	var Body = mail.GetRequestBody(mailSetup)
	request.Body = Body
	_, err := sendgrid.API(request)
	if err != nil {
		return err
	}
	return nil
}

// SendPasswordResetEmail ......
func (sgc *ClientSendGrid) SendPasswordResetEmail(
	email string,
	verifyURL string, username string) error {
	from := mail.NewEmail("Zicops Admin", "noreply@zicops.com")
	subject := "Link to reset your password"
	to := mail.NewEmail(username, email)
	plainTextContent := "Follow the link to reset your password: " + verifyURL
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	_, err := sgc.client.Send(message)
	if err != nil {
		return err
	}
	return nil
}

// SendInviteToLspEmail ......
func (sgc *ClientSendGrid) SendInviteToLspEmail(
	email string,
	verifyURL string, username string) error {
	from := mail.NewEmail("Zicops Admin", "noreply@zicops.com")
	subject := "You have been invited to a learning space"
	to := mail.NewEmail(username, email)
	plainTextContent := "Follow the link to view all your learning spaces: " + verifyURL
	message := mail.NewSingleEmail(from, subject, to, plainTextContent, "")
	_, err := sgc.client.Send(message)
	if err != nil {
		return err
	}
	return nil
}
