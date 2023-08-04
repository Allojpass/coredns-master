package main

//go:generate go run directives_generate.go
//go:generate go run owners_generate.go

import (
	"fmt"
	_ "github.com/coredns/coredns/core/plugin" // Plug in CoreDNS.
	"github.com/coredns/coredns/coremain"
	"github.com/coredns/coredns/plugin/pkg/log"
)

func main() {
	fmt.Println("fmt启动！")
	log.Info("log启动！")
	coremain.Run()
}
