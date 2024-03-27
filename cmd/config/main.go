package config

import (
	"net/url"
	"strconv"
	"strings"
)

var opt options

type options struct {
	ServerHost    string
	ServerPort    int
	ShortLinkHost string
	ShortLinkPort int
}

func (o *options) init() bool {
	return o.ServerHost != "" && o.ServerPort != 0 && o.ShortLinkHost != "" && o.ShortLinkPort != 0
}

func (o *options) GetServer() string {
	if !o.init() {
		panic("options not initialised")
	}
	return o.ServerHost + ":" + strconv.Itoa(o.ServerPort)
}

func (o *options) GetShortLinkServer() string {
	if !o.init() {
		panic("options not initialised")
	}
	return "http://" + o.ShortLinkHost + ":" + strconv.Itoa(o.ShortLinkPort)
}

// server must contains host:port
func (o *options) SetServer(server string) {

	s := strings.Split(server, ":")

	if len(s) != 2 {
		panic("invalid server argument")
	}
	sp, err := strconv.Atoi(s[1])
	if err != nil {
		panic("invalid server port in argument")
	}

	o.ServerHost = s[0]
	o.ServerPort = sp
}

// shortLinkHost must contains host:port
func (o *options) SetShortLinkServer(shortLinkServer string) {

	sls, err := url.Parse(shortLinkServer)

	if err != nil {
		panic("invalid shortLinkServer argument")
	}
	slp, err := strconv.Atoi(sls.Port())
	if err != nil {
		panic("invalid shortLinkServer port in argument")
	}

	o.ShortLinkHost = sls.Hostname()
	o.ShortLinkPort = slp
}

// constructor for options
func newOptions() *options {
	return &options{
		ServerHost:    "localhost",
		ServerPort:    8080,
		ShortLinkHost: "localhost",
		ShortLinkPort: 8080,
	}
}

func GetOptions() options {
	if !opt.init() {
		opt = *newOptions()
	}

	return opt
}
