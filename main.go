package main

import (
	"github.com/hinak0/ClashConfigConverter/config"
	"github.com/hinak0/ClashConfigConverter/generator"
	"github.com/hinak0/ClashConfigConverter/log"
)

func main() {
	c := config.Parse()
	log.Infoln("Parse config.yaml success.")
	generator.Integrate(c)
	log.Infoln("Generate target success.")
}
