// пакеты исполняемых приложений должны называться main
package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/handlers"
	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"go.uber.org/zap"
	"log"
	"net/http"
)

func main() {
	// обрабатываем аргументы командной строки
	parseFlags()

	if err := logger.Initialize(flagLogLevel); err != nil {
		panic(err)
	}

	logger.Log.Info("Running server", zap.String("address", flagRunAddr))
	log.Fatal(http.ListenAndServe(flagRunAddr, logger.RequestLogger(handlers.AliceSkill)))
}
