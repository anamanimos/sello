package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"time"

	"sello/internal/database"
	"sello/internal/models"
)

// SendEmailOTP sends OTP via SMTP (supports both TLS and plain)
func SendEmailOTP(email, code string) error {
	var settings []models.SystemSetting
	database.GetDB().Where("key IN ?", []string{"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from"}).Find(&settings)

	config := make(map[string]string)
	for _, s := range settings {
		config[s.Key] = s.Value
	}

	if config["smtp_host"] == "" {
		return fmt.Errorf("SMTP settings not configured")
	}

	sender := config["smtp_from"]
	if sender == "" {
		sender = config["smtp_user"]
	}

	subject := "Kode Verifikasi - Sello"
	htmlBody := RenderOTPEmail(code)

	// Build MIME message with HTML content
	boundary := "----=_Part_0_" + fmt.Sprintf("%d", time.Now().UnixNano())
	msg := []byte("From: " + sender + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		"Kode verifikasi Sello Anda: " + code + "\n\nKode ini berlaku selama 15 menit.\nJangan bagikan kode ini kepada siapapun.\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		htmlBody + "\r\n" +
		"\r\n" +
		"--" + boundary + "--\r\n")

	addr := fmt.Sprintf("%s:%s", config["smtp_host"], config["smtp_port"])

	// Connect to SMTP server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[SMTP] Failed to connect to %s: %v", addr, err)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config["smtp_host"])
	if err != nil {
		log.Printf("[SMTP] Failed to create client: %v", err)
		return err
	}
	defer client.Close()

	// Authenticate if credentials provided
	if config["smtp_user"] != "" && config["smtp_pass"] != "" {
		auth := LoginAuth(config["smtp_user"], config["smtp_pass"])
		if err = client.Auth(auth); err != nil {
			log.Printf("[SMTP] Auth failed: %v", err)
			return err
		}
	}

	// Send email
	if err = client.Mail(sender); err != nil {
		log.Printf("[SMTP] MAIL FROM failed: %v", err)
		return err
	}
	if err = client.Rcpt(email); err != nil {
		log.Printf("[SMTP] RCPT TO failed: %v", err)
		return err
	}

	w, err := client.Data()
	if err != nil {
		log.Printf("[SMTP] DATA failed: %v", err)
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		log.Printf("[SMTP] Write failed: %v", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("[SMTP] Close failed: %v", err)
		return err
	}

	log.Printf("[SMTP] Email sent successfully to %s", email)
	return client.Quit()
}

// SendPasswordResetEmail sends a password reset link via SMTP
func SendPasswordResetEmail(email, resetLink string) error {
	return sendHTMLEmail(email, "Atur Ulang Kata Sandi - Sello", RenderPasswordResetEmail(resetLink))
}

// SendMagicLinkEmail sends a magic link login email via SMTP
func SendMagicLinkEmail(email, loginLink string) error {
	return sendHTMLEmail(email, "Masuk ke Akun Anda - Sello", RenderMagicLinkEmail(loginLink))
}

// SendWelcomeEmail sends a welcome email to new users
func SendWelcomeEmail(email, userName string) error {
	return sendHTMLEmail(email, "Selamat Datang di Sello! 🎉", RenderWelcomeEmail(userName))
}

// SendPasswordResetWhatsApp sends a password reset link via WhatsApp
func SendPasswordResetWhatsApp(phone, resetLink string) error {
	message := fmt.Sprintf(`🔐 *Reset Password Sello*

Hai! Kamu meminta untuk mengatur ulang password akun Sello kamu.

Klik link berikut untuk reset password:
%s

⏰ Link ini berlaku 24 jam.

Jika kamu tidak meminta reset password, abaikan pesan ini.

— Tim Sello`, resetLink)

	return sendWhatsAppMessage(phone, message)
}

// sendWhatsAppMessage is a helper to send WhatsApp messages via gateway
func sendWhatsAppMessage(phone, message string) error {
	// Get gateway settings
	var gatewaySetting models.SystemSetting
	gatewayURL := "http://localhost:8000"
	if err := database.GetDB().Where("key = ?", "whatsapp_gateway_url").First(&gatewaySetting).Error; err == nil {
		gatewayURL = gatewaySetting.Value
	}

	// Build request
	payload := map[string]string{
		"phone":   phone,
		"message": message,
	}

	jsonPayload, _ := json.Marshal(payload)

	resp, err := http.Post(gatewayURL+"/send/message", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to send WhatsApp: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WhatsApp gateway returned status %d", resp.StatusCode)
	}

	return nil
}
// sendHTMLEmail is a helper function to send HTML emails
func sendHTMLEmail(email, subject, htmlBody string) error {
	var settings []models.SystemSetting
	database.GetDB().Where("key IN ?", []string{"smtp_host", "smtp_port", "smtp_user", "smtp_pass", "smtp_from"}).Find(&settings)

	config := make(map[string]string)
	for _, s := range settings {
		config[s.Key] = s.Value
	}

	if config["smtp_host"] == "" {
		return fmt.Errorf("SMTP settings not configured")
	}

	sender := config["smtp_from"]
	if sender == "" {
		sender = config["smtp_user"]
	}

	// Build MIME message with HTML content
	boundary := "----=_Part_0_" + fmt.Sprintf("%d", time.Now().UnixNano())
	plainText := "Email ini memerlukan klien email yang mendukung HTML untuk ditampilkan dengan benar."
	
	msg := []byte("From: " + sender + "\r\n" +
		"To: " + email + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: multipart/alternative; boundary=\"" + boundary + "\"\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/plain; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		plainText + "\r\n" +
		"\r\n" +
		"--" + boundary + "\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		htmlBody + "\r\n" +
		"\r\n" +
		"--" + boundary + "--\r\n")

	addr := fmt.Sprintf("%s:%s", config["smtp_host"], config["smtp_port"])

	// Connect to SMTP server
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		log.Printf("[SMTP] Failed to connect to %s: %v", addr, err)
		return err
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, config["smtp_host"])
	if err != nil {
		log.Printf("[SMTP] Failed to create client: %v", err)
		return err
	}
	defer client.Close()

	// Authenticate if credentials provided
	if config["smtp_user"] != "" && config["smtp_pass"] != "" {
		auth := LoginAuth(config["smtp_user"], config["smtp_pass"])
		if err = client.Auth(auth); err != nil {
			log.Printf("[SMTP] Auth failed: %v", err)
			return err
		}
	}

	// Send email
	if err = client.Mail(sender); err != nil {
		log.Printf("[SMTP] MAIL FROM failed: %v", err)
		return err
	}
	if err = client.Rcpt(email); err != nil {
		log.Printf("[SMTP] RCPT TO failed: %v", err)
		return err
	}

	w, err := client.Data()
	if err != nil {
		log.Printf("[SMTP] DATA failed: %v", err)
		return err
	}

	_, err = w.Write(msg)
	if err != nil {
		log.Printf("[SMTP] Write failed: %v", err)
		return err
	}

	err = w.Close()
	if err != nil {
		log.Printf("[SMTP] Close failed: %v", err)
		return err
	}

	log.Printf("[SMTP] HTML Email sent successfully to %s", email)
	return client.Quit()
}

