package mailer

import "net/smtp"

func SendEmail(host string, port string, login string, pass string, toEmails []string, message string) error {
	body := []byte(message)

	auth := smtp.PlainAuth("", login, pass, host)

	err := smtp.SendMail(host+":"+port, auth, login, toEmails, body)

	return err
}
