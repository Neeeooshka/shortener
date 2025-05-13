// пакеты исполняемых приложений должны называться main
package main

import (
	"database/sql"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/Neeeooshka/alice-skill.git/internal/app"

	"github.com/Neeeooshka/alice-skill.git/internal/logger"
	"github.com/Neeeooshka/alice-skill.git/internal/store/pg"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor"
	"github.com/Neeeooshka/alice-skill.git/pkg/compressor/gzip"
	"github.com/Neeeooshka/alice-skill.git/pkg/logger/zap"
)

var (
	flagRunAddr     string
	flagLogLevel    string
	flagDatabaseURI string
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

	appInstance := app.NewSkillApp(pg.NewStore(conn))

	zapLogger.Info("Running server", zapLogger.String("address", flagRunAddr))
	// оборачиваем хендлер в middleware с логированием и поддержкой gzip
	log.Fatal(http.ListenAndServe(flagRunAddr, logger.RequestLogger(compressor.IncludeCompressor(appInstance.AliceSkill, gzip.NewGzipCompressor()))))
}

func parseFlags() {
	flag.StringVar(&flagRunAddr, "a", ":8080", "address and port to run server")
	flag.StringVar(&flagLogLevel, "l", "info", "log level")
	flag.StringVar(&flagDatabaseURI, "d", "", "database URI")
	flag.Parse()

	if envRunAddr := os.Getenv("RUN_ADDR"); envRunAddr != "" {
		flagRunAddr = envRunAddr
	}
	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		flagLogLevel = envLogLevel
	}
	if envDatabaseURI := os.Getenv("DATABASE_URI"); envDatabaseURI != "" {
		flagDatabaseURI = envDatabaseURI
	}
}
