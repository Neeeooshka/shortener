// пакеты исполняемых приложений должны называться main
package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/Neeeooshka/alice-skill.git/internal/server"
	"log"
)

func main() {

	var sh []server.Handler

	sh = append(sh, server.Handler{Route: "/", Handler: handlers.AliceSkill})
	log.Fatal(server.RunHTTPServer(sh))
}
