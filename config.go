// Copyright 2020 Changkun Ou. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	Addr   string `yaml:"addr"`
	Sender struct {
		Mailgun struct {
			Email   string `yaml:"email"`
			Enable  bool   `yaml:"enable"`
			Domain  string `yaml:"domain"`
			APIKey  string `yaml:"apikey"`
			APIBase string `yaml:"apibase"`
		} `yaml:"mailgun"`
		Custom struct {
			SMTPHost string `yaml:"smtp_host"`
			SMTPPort string `yaml:"smtp_port"`
			Email    string `yaml:"email"`
			Username string `yaml:"username"`
			Passcode string `yaml:"passcode"`
		} `yaml:"custom"`
	} `yaml:"sender"`
	Default struct {
		Notifiers []string `yaml:"notifiers"`
		Interval  int      `yaml:"interval"`
	} `yaml:"default"`
	Monitors []*monitor `yaml:"monitors"`
}

func (c *config) parse() {
	f := os.Getenv("UPBOT_CONF")
	d, err := os.ReadFile(f)
	if err != nil {
		d, err = os.ReadFile("./configs/config.yml")
		if err != nil {
			log.Fatalf("cannot read config, err: %v\n", err)
		}
	}
	err = yaml.Unmarshal(d, c)
	if err != nil {
		log.Fatalf("cannot parse config, err: %v\n", err)
	}
}

var conf config

func init() {
	conf.parse()
}
