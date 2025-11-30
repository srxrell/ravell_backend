package utils

import (
	"fmt"
	"ravell_backend/config"

	"gopkg.in/gomail.v2"
)

func SendOTPEmail(email, username, otpCode string) error {
	cfg := config.LoadConfig()
	
	// Если SMTP не настроен, логируем OTP (для разработки)
	if cfg.SMTPUser == "" || cfg.SMTPPass == "" {
		fmt.Printf("📧 [DEV MODE] OTP for %s (%s): %s\n", username, email, otpCode)
		return nil
	}
	
	m := gomail.NewMessage()
	m.SetHeader("From", cfg.FromEmail)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "Код подтверждения для Stories App")
	
	body := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; background-color: #f4f4f4; padding: 20px;">
			<div style="background: white; padding: 30px; border-radius: 10px; max-width: 500px; margin: 0 auto;">
				<h2>Подтверждение регистрации</h2>
				<p>Здравствуйте, <strong>%s</strong>!</p>
				<p>Для завершения регистрации в Stories App используйте следующий код подтверждения:</p>
				<div style="font-size: 32px; font-weight: bold; color: #2563eb; text-align: center; margin: 20px 0;">%s</div>
				<p>Код действителен в течение 15 минут.</p>
				<div style="margin-top: 20px; font-size: 12px; color: #666;">
					<p>Если вы не регистрировались в нашем сервисе, проигнорируйте это письмо.</p>
				</div>
			</div>
		</div>
	`, username, otpCode)
	
	m.SetBody("text/html", body)
	
	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}
	
	return nil
}
