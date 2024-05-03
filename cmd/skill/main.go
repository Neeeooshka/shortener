// пакеты исполняемых приложений должны называться main
package main

import (
	"github.com/Neeeooshka/alice-skill.git/internal/gzip"
	"github.com/Neeeooshka/alice-skill.git/internal/zap"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	"log"
	"net/http"
)

func main() {
	parseFlags()

	appInstance := newApp(nil)

	logger, err := zap.NewZapLogger("info")
	if err != nil {
		panic(err)
	}

	logger.Info("Running server", logger.String("address", flagRunAddr))
	// оборачиваем хендлер в middleware с логированием и поддержкой gzip
	log.Fatal(http.ListenAndServe(flagRunAddr, zap.RequestLogger(compressor.IncludeCompressor(appInstance.AliceSkill, gzip.NewGzipCompressor()))))
}
