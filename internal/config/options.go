package config

import (
	"errors"
	"flag"
	"net/url"
	"strconv"
	"strings"
)

var opt options

type options struct {
	ServerAddress serverAddress
	BaseURL       BaseURL
}

func (o *options) GetServer() string {
	return opt.ServerAddress.String()
}

func (o *options) GetBaseURL() string {
	return opt.BaseURL.String()
}

type serverAddress struct {
	Host string
	Port int
}

func (s *serverAddress) String() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (s *serverAddress) Set(flag string) error {

	ss := strings.Split(flag, ":")

	if len(ss) != 2 {
		return errors.New("invalid server argument")
	}
	sp, err := strconv.Atoi(ss[1])
	if err != nil {
		return err
	}

	s.Host = ss[0]
	s.Port = sp

	return nil
}

type BaseURL struct {
	Host string
	Port int
}

func (b *BaseURL) String() string {
	return "http://" + b.Host + ":" + strconv.Itoa(b.Port)
}

func (b *BaseURL) Set(flag string) error {

	bu, err := url.Parse(flag)

	if err != nil {
		return err
	}
	bup, err := strconv.Atoi(bu.Port())
	if err != nil {
		return err
	}

	b.Host = bu.Hostname()
	b.Port = bup

	return nil
}

func init() {
	// default values
	opt.ServerAddress.Host = "localhost"
	opt.ServerAddress.Port = 8080
	opt.BaseURL.Host = "localhost"
	opt.BaseURL.Port = 8080

	_ = flag.Value(&opt.ServerAddress)
	_ = flag.Value(&opt.BaseURL)
	flag.Var(&opt.ServerAddress, "a", "Server address - host:port")
	flag.Var(&opt.BaseURL, "b", "Server ShortLink Base address - protocol://host:port")
	flag.Parse()
}

func GetOptions() *options {
	return &opt
}
