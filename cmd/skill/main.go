// пакеты исполняемых приложений должны называться main
package main

import (
	"database/sql"
	"log"
	"net/http"

	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/Neeeooshka/alice-skill.git/internal/store/pg"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor/gzip"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
)

func main() {
	parseFlags()

	zapLogger, err := zap.NewZapLogger("info")
	if err != nil {
		panic(err)
	}

	// создаём соединение с СУБД PostgreSQL с помощью аргумента командной строки
	conn, err := sql.Open("pgx", flagDatabaseURI)
	if err != nil {
		panic(err)
	}

	appInstance := newApp(pg.NewStore(conn))

	zapLogger.Info("Running server", zapLogger.String("address", flagRunAddr))
	// оборачиваем хендлер в middleware с логированием и поддержкой gzip
	log.Fatal(http.ListenAndServe(flagRunAddr, logger.RequestLogger(compressor.IncludeCompressor(appInstance.AliceSkill, gzip.NewGzipCompressor()))))
}
