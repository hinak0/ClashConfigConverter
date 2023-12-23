package main

import (
	"github.com/hinak0/ClashConfigConverter/config"
	"github.com/hinak0/ClashConfigConverter/generator"
)

func main() {
	c := config.Parse()
	generator.Integrate(c)
}
