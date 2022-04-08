package srv_repeater

import (
	"fmt"
	"github.com/gazercloud/gazer_repeater/credentials"
	"github.com/gazercloud/gazer_repeater/logger"
	"net/smtp"
)

func SendEMail(subj string, to string, content string) error {
	logger.Println("[SendEMail]", to)

	//
	// ses-smtp-user.20210518-171317
	// "Statement": [{"Effect":"Allow","Action":"ses:SendRawEmail","Resource":"*"}]
	// ses-smtp-user.20210518-171317
	// username: AKIATM76G7CYOW6N5XOH
	// password: BKfJcaGIezSMo/+3kjEsw8cbaSgwOW2Fs1XuOwaxJmQf
	var emailAuth smtp.Auth

	emailHost := credentials.EmailSmtpServer
	emailFrom := credentials.EmailSmtpFrom
	emailPassword := credentials.EmailSmtpPassword
	emailPort := 587

	emailAuth = smtp.PlainAuth("", credentials.EmailSmtpUser, emailPassword, emailHost)

	emailBody := content

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	subject := "Subject: " + subj + "\n"
	toAddr := "To: " + to + "\n"
	msg := []byte(subject + toAddr + mime + "\n" + emailBody)
	addr := emailHost + ":" + fmt.Sprint(emailPort)
	logger.Println("[SendEMail]", "email to ", addr)

	if err := smtp.SendMail(addr, emailAuth, emailFrom, []string{to}, msg); err != nil {
		logger.Println("[SendEMail]", "[error]", to)
		return err
	}
	return nil
}
