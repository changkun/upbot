// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"log"
	"net/smtp"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

var mg *mailgun.MailgunImpl

func init() {
	mg = mailgun.NewMailgun(
		conf.Sender.Mailgun.Domain, conf.Sender.Mailgun.APIKey)
	mg.SetAPIBase(conf.Sender.Mailgun.APIBase)
}

func sendMailgunEmail(ctx context.Context, subject, body string, notifiers []string) error {
	for _, recipient := range notifiers {
		r := recipient
		go func() {
			msg := mg.NewMessage(conf.Sender.Mailgun.Email, subject, body, r)

			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()

			_, _, err := mg.Send(ctx, msg)
			if err != nil {
				log.Printf("failed to send email: %v", err)
			}
		}()
	}
	return nil
}

func sendEmail(ctx context.Context, subject, body string, notifiers []string) error {
	if conf.Sender.Mailgun.Enable {
		return sendMailgunEmail(ctx, subject, body, notifiers)
	}

	err := smtp.SendMail(
		conf.Sender.Custom.SMTPHost+":"+conf.Sender.Custom.SMTPPort,
		smtp.PlainAuth("",
			conf.Sender.Custom.Username,
			conf.Sender.Custom.Passcode,
			conf.Sender.Custom.SMTPHost),
		conf.Sender.Custom.Email, notifiers,
		// rfc822format, see:
		// https://docs.microsoft.com/en-us/previous-versions/office/developer/exchange-server-2010/aa493918(v=exchg.140)
		[]byte(fmt.Sprintf("Subject: %s\r\nFrom: %s <%s>\r\n%s",
			// Content-Type: text/plain; charset=utf-8; format=flowed
			// Content-Transfer-Encoding: 7bit
			// Content-Language: en-US
			subject,
			"upbot", conf.Sender.Custom.Email,
			body,
		)))
	if err != nil {
		return err
	}
	return nil
}
