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
	verifyURL string, username string, orgName string, lspName string) error {
	mailSetup := mail.NewV3Mail()
	from := mail.NewEmail("Zicops Admin", "no_reply@zicops.com")
	to := mail.NewEmail(username, email)
	mailSetup.SetFrom(from)
	mailSetup.SetTemplateID("d-7aa878e4e4e346d3bfd58561e6a59ef8")
	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("login_url", verifyURL)
	p.SetDynamicTemplateData("organization_name", orgName)
	p.SetDynamicTemplateData("lsp_name", lspName)
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
	mailSetup := mail.NewV3Mail()
	from := mail.NewEmail("Zicops Admin", "no_reply@zicops.com")
	to := mail.NewEmail(username, email)
	mailSetup.SetFrom(from)
	mailSetup.SetTemplateID("d-1d5dbc9466574b148cc9add368a37b26")
	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("reset_pwd_url", verifyURL)
	p.SetDynamicTemplateData("username", email)
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

// SendInviteToLspEmail ......
func (sgc *ClientSendGrid) SendInviteToLspEmail(
	email string,
	verifyURL string,
	orgName string, lspName string) error {
	mailSetup := mail.NewV3Mail()
	from := mail.NewEmail("Zicops Admin", "no_reply@zicops.com")
	to := mail.NewEmail(email, email)
	mailSetup.SetFrom(from)
	mailSetup.SetTemplateID("d-2b155eb19b1e4e49b4c136aa8b2307a4")
	p := mail.NewPersonalization()
	p.AddTos(to)
	p.SetDynamicTemplateData("organization_name", orgName)
	p.SetDynamicTemplateData("lsp_name", lspName)
	p.SetDynamicTemplateData("link", verifyURL)
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
