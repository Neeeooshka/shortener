package config

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

type Options struct {
	ServerAddress ServerAddress
	BaseURL       BaseURL
}

func (o *Options) GetServer() string {
	return o.ServerAddress.String()
}

func (o *Options) GetBaseURL() string {
	return o.BaseURL.String()
}

type ServerAddress struct {
	Host string
	Port int
}

func (s *ServerAddress) String() string {
	return s.Host + ":" + strconv.Itoa(s.Port)
}

func (s *ServerAddress) Set(flag string) error {

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

func NewOptions() Options {
	return Options{
		ServerAddress: ServerAddress{Host: "localhost", Port: 8080},
		BaseURL:       BaseURL{Host: "localhost", Port: 8080},
	}
}
