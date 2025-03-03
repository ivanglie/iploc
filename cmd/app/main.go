package main

import (
	"fmt"
	"os"

	"github.com/ivanglie/iploc/pkg/log"
	"github.com/rs/zerolog"

	"github.com/ivanglie/iploc/internal/server"
	"github.com/jessevdk/go-flags"
)

var (
	opts struct {
		Token string `long:"token" env:"TOKEN" description:"IP2Location token"`
		Dbg   bool   `long:"dbg" env:"DEBUG" description:"Use debug"`
		Local bool   `long:"local" env:"LOCAL" description:"For local development"`
	}
)

func main() {
	p := flags.NewParser(&opts, flags.PrintErrors|flags.PassDoubleDash|flags.HelpFlag)
	if _, err := p.Parse(); err != nil {
		if err.(*flags.Error).Type != flags.ErrHelp {
			fmt.Printf("[ERROR] iploc error: %v", err)
		}
		os.Exit(2)
	}

	if opts.Dbg {
		log.SetLogConfig(zerolog.DebugLevel, os.Stdout)
	}

	s := server.New(":8080")

	log.Info("Listening...")
	if err := s.Start(opts.Local, opts.Token, "."); err != nil {
		log.Error(err.Error())
	}
}
