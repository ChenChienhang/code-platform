package main

import (
	_ "code-platform/boot"
	_ "code-platform/router"
	"github.com/gogf/gf/frame/g"
)

func main() {
	s := g.Server()
	s.Run()
}
