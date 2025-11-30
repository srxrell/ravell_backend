package utils

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
)

func SendOTPEmail(to, username, otp string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")

	// Детальное логирование конфигурации
	fmt.Printf("🔧 Email Configuration:\n")
	fmt.Printf("   Host: %s\n", smtpHost)
	fmt.Printf("   Port: %s\n", smtpPort)
	fmt.Printf("   User: %s\n", smtpUser)
	fmt.Printf("   From: %s\n", fromEmail)
	fmt.Printf("   Pass set: %v\n", smtpPass != "")
	fmt.Printf("   To: %s\n", to)

	if smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("email credentials not configured - check SMTP_USER and SMTP_PASS")
	}

	// Пробуем оба метода отправки
	err := sendWithStartTLS(to, username, otp, smtpHost, smtpPort, smtpUser, smtpPass, fromEmail)
	if err != nil {
		fmt.Printf("❌ StartTLS failed: %v\n", err)
		fmt.Printf("🔄 Trying standard SMTP...\n")
		return sendStandardSMTP(to, username, otp, smtpHost, smtpPort, smtpUser, smtpPass, fromEmail)
	}

	fmt.Printf("✅ Email sent successfully via StartTLS to: %s\n", to)
	return nil
}

func sendStandardSMTP(to, username, otp, host, port, user, pass, from string) error {
	auth := smtp.PlainAuth("", user, pass, host)

	subject := "Your OTP Verification Code"
	body := fmt.Sprintf(`
Hello %s,

Your OTP verification code is: %s

This code will expire in 15 minutes.

If you didn't request this code, please ignore this email.

Best regards,
Ravell Team
`, username, otp)

	message := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" + body)

	address := fmt.Sprintf("%s:%s", host, port)
	fmt.Printf("📧 Attempting standard SMTP to: %s\n", address)

	err := smtp.SendMail(address, auth, from, []string{to}, message)
	if err != nil {
		return fmt.Errorf("standard SMTP failed: %v", err)
	}

	return nil
}

func sendWithStartTLS(to, username, otp, host, port, user, pass, from string) error {
	fmt.Printf("🔐 Attempting StartTLS connection to %s:%s\n", host, port)

	// Connect to SMTP server
	client, err := smtp.Dial(host + ":" + port)
	if err != nil {
		return fmt.Errorf("dial failed: %v", err)
	}
	defer client.Close()

	// Send HELO/EHLO
	if err = client.Hello("localhost"); err != nil {
		return fmt.Errorf("hello failed: %v", err)
	}

	// Start TLS
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: host,
			InsecureSkipVerify: false,
		}
		if err = client.StartTLS(tlsConfig); err != nil {
			return fmt.Errorf("starttls failed: %v", err)
		}
		fmt.Printf("🔒 TLS connection established\n")
	}

	// Authenticate
	auth := smtp.PlainAuth("", user, pass, host)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("auth failed: %v", err)
	}
	fmt.Printf("✅ Authentication successful\n")

	// Set sender and recipient
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("mail from failed: %v", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("rcpt to failed: %v", err)
	}

	// Send email body
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("data failed: %v", err)
	}

	subject := "Your OTP Verification Code"
	body := fmt.Sprintf("Hello %s,\r\n\r\nYour OTP verification code is: %s\r\n\r\nThis code will expire in 15 minutes.\r\n\r\nBest regards,\r\nRavell Team", username, otp)
	
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s", to, subject, body)

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("write failed: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close failed: %v", err)
	}

	client.Quit()
	return nil
}