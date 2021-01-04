package main

import (
	_ "code-platform/boot"
	_ "code-platform/router"
	"github.com/gogf/gf/frame/g"
	"github.com/gogf/swagger"
)

func main() {
	s := g.Server()
	s.Plugin(&swagger.Swagger{})
	s.Run()
}
