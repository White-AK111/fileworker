// Package config provide and Initialize all settings for application
package config

import (
	"crypto/sha256"
	"flag"
	"github.com/kkyr/fig"
	"hash"
	"log"
	"os"
	"path/filepath"
)

const (
	usagePath = "use this flag for set source directory"
	usageRm   = "use this flag for delete duplicate files"
	usageCp   = "use this flag for random copy files"
	usageGo   = "use this flag for set max count of goroutines"
)

// Config structure for all settings of application
type Config struct {
	App struct {
		HashAlgorithm    hash.Hash   // hash algorithm for use, don't load from configuration file
		ErrorLogger      *log.Logger // logger for use, don't load from configuration file
		SourcePath       string      `fig:"sourcePath" default:"."`        // source directory
		CountGoroutine   int         `fig:"countGoroutine" default:"10"`   // count of goroutines
		CountRndCopyIter int         `fig:"countRndCopyIter" default:"10"` // random count for create copy of files
		SizeCopyBuffer   int         `fig:"sizeCopyBuffer" default:"512"`  // copy buffer size
		FlagDelete       bool        `fig:"flagDelete"`                    // flag for delete duplicate files
		FlagRandCopy     bool        `fig:"flagRandCopy"`                  // flag fo random copy files
		RunInTest        bool        `fig:"runInTest"`                     // flag for testing, don't get approval fo delete from user
	} `fig:"app"`
}

// Init function for initialize Config structure
func Init() (*Config, error) {
	var cfg = Config{}
	err := fig.Load(&cfg, fig.Dirs("../", "./", "./..."), fig.File("config.yaml"))
	if err != nil {
		log.Fatalf("can't load configuration file: %s", err)
		return nil, err
	}

	cfg.App.ErrorLogger = NewBuiltinLogger().logger
	cfg.App.HashAlgorithm = sha256.New()

	return &cfg, err
}

// InitFlags method for initialize flags and prepare source path to ABS
func (c *Config) InitFlags() error {
	flag.StringVar(&c.App.SourcePath, "path", c.App.SourcePath, usagePath)
	flag.BoolVar(&c.App.FlagDelete, "rm", c.App.FlagDelete, usageRm)
	flag.BoolVar(&c.App.FlagRandCopy, "cp", c.App.FlagRandCopy, usageCp)
	flag.IntVar(&c.App.CountGoroutine, "go", c.App.CountGoroutine, usageGo)
	flag.Parse()

	if err := c.setABSPath(); err != nil {
		c.App.ErrorLogger.Fatalf("error on get ABS path from source path %q: %v\n", c.App.SourcePath, err)
		return err
	}

	return nil
}

// setABSPath method prepare source path to ABS
func (c *Config) setABSPath() error {
	// get absolut filepath for source path
	sourcePath, err := filepath.Abs(c.App.SourcePath)
	if err != nil {
		return err
	}
	c.App.SourcePath = sourcePath

	return nil
}

// BuiltinLogger custom logger
type BuiltinLogger struct {
	logger *log.Logger
}

// NewBuiltinLogger function custom logger initialize
func NewBuiltinLogger() *BuiltinLogger {
	return &BuiltinLogger{logger: log.New(os.Stdout, "", 5)}
}

// Debug method for print debug
func (l *BuiltinLogger) Debug(args ...interface{}) {
	l.logger.Println(args...)
}

// Debugf method for print formatted debug
func (l *BuiltinLogger) Debugf(format string, args ...interface{}) {
	l.logger.Printf(format, args...)
}
