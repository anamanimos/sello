package services

import (
	"fmt"
	"strings"
)

// Logo URL dari sello.id
const SelloLogoURL = "https://sello.id/static/images/sello-logo.webp"

// Design system colors matching sello.id landing page
const (
	colorPrimary    = "#22c55e" // hsl(152, 76%, 45%)
	colorPrimaryAlt = "#4ade80" // hsl(145, 85%, 55%)
	colorBackground = "#f7faf8" // hsl(140, 20%, 97%)
	colorForeground = "#1a2e22" // hsl(150, 30%, 10%)
	colorMuted      = "#6b7c73" // hsl(150, 15%, 40%)
	colorAccentBg   = "#dcfce7" // hsl(145, 70%, 92%)
	colorBorder     = "#d4e5db" // hsl(150, 15%, 88%)
)

// Base email template matching sello.id landing page design
func baseEmailTemplate(title, content, footerText string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="id">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <title>%s</title>
    <!--[if mso]>
    <noscript>
        <xml>
            <o:OfficeDocumentSettings>
                <o:PixelsPerInch>96</o:PixelsPerInch>
            </o:OfficeDocumentSettings>
        </xml>
    </noscript>
    <![endif]-->
    <link href="https://fonts.googleapis.com/css2?family=Work+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
</head>
<body style="margin: 0; padding: 0; background-color: %s; font-family: 'Work Sans', -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;">
    <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: %s;">
        <tr>
            <td align="center" style="padding: 40px 20px;">
                <!-- Container -->
                <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="max-width: 560px;">
                    
                    <!-- Logo Header -->
                    <tr>
                        <td align="center" style="padding-bottom: 24px;">
                            <img src="%s" alt="Sello" width="100" style="display: block; max-width: 100px; height: auto;">
                        </td>
                    </tr>
                    
                    <!-- Main Card -->
                    <tr>
                        <td>
                            <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background: linear-gradient(135deg, %s 0%%, %s 100%%); border-radius: 12px 12px 0 0;">
                                <tr>
                                    <td style="padding: 24px 32px;">
                                        <h1 style="margin: 0; color: #ffffff; font-size: 22px; font-weight: 600; text-align: center;">%s</h1>
                                    </td>
                                </tr>
                            </table>
                            <table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="background-color: #ffffff; border-radius: 0 0 12px 12px; border: 1px solid %s; border-top: none;">
                                <tr>
                                    <td style="padding: 32px;">
                                        %s
                                    </td>
                                </tr>
                            </table>
                        </td>
                    </tr>
                    
                    <!-- Footer -->
                    <tr>
                        <td align="center" style="padding-top: 24px;">
                            <p style="margin: 0 0 8px 0; color: %s; font-size: 13px;">%s</p>
                            <p style="margin: 0; color: %s; font-size: 12px;">
                                © 2024 Sello. Semua hak dilindungi.<br>
                                <a href="https://sello.id" style="color: %s; text-decoration: none; font-weight: 500;">sello.id</a>
                            </p>
                        </td>
                    </tr>
                    
                </table>
            </td>
        </tr>
    </table>
</body>
</html>`, title, colorBackground, colorBackground, SelloLogoURL, colorPrimary, colorPrimaryAlt, title, colorBorder, content, colorMuted, footerText, colorMuted, colorPrimary)
}

// RenderOTPEmail generates HTML for OTP verification email
func RenderOTPEmail(code string) string {
	digits := strings.Split(code, "")
	digitBoxes := ""
	for _, d := range digits {
		digitBoxes += fmt.Sprintf(`<td style="width: 42px; height: 50px; background: linear-gradient(135deg, %s 0%%, %s 100%%); border-radius: 8px; text-align: center; vertical-align: middle;">
			<span style="color: #ffffff; font-size: 24px; font-weight: 700;">%s</span>
		</td>
		<td style="width: 6px;"></td>`, colorPrimary, colorPrimaryAlt, d)
	}

	content := fmt.Sprintf(`
		<p style="margin: 0 0 20px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Hai! 👋
		</p>
		<p style="margin: 0 0 24px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Gunakan kode verifikasi berikut untuk melanjutkan:
		</p>
		
		<!-- OTP Code Display -->
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 24px;">
			<tr>
				<td align="center">
					<table role="presentation" cellspacing="0" cellpadding="0" border="0">
						<tr>
							%s
						</tr>
					</table>
				</td>
			</tr>
		</table>
		
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 20px;">
			<tr>
				<td align="center" style="background-color: %s; border-radius: 8px; padding: 12px 16px;">
					<p style="margin: 0; color: %s; font-size: 13px;">
						⏱️ Kode ini berlaku selama <strong>15 menit</strong>
					</p>
				</td>
			</tr>
		</table>
		
		<hr style="border: none; border-top: 1px solid %s; margin: 20px 0;">
		
		<p style="margin: 0; color: %s; font-size: 12px; text-align: center;">
			🔒 Jangan bagikan kode ini kepada siapapun.<br>
			Tim Sello tidak akan pernah meminta kode verifikasi Anda.
		</p>
	`, colorForeground, colorForeground, digitBoxes, colorAccentBg, colorForeground, colorBorder, colorMuted)

	return baseEmailTemplate("Kode Verifikasi", content, "Email ini dikirim secara otomatis. Mohon jangan membalas email ini.")
}

// RenderPasswordResetEmail generates HTML for password reset email
func RenderPasswordResetEmail(resetLink string) string {
	content := fmt.Sprintf(`
		<p style="margin: 0 0 16px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Hai! 👋
		</p>
		<p style="margin: 0 0 24px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Kami menerima permintaan untuk mengatur ulang kata sandi akun Sello Anda.
		</p>
		
		<!-- Reset Button -->
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 24px;">
			<tr>
				<td align="center">
					<a href="%s" style="display: inline-block; background: linear-gradient(135deg, %s 0%%, %s 100%%); color: #ffffff; text-decoration: none; font-size: 15px; font-weight: 600; padding: 14px 32px; border-radius: 8px;">
						🔐 Atur Ulang Kata Sandi
					</a>
				</td>
			</tr>
		</table>
		
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 20px;">
			<tr>
				<td align="center" style="background-color: %s; border-radius: 8px; padding: 12px 16px;">
					<p style="margin: 0; color: %s; font-size: 13px;">
						⏱️ Link ini berlaku selama <strong>1 jam</strong>
					</p>
				</td>
			</tr>
		</table>
		
		<hr style="border: none; border-top: 1px solid %s; margin: 20px 0;">
		
		<p style="margin: 0 0 12px 0; color: %s; font-size: 12px; text-align: center;">
			Jika tombol tidak berfungsi, salin dan tempel link berikut:
		</p>
		<p style="margin: 0 0 20px 0; color: %s; font-size: 11px; text-align: center; word-break: break-all;">
			%s
		</p>
		
		<p style="margin: 0; color: %s; font-size: 12px; text-align: center;">
			🔒 Jika Anda tidak meminta reset kata sandi, abaikan email ini.
		</p>
	`, colorForeground, colorForeground, resetLink, colorPrimary, colorPrimaryAlt, colorAccentBg, colorForeground, colorBorder, colorMuted, colorPrimary, resetLink, colorMuted)

	return baseEmailTemplate("Atur Ulang Kata Sandi", content, "Email ini dikirim secara otomatis. Mohon jangan membalas email ini.")
}

// RenderMagicLinkEmail generates HTML for magic link login email
func RenderMagicLinkEmail(loginLink string) string {
	content := fmt.Sprintf(`
		<p style="margin: 0 0 16px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Hai! 👋
		</p>
		<p style="margin: 0 0 24px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Klik tombol di bawah untuk masuk ke akun Sello Anda secara instan.
		</p>
		
		<!-- Login Button -->
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 24px;">
			<tr>
				<td align="center">
					<a href="%s" style="display: inline-block; background: linear-gradient(135deg, %s 0%%, %s 100%%); color: #ffffff; text-decoration: none; font-size: 15px; font-weight: 600; padding: 14px 32px; border-radius: 8px;">
						✨ Masuk ke Sello
					</a>
				</td>
			</tr>
		</table>
		
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 20px;">
			<tr>
				<td align="center" style="background-color: %s; border-radius: 8px; padding: 12px 16px;">
					<p style="margin: 0; color: %s; font-size: 13px;">
						⏱️ Link ini berlaku selama <strong>15 menit</strong>
					</p>
				</td>
			</tr>
		</table>
		
		<hr style="border: none; border-top: 1px solid %s; margin: 20px 0;">
		
		<p style="margin: 0 0 12px 0; color: %s; font-size: 12px; text-align: center;">
			Jika tombol tidak berfungsi, salin dan tempel link berikut:
		</p>
		<p style="margin: 0 0 20px 0; color: %s; font-size: 11px; text-align: center; word-break: break-all;">
			%s
		</p>
		
		<p style="margin: 0; color: %s; font-size: 12px; text-align: center;">
			🔒 Jika Anda tidak meminta link ini, abaikan email ini.
		</p>
	`, colorForeground, colorForeground, loginLink, colorPrimary, colorPrimaryAlt, colorAccentBg, colorForeground, colorBorder, colorMuted, colorPrimary, loginLink, colorMuted)

	return baseEmailTemplate("Masuk ke Akun Anda", content, "Email ini dikirim secara otomatis. Mohon jangan membalas email ini.")
}

// RenderWelcomeEmail generates HTML for welcome email
func RenderWelcomeEmail(userName string) string {
	content := fmt.Sprintf(`
		<p style="margin: 0 0 16px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Hai <strong>%s</strong>! 🎉
		</p>
		<p style="margin: 0 0 24px 0; color: %s; font-size: 15px; line-height: 1.6; text-align: center;">
			Selamat datang di Sello! Akun Anda telah berhasil dibuat.
		</p>
		
		<!-- Features Box -->
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 24px; background-color: %s; border-radius: 8px;">
			<tr>
				<td style="padding: 20px;">
					<p style="margin: 0 0 12px 0; color: %s; font-size: 14px; font-weight: 600;">Yang bisa Anda lakukan:</p>
					<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%">
						<tr>
							<td style="padding: 6px 0; color: %s; font-size: 13px;">✅ Buat katalog produk yang menarik</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: %s; font-size: 13px;">✅ Terima pesanan via WhatsApp</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: %s; font-size: 13px;">✅ Pantau analitik penjualan</td>
						</tr>
						<tr>
							<td style="padding: 6px 0; color: %s; font-size: 13px;">✅ Kustomisasi tampilan toko</td>
						</tr>
					</table>
				</td>
			</tr>
		</table>
		
		<!-- CTA Button -->
		<table role="presentation" cellspacing="0" cellpadding="0" border="0" width="100%%" style="margin-bottom: 20px;">
			<tr>
				<td align="center">
					<a href="https://sello.id/dashboard" style="display: inline-block; background: linear-gradient(135deg, %s 0%%, %s 100%%); color: #ffffff; text-decoration: none; font-size: 15px; font-weight: 600; padding: 14px 32px; border-radius: 8px;">
						🚀 Mulai Jualan Sekarang
					</a>
				</td>
			</tr>
		</table>
		
		<hr style="border: none; border-top: 1px solid %s; margin: 20px 0;">
		
		<p style="margin: 0; color: %s; font-size: 12px; text-align: center;">
			Butuh bantuan? Hubungi kami di <a href="mailto:support@sello.id" style="color: %s; text-decoration: none; font-weight: 500;">support@sello.id</a>
		</p>
	`, colorForeground, userName, colorForeground, colorAccentBg, colorForeground, colorForeground, colorForeground, colorForeground, colorForeground, colorPrimary, colorPrimaryAlt, colorBorder, colorMuted, colorPrimary)

	return baseEmailTemplate("Selamat Datang di Sello!", content, "Anda menerima email ini karena mendaftar di sello.id")
}
