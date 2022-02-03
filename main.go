// Package main
// В качестве завершающего задания нужно выполнить программу поиска дубликатов файлов. Дубликаты файлов - это файлы, которые совпадают по имени файла и по его размеру. Нужно написать консольную программу, которая проверяет наличие дублирующихся файлов.
// Программа должна работать на локальном компьютере и получать на вход путь до директории. Программа должна вывести в стандартный поток вывода список дублирующихся файлов, которые находятся как в директории, так и в поддиректориях директории,  переданной через аргумент командной строки. Данная функция должна работать эффективно при помощи распараллеливания программы
// Программа должна принимать дополнительный ключ - возможность удаления обнаруженных дубликатов файлов после поиска. Дополнительно нужно придумать, как обезопасить пользователей от случайного удаления файлов. В качестве ключей желательно придерживаться общепринятых практик по использованию командных опций.
// Критерии приемки программы:
// Программа компилируется
// Программа выполняет функциональность, описанную выше.
// Программа покрыта тестами
// Программа содержит документацию и примеры использования
// Программа обладает флагом “-h/--help” для краткого объяснения функциональности
// Программа должна уведомлять пользователя об ошибках, возникающих во время выполнения
// Дополнительно можете выполнить следующие задания:
// Написать программу которая по случайному принципу генерирует копии уже имеющихся файлов, относительно указанной директории
// Сравнить производительность программы в однопоточном и многопоточном режимах
package main

import (
	"context"
	"io"
	"log"

	"github.com/White-AK111/fileworker/config"
	"github.com/White-AK111/fileworker/filework"

	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"

	jaeger "github.com/uber/jaeger-client-go/config"
)

type zapWrapper struct {
	logger *zap.Logger
}

// Error logs a message at error priority
func (w *zapWrapper) Error(msg string) {
	w.logger.Error(msg)
}

// Infof logs a message at info priority
func (w *zapWrapper) Infof(msg string, args ...interface{}) {
	w.logger.Sugar().Infof(msg, args...)
}

func main() {
	// init configuration
	log.Printf("Start load configuration.\n")
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("Error on load configration file: %s", err)
	}
	cfg.App.Logger.With(zap.String("configuration file", "config.yaml")).Info("Configuration file successfully load.")

	// flushes buffer, if any
	defer cfg.App.Logger.Sync()

	// add tracer
	tracer, closer := initJaeger("fileworker", cfg.App.Logger)
	defer closer.Close()
	cfg.App.Tracer = tracer

	span, ctx := opentracing.StartSpanFromContextWithTracer(context.Background(), cfg.App.Tracer, "main")
	defer span.Finish()

	// init flags
	cfg.App.Logger.Info("Start init flags.")
	err = cfg.InitFlags()
	if err != nil {
		cfg.App.Logger.Warn("Error on init flags",
			zap.Error(err),
		)
	}
	cfg.App.Logger.Info("Flags successfully init.")

	// exec function for find and delete files
	cfg.App.Logger.Info("Start find duplicated files.")
	err = filework.DoDuplicateFiles(cfg, ctx)
	if err != nil {
		cfg.App.Logger.Fatal("Error on duplicate files function",
			zap.Error(err),
		)
	}
	cfg.App.Logger.Info("Successfully find duplicated files.")

	// exec function for create random copy files
	cfg.App.Logger.Info("Start create random copy files.")
	if cfg.App.FlagRandCopy {
		err = filework.DoRandomCopyFiles(cfg, ctx)
		if err != nil {
			cfg.App.Logger.Fatal("Error on random copy files function",
				zap.Error(err),
			)
		}
	}
	cfg.App.Logger.Info("Successfully create random copy files.")
}

func initJaeger(service string, logger *zap.Logger) (opentracing.Tracer, io.Closer) {
	cfg := &jaeger.Configuration{
		ServiceName: service,
		Sampler: &jaeger.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaeger.ReporterConfig{
			LogSpans: true,
		},
	}

	tracer, closer, err := cfg.NewTracer(jaeger.Logger(&zapWrapper{logger: logger}))
	if err != nil {
		logger.Fatal("ERROR: cannot init Jaeger",
			zap.Error(err),
		)
	}

	return tracer, closer
}
