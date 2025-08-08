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

// SendEmail sends an email notification about IP changes using SMTP with TLS.
// It constructs the email message with proper headers and sends it via the configured SMTP server.
//
// Parameters:
//   - config: Application configuration containing SMTP settings
//   - messageToSend: The message body to include in the email
//
// Returns an error if the email could not be sent due to connection, authentication, or other issues.
func SendEmail(config *config.Config, messageToSend string) error {

	// Construct sender email address from username and domain
	fromMail := fmt.Sprintf("%s@%s", config.MailFrom, config.MailDomain)
	from := mail.Address{Name: "", Address: fromMail}
	to := mail.Address{Name: "", Address: config.Destination}
	subj := "Home IP has changed"

	// Setup email headers
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj

	// Construct the complete email message with headers and body
	var message string
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + messageToSend

	// Prepare SMTP server connection details
	servername := fmt.Sprintf("%s:%d", config.SMTPHost, config.SMTPPort)

	host, _, _ := net.SplitHostPort(servername)

	// Create SMTP authentication
	auth := smtp.PlainAuth("", config.SMTPName, config.SMTPPassword, host)

	// Configure TLS settings
	tlsconfig := &tls.Config{
		InsecureSkipVerify: !config.SMTPTLSValidation, // Skip certificate validation if configured
		ServerName:         host,
	}

	// Establish TLS connection to SMTP server
	// This is required for SMTP servers running on port 465 that require SSL from the start
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		return err
	}

	// Create SMTP client from the TLS connection
	c, err := smtp.NewClient(conn, host)
	if err != nil {
		return err
	}

	// Authenticate with the SMTP server
	if err = c.Auth(auth); err != nil {
		return err
	}

	// Set the sender address
	if err = c.Mail(from.Address); err != nil {
		return err
	}

	// Set the recipient address
	if err = c.Rcpt(to.Address); err != nil {
		return err
	}

	// Send the email data
	w, err := c.Data()
	if err != nil {
		return err
	}

	// Write the email message
	_, err = w.Write([]byte(message))
	if err != nil {
		return err
	}

	// Close the data writer
	err = w.Close()
	if err != nil {
		return err
	}

	// Close the SMTP connection
	c.Quit()
	log.Println("Mail sent successfully")

	return nil
}
