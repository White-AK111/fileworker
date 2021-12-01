// Package filework contains function for find duplicate file in source directory
// files compares by hash
// optionality user can delete all duplicate files and create random cope of files
package filework

import (
	"encoding/hex"
	"github.com/White-AK111/fileworker/config"
	"go.uber.org/zap"
	"io"
	"os"
	"sync"
	"time"
)

// workerPool struct for management goroutines
type workerPool struct {
	wg            sync.WaitGroup
	resultChan    chan FileEntity
	semaphoreChan chan struct{}
	mu            sync.Mutex
}

// filesInfo struct for temporary save result of work
type filesInfo struct {
	allFilesList       []FileEntity // list with all files
	duplicateFilesList []FileEntity // list with duplicate files
	deleteFilesList    []FileEntity // list with deleted files
	randomFilesList    []FileEntity // list with random create files
	directoryList      []string
}

// newWorkerPool method initialize new WorkerPool, return *workerPool
func newWorkerPool(N int) *workerPool {
	return &workerPool{
		wg:            sync.WaitGroup{},
		resultChan:    make(chan FileEntity, N),
		semaphoreChan: make(chan struct{}, N),
	}
}

// FileEntity struct for save file information
type FileEntity struct {
	OriginalFile *FileEntity // pointer to original file
	Create       time.Time   // time of create file
	Name         string      // name of file
	Path         string      // path to file with name of file
	Hash         string      // hash of file
	Size         int64       // size of file
}

// getHashOfFile method get hash of file, return error
func (f *FileEntity) getHashOfFile(cfg *config.Config) error {
	file, err := os.Open(f.Path)
	if err != nil {
		return err
	}
	defer fileClose(cfg, file)
	cfg.App.Logger.With(zap.String("file", f.Path)).Debug("Open file for get hash.")

	cfg.App.HashAlgorithm.Reset()
	if _, err := io.Copy(cfg.App.HashAlgorithm, file); err != nil {
		return err
	}

	hashInBytes := cfg.App.HashAlgorithm.Sum(nil)
	f.Hash = hex.EncodeToString(hashInBytes)
	cfg.App.Logger.With(zap.String("file", f.Path), zap.String("hash", f.Hash)).Debug("Get file hash.")

	return nil
}

// contains method check contains file in slice of file from i position, return bool
func (f *FileEntity) contains(fl []FileEntity, it int) bool {
	for i := it + 1; i < len(fl); i++ {
		if fl[i].Hash == f.Hash {
			f.OriginalFile = &fl[i]
			return true
		}
	}
	return false
}

// newFileEntity method initialize new FileEntity, return *FileEntity
func newFileEntity() *FileEntity {
	return &FileEntity{
		Create: time.Now(),
		Name:   "",
		Path:   "",
		Hash:   "",
		Size:   0,
	}
}

// findAllFiles function find all files in source directory without directories, save files info in filesInfo struct, return error
func findAllFiles(cfg *config.Config, fInfo *filesInfo) error {
	wp := newWorkerPool(cfg.App.CountGoroutine)
	defer wp.wg.Wait()

	wp.wg.Add(1)
	lsFiles(cfg.App.SourcePath, cfg, wp, fInfo)

	return nil
}

// lsFiles recursive function for find files in all directories
func lsFiles(dir string, cfg *config.Config, wp *workerPool, fInfo *filesInfo) {
	// block while full
	wp.semaphoreChan <- struct{}{}

	go func() {
		defer catchRecover(cfg)
		defer func() {
			wp.mu.Unlock()
			// read to release a slot
			<-wp.semaphoreChan
			wp.wg.Done()
		}()

		wp.mu.Lock()
		cfg.App.Logger.With(zap.String("directory", dir)).Debug("Open directory.")
		file, err := os.Open(dir)
		if err != nil {
			cfg.App.Logger.Error("Error opening directory",
				zap.String("directory", dir),
				zap.Error(err),
			)
		}

		defer fileClose(cfg, file)

		// loads all children files into memory
		files, err := file.Readdir(-1)
		if err != nil {
			cfg.App.Logger.Error("Error reading directory",
				zap.String("directory", dir),
				zap.Error(err),
			)
		}

		for _, f := range files {
			path := dir + "/" + f.Name()
			if f.IsDir() {
				cfg.App.Logger.With(zap.String("directory", path)).Debug("Go to child directory.")
				fInfo.directoryList = append(fInfo.directoryList, path)
				wp.wg.Add(1)
				if cfg.App.DoPanic {
					panic("Panic!")
				}
				go lsFiles(path, cfg, wp, fInfo)
			} else {
				cfg.App.Logger.With(zap.String("file", path)).Debug("Find file in directory.")
				fe := newFileEntity()
				fe.Name = f.Name()
				fe.Path = path
				fe.Create = f.ModTime()
				// get hash of file
				if err = fe.getHashOfFile(cfg); err != nil {
					cfg.App.Logger.Error("Can't get hash of file",
						zap.String("file", path),
						zap.Error(err),
					)
				}
				fe.Size = f.Size()
				fInfo.allFilesList = append(fInfo.allFilesList, *fe)
			}
		}
	}()
}

// fileClose function for defer close file or directory
func fileClose(cfg *config.Config, file *os.File) {
	cfg.App.Logger.With(zap.String("path", file.Name())).Debug("Close file or directory.")
	err := file.Close()
	if err != nil {
		cfg.App.Logger.Error("Error on defer close file or directory",
			zap.String("path", file.Name()),
			zap.Error(err),
		)
	}
}

// byteCopy function for copy file by use buffer
func byteCopy(cfg *config.Config, source *os.File, destination *os.File) error {
	cfg.App.Logger.With(zap.String("source", source.Name()), zap.String("destination", destination.Name())).Debug("Copy file.")
	buf := make([]byte, cfg.App.SizeCopyBuffer)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			cfg.App.Logger.Error("Error on byte read from file",
				zap.String("file", source.Name()),
				zap.Error(err),
			)
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			cfg.App.Logger.Error("Error on byte write to file",
				zap.String("file", destination.Name()),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

// catchRecover func for do recover in another functions
func catchRecover(cfg *config.Config) {
	if r := recover(); r != nil {
		cfg.App.Logger.Panic("Catch panic!")
	}
}
