package mail

import (
	"bytes"
	"fmt"
	"html/template"
	"net/smtp"
)

type Mailer struct {
	User string
	Pwd  string
	Host string
	Port string
}

func NewMailer(user, pwd, host, port string) *Mailer {
	return &Mailer{
		User: user,
		Pwd:  pwd,
		Host: host,
		Port: port,
	}
}

type SignupData struct {

	Name string
	Email string
	Token string
}

type ActivateOrResetData struct {
	Name  string
	Email string `json:"email" validate:"required,email"`
	Token string `json:"token" validate:"required"`
}

func (m *Mailer) SendWelcomeEmail(data SignupData) error {

	tmpl, err := template.New("email").Parse(welcome_template)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	recipients := []string{data.Email}

	subject := "Welcome to Poult - Account Activation Required"
	msg := fmt.Sprintf("Subject: %s\nTo: %s\nContent-Type: text/html\n\n%s", subject, data.Email, emailBody.String())

	auth := smtp.PlainAuth("", m.User, m.Pwd, m.Host)
	err = smtp.SendMail(m.Host+":"+m.Port, auth, m.User, recipients, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (m *Mailer) SendActivateEmail(data ActivateOrResetData) error {

	tmpl, err := template.New("email").Parse(activate_template)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	recipients := []string{data.Email}

	subject := "Poult - Account Activation Required"
	msg := fmt.Sprintf("Subject: %s\nTo: %s\nContent-Type: text/html\n\n%s", subject, data.Email, emailBody.String())

	auth := smtp.PlainAuth("", m.User, m.Pwd, m.Host)
	err = smtp.SendMail(m.Host+":"+m.Port, auth, m.User, recipients, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (m *Mailer) SendPasswordResetEmail(data ActivateOrResetData) error {

	tmpl, err := template.New("email").Parse(pwdreset_template)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	recipients := []string{data.Email}

	subject := "Password Reset Request for Poult"
	msg := fmt.Sprintf("Subject: %s\nTo: %s\nContent-Type: text/html\n\n%s", subject, data.Email, emailBody.String())

	auth := smtp.PlainAuth("", m.User, m.Pwd, m.Host)
	err = smtp.SendMail(m.Host+":"+m.Port, auth, m.User, recipients, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}

func (m *Mailer) SendResetCompletedEmail(data ActivateOrResetData) error {

	tmpl, err := template.New("email").Parse(completedreset_template)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %v", err)
	}

	var emailBody bytes.Buffer
	if err := tmpl.Execute(&emailBody, data); err != nil {
		return fmt.Errorf("failed to execute email template: %v", err)
	}

	recipients := []string{data.Email}

	subject := "Password Changed for Poult"
	msg := fmt.Sprintf("Subject: %s\nTo: %s\nContent-Type: text/html\n\n%s", subject, data.Email, emailBody.String())

	auth := smtp.PlainAuth("", m.User, m.Pwd, m.Host)
	err = smtp.SendMail(m.Host+":"+m.Port, auth, m.User, recipients, []byte(msg))
	if err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
