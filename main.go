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
	"github.com/White-AK111/fileworker/config"
	"github.com/White-AK111/fileworker/filework"
	"go.uber.org/zap"
	"log"
)

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
	err = filework.DoDuplicateFiles(cfg)
	if err != nil {
		cfg.App.Logger.Fatal("Error on duplicate files function",
			zap.Error(err),
		)
	}
	cfg.App.Logger.Info("Successfully find duplicated files.")

	// exec function for create random copy files
	cfg.App.Logger.Info("Start create random copy files.")
	if cfg.App.FlagRandCopy {
		err = filework.DoRandomCopyFiles(cfg)
		if err != nil {
			cfg.App.Logger.Fatal("Error on random copy files function",
				zap.Error(err),
			)
		}
	}
	cfg.App.Logger.Info("Successfully create random copy files.")
}
