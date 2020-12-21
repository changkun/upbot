// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"
)

type status string

const (
	statusStart status = "START"
	statusUp           = "UP"
	statusDown         = "DOWN"
)

type monitor struct {
	Name         string   `yaml:"name"`
	URL          string   `yaml:"url"`
	ExpectedCode []int    `yaml:"expected_code"`
	Notifiers    []string `yaml:"notifiers"`
	Interval     int      `yaml:"interval"`

	mu     sync.Mutex `yaml:"-"`
	Code   int        `yaml:"-"`
	Status status     `yaml:"-"`
	Since  time.Time  `yaml:"-"`
}

// CheckHealth checks the monitor's up health. If the targed does not
// match expected status code, it will send emails to configured recipients.
func (m *monitor) CheckHealth(ctx context.Context, notifiers []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(m.URL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	m.Code = resp.StatusCode

	msg := fmt.Sprintf("checking %s (%s)", m.Name, m.URL)
	defer func() {
		log.Println(msg)
	}()

	// check expected response code
	match := false
	for _, code := range m.ExpectedCode {
		if resp.StatusCode == code {
			match = true
			break
		}
	}

	if match {
		msg = fmt.Sprintf("%s: OK (%d)", msg, resp.StatusCode)
		switch m.Status { // old status
		case statusStart: // was start, now up
			m.Status = statusUp
			m.Since = time.Now().UTC()
		case statusDown: // was down, now up
			m.Status = statusUp
			m.notify(ctx, statusDown, statusUp, notifiers)
			msg = fmt.Sprintf("%s, downtime: %s", msg, time.Since(m.Since))

			m.Since = time.Now().UTC()
		case statusUp: // was up, still up => nothing to do
		default:
			log.Println("unknown status")
		}
		return nil
	}

	msg = fmt.Sprintf("%s: DOWN (%d)", msg, resp.StatusCode)
	switch m.Status {
	case statusStart:
		m.Status = statusDown
		m.Since = time.Now().UTC()
	case statusDown: // was down, still down ==> nothing to do
	case statusUp: // was up, now down
		m.Status = statusDown
		m.notify(ctx, statusUp, statusDown, notifiers)
		msg = fmt.Sprintf("%s, uptime: %s", msg, time.Since(m.Since))

		m.Since = time.Now().UTC()
	default:
		log.Println("unknown status")
	}
	return errors.New("unexpected status")
}

func (m *monitor) notify(ctx context.Context, old, new status, notifiers []string) {
	log.Println("notify notifiers...", notifiers)

	const (
		subjectTmpl = "upbot: %s is %s (%s)"
		bodyTmpl    = `Hi,

The monitor %s (%s) is currently %s (HTTP %d - %s) (It was %s for %s).

Event timestamp: %s

Best regards,
changkun/upbot
`
	)
	var (
		subject string
		body    string
	)
	switch new {
	case statusStart:
		return
	case statusUp, statusDown:
		subject = fmt.Sprintf(subjectTmpl, m.Name, new, m.URL)
		body = fmt.Sprintf(
			bodyTmpl,
			m.Name, m.URL,
			new, m.Code, http.StatusText(m.Code),
			old, time.Since(m.Since),
			time.Now().UTC().Format(time.RFC1123))
	default:
		return
	}

	err := sendEmail(ctx, subject, body, notifiers)
	if err != nil {
		log.Printf("send email failed: %v", err)
	}
}