// LoginAuth implements the LOGIN authentication mechanism
type loginAuth struct {
	username, password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, nil
		}
	}
	return nil, nil
}

// SendWhatsappOTP sends OTP via Custom Gateway (go-whatsapp-web-multidevice)
func SendWhatsappOTP(phone, code string) error {
	var settings []models.SystemSetting
	database.GetDB().Where("key IN ?", []string{"wa_url", "wa_username", "wa_password"}).Find(&settings)

	config := make(map[string]string)
	for _, s := range settings {
		config[s.Key] = s.Value
	}

	if config["wa_url"] == "" {
		return fmt.Errorf("WhatsApp Gateway settings not configured")
	}

	// Format Request Payload
	// Phone: 08x -> 628x@s.whatsapp.net
	formattedPhone := phone
	if strings.HasPrefix(phone, "08") {
		formattedPhone = "62" + phone[1:]
	}
	if !strings.HasSuffix(formattedPhone, "@s.whatsapp.net") {
		formattedPhone = formattedPhone + "@s.whatsapp.net"
	}

	url := fmt.Sprintf("%s/send/message", strings.TrimRight(config["wa_url"], "/"))
	message := fmt.Sprintf("Kode OTP Sello Anda: *%s*\n\nJangan berikan kode ini kepada siapapun.", code)

	payload := map[string]interface{}{
		"phone":        formattedPhone,
		"message":      message,
		"is_forwarded": false,
		"duration":     0, // Valid: 0 (no expiry), 86400 (24h), 604800 (7d), 7776000 (90d)
	}
	
	jsonPayload, _ := json.Marshal(payload)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	// Basic Auth
	req.SetBasicAuth(config["wa_username"], config["wa_password"])
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Failed to send WA:", err)
		return err
	}
	defer resp.Body.Close()

	// Read response body for debugging
	var respBody bytes.Buffer
	respBody.ReadFrom(resp.Body)
	log.Printf("[WhatsApp] Request to %s, Status: %d, Response: %s", url, resp.StatusCode, respBody.String())

	if resp.StatusCode >= 400 {
		return fmt.Errorf("gateway responded with status: %d - %s", resp.StatusCode, respBody.String())
	}

	return nil
}

// GetDefaultVerificationMethod returns "email" or "whatsapp"
func GetDefaultVerificationMethod() string {
	var setting models.SystemSetting
	if err := database.GetDB().Where("key = ?", "verification_default").First(&setting).Error; err != nil {
		return "email" // Default fallback
	}
	return setting.Value
}
