package mail

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/mail"
	"net/smtp"

	config "github.com/a-castellano/home-ip-notifier/config"
)

func SendEmail(config *config.Config, messageToSend string) error {

	fromMail := fmt.Sprintf("%s@%s", config.MailFrom, config.MailDomain)
	from := mail.Address{Name: "", Address: fromMail}
	to := mail.Address{Name: "", Address: config.Destination}
	subj := "Home IP has changed"

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + messageToSend

	// Connect to the SMTP Server
	servername := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	host, _, _ := net.SplitHostPort(servername)

	auth := smtp.PlainAuth("", config.SMTPName, config.SMTPPassword, host)

	// TLS config
	tlsconfig := &tls.Config{
		InsecureSkipVerify: config.SMTPTLSValidation,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		return err
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	c.Quit()
	log.Println("Mail sent successfully")

	return nil
}

// SendEmailWithoutTLS envía emails usando SMTP sin TLS (puerto 25)
func SendEmailWithoutTLS(config *config.Config, messageToSend string) error {

	fromMail := fmt.Sprintf("%s@%s", config.MailFrom, config.MailDomain)
	from := mail.Address{Name: "", Address: fromMail}
	to := mail.Address{Name: "", Address: config.Destination}
	subj := "Home IP has changed"

	// Setup headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Setup message
	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + messageToSend

	// Connect to the SMTP Server without TLS
	servername := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	// Conectar directamente sin TLS
	c, err := smtp.Dial(servername)
	if err != nil {
		return err
	}
	defer c.Quit()

	// Auth (si es necesario)
	if config.SMTPName != "" && config.SMTPPassword != "" {
		auth := smtp.PlainAuth("", config.SMTPName, config.SMTPPassword, config.SMTPHost)
		if err = c.Auth(auth); err != nil {
			return err
		}
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Data
	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	err = w.Close()
	if err != nil {
		return err
	}

	log.Println("Mail sent successfully without TLS")

	return nil
}
