// пакеты исполняемых приложений должны называться main
package main

import (
	logger2 "github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor/gzip"
	zap2 "github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
	"log"
	"net/http"
)

func main() {
	parseFlags()

	appInstance := newApp(nil)

	logger, err := zap2.NewZapLogger("info")
	if err != nil {
		panic(err)
	}

	logger.Info("Running server", logger.String("address", flagRunAddr))
	// оборачиваем хендлер в middleware с логированием и поддержкой gzip
	log.Fatal(http.ListenAndServe(flagRunAddr, logger2.RequestLogger(compressor.IncludeCompressor(appInstance.AliceSkill, gzip.NewGzipCompressor()))))
}
