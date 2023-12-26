package main

//go:generate go run directives_generate.go
//go:generate go run owners_generate.go

import (
	_ "github.com/coredns/coredns/core/plugin" // Plug in CoreDNS.
	"github.com/coredns/coredns/coremain"
	"github.com/coredns/coredns/plugin/pkg/log"
)

func main() {
	log.Info("coreDNS start!")
	coremain.Run()
}
