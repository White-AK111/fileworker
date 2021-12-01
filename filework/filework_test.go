package filework

import (
	"github.com/White-AK111/fileworker/config"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"testing"
)

// TestDoDuplicateFiles test for DoDuplicateFiles function
func TestDoDuplicateFiles(t *testing.T) {
	cfg, err := config.Init()
	if err != nil {
		t.Fatalf("error on load configuration file: %s", err)
	}

	cfg.App.FlagDelete = true
	cfg.App.SourcePath = "../TestFiles"
	cfg.App.RunInTest = true

	filesB, _ := ioutil.ReadDir(cfg.App.SourcePath)

	err = DoDuplicateFiles(cfg)
	if err != nil {
		t.Fatalf("error on duplicate files function: %s", err)
	}

	filesA, _ := ioutil.ReadDir(cfg.App.SourcePath)

	assert.NotEqual(t, filesB, filesA, "count of file don't changes, before: %d, after: %d", len(filesB), len(filesA))
}

// TestDoRandomCopyFiles test for DoRandomCopyFiles function
func TestDoRandomCopyFiles(t *testing.T) {
	cfg, err := config.Init()
	if err != nil {
		t.Fatalf("error on load configuration file: %s", err)
	}

	cfg.App.FlagRandCopy = true
	cfg.App.FlagDelete = false
	cfg.App.SourcePath = "../TestFiles"
	cfg.App.RunInTest = true

	filesB, _ := ioutil.ReadDir(cfg.App.SourcePath)

	err = DoRandomCopyFiles(cfg)
	if err != nil {
		t.Fatalf("error on random copy files function: %s", err)
	}

	filesA, _ := ioutil.ReadDir(cfg.App.SourcePath)

	assert.NotEqual(t, filesB, filesA, "count of file don't changes, before: %d, after: %d", len(filesB), len(filesA))
}

// BenchmarkDoDuplicateFiles_1go bench for DoDuplicateFiles function, use 1 goroutine
func BenchmarkDoDuplicateFiles_1go(b *testing.B) {
	cfg, err := config.Init()
	if err != nil {
		b.Fatalf("error on load configuration file: %s", err)
	}

	cfg.App.FlagDelete = false
	cfg.App.SourcePath = "../TestFiles"
	cfg.App.RunInTest = true
	cfg.App.CountGoroutine = 1

	for i := 0; i < b.N; i++ {
		if err := DoDuplicateFiles(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDoRandomCopyFiles_1go bench for DoRandomCopyFiles function, use 1 goroutine
func BenchmarkDoRandomCopyFiles_1go(b *testing.B) {
	cfg, err := config.Init()
	if err != nil {
		b.Fatalf("error on load configuration file: %s", err)
	}

	cfg.App.FlagRandCopy = true
	cfg.App.FlagDelete = false
	cfg.App.SourcePath = "../TestFiles"
	cfg.App.RunInTest = true
	cfg.App.CountGoroutine = 1
	cfg.App.CountRndCopyIter = 1000

	for i := 0; i < b.N; i++ {
		if err := DoRandomCopyFiles(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDoDuplicateFiles_1000go bench for DoDuplicateFiles function, use 1000 goroutine
func BenchmarkDoDuplicateFiles_1000go(b *testing.B) {
	cfg, err := config.Init()
	if err != nil {
		b.Fatalf("error on load configuration file: %s", err)
	}

	cfg.App.FlagDelete = false
	cfg.App.SourcePath = "../TestFiles"
	cfg.App.RunInTest = true
	cfg.App.CountGoroutine = 1000

	for i := 0; i < b.N; i++ {
		if err := DoDuplicateFiles(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDoRandomCopyFiles_1000go bench for DoRandomCopyFiles function, use 1000 goroutine
func BenchmarkDoRandomCopyFiles_1000go(b *testing.B) {
	cfg, err := config.Init()
	if err != nil {
		b.Fatalf("error on load configuration file: %s", err)
	}

	cfg.App.FlagRandCopy = true
	cfg.App.FlagDelete = false
	cfg.App.SourcePath = "../TestFiles"
	cfg.App.RunInTest = true
	cfg.App.CountGoroutine = 1000
	cfg.App.CountRndCopyIter = 1000

	for i := 0; i < b.N; i++ {
		if err := DoRandomCopyFiles(cfg); err != nil {
			b.Fatal(err)
		}
	}
}

// ExampleDoDuplicateFiles example for use DoDuplicateFiles function
func ExampleDoDuplicateFiles() {
	// set source directory in config.yaml file or use flags (-h for help)
	// ...

	// init configuration for app
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("error on load configration file: %s", err)
	}

	// init flags for app
	err = cfg.InitFlags()
	if err != nil {
		log.Fatalf("error on initialize flags: %s", err)
	}

	// function delete find duplicate files in source directory and delete this if flag -rm use (or it's set in config file config.yaml)
	err = DoDuplicateFiles(cfg)
	if err != nil {
		cfg.App.Logger.Fatal("Error on duplicate files function",
			zap.Error(err),
		)
	}
}

// ExampleDoRandomCopyFiles example for use DoRandomCopyFiles function
func ExampleDoRandomCopyFiles() {
	// set source directory in config.yaml file or use flags (-h for help)
	// ...

	// init configuration for app
	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("error on load configration file: %s", err)
	}

	// init flags for app
	err = cfg.InitFlags()
	if err != nil {
		log.Fatalf("error on initialize flags: %s", err)
	}

	// function delete random create files in source directory use flag -cp (or it's set in config file config.yaml)
	if cfg.App.FlagRandCopy {
		err = DoRandomCopyFiles(cfg)
		if err != nil {
			cfg.App.Logger.Fatal("Error on random copy files function",
				zap.Error(err),
			)
		}
	}
}
