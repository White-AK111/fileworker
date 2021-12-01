package filework

import (
	"context"
	"fmt"
	"github.com/White-AK111/fileworker/config"
	"github.com/opentracing/opentracing-go"
	"go.uber.org/zap"
	"math/rand"
	"os"
	"sort"
	"time"
)

// DoRandomCopyFiles function for create random cope of files, return error
func DoRandomCopyFiles(cfg *config.Config, ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, cfg.App.Tracer, "deleteFiles")
	defer span.Finish()

	fInfo := filesInfo{}
	fInfo.directoryList = append(fInfo.directoryList, cfg.App.SourcePath)

	err := findAllFiles(cfg, &fInfo, ctx)
	if err != nil {
		cfg.App.Logger.Error("Error on find all files in source path.",
			zap.String("path", cfg.App.SourcePath),
			zap.Error(err),
		)
		return err
	}

	// sort files, files in root directory priority are considered like original
	cfg.App.Logger.Debug("Sort files, files in root directory priority are considered like original.")
	sort.Slice(fInfo.allFilesList, func(i, j int) bool {
		return fInfo.allFilesList[i].Path < fInfo.allFilesList[j].Path
	})

	// sort directories, root directory is priority
	cfg.App.Logger.Debug("Sort directories, root directory is priority.")
	sort.Slice(fInfo.directoryList, func(i, j int) bool {
		return fInfo.directoryList[i] < fInfo.directoryList[j]
	})

	// copy files
	cfg.App.Logger.Debug("Copy files.")
	err = copyFiles(cfg, &fInfo, ctx)
	if err != nil {
		cfg.App.Logger.Error("Error on copy files.",
			zap.Error(err),
		)
		return err
	}

	fmt.Printf("Count created random copy files: %d\n", len(fInfo.randomFilesList))
	fmt.Printf("Total files after random copy: %d\n", len(fInfo.allFilesList)+len(fInfo.randomFilesList))

	return nil
}

// copyFiles function for random copy files
func copyFiles(cfg *config.Config, fInfo *filesInfo, ctx context.Context) error {
	span, ctx := opentracing.StartSpanFromContextWithTracer(ctx, cfg.App.Tracer, "copyFiles")
	defer span.Finish()

	wp := newWorkerPool(cfg.App.CountGoroutine)
	rCount := rand.Intn(cfg.App.CountRndCopyIter)
	defer wp.wg.Wait()

	for i := 0; i < rCount; i++ {
		wp.wg.Add(1)
		go func(i int) {
			defer func() {
				wp.mu.Unlock()
				// read to release a slot
				<-wp.semaphoreChan
				wp.wg.Done()
			}()
			// block while full
			wp.semaphoreChan <- struct{}{}
			wp.mu.Lock()

			rand.Seed(time.Now().UnixNano())
			rFile := rand.Intn(len(fInfo.allFilesList) - 1)
			rDir := rand.Intn(len(fInfo.directoryList) - 1)

			// check exist file before copy
			pathNewFile := fInfo.directoryList[rDir] + "/copy_" + fInfo.allFilesList[rFile].Name
			cfg.App.Logger.With(zap.String("source", fInfo.allFilesList[rFile].Path), zap.String("destination", pathNewFile)).Debug("Start copy file.")
			if _, err := os.Stat(pathNewFile); os.IsNotExist(err) {
				source, err := os.Open(fInfo.allFilesList[rFile].Path)
				if err != nil {
					cfg.App.Logger.Error("Error on open file.",
						zap.String("file", fInfo.allFilesList[rFile].Path),
						zap.Error(err),
					)
				}
				defer fileClose(cfg, source, ctx)

				destination, err := os.Create(pathNewFile)
				if err != nil {
					cfg.App.Logger.Error("Error on create file.",
						zap.String("file", pathNewFile),
						zap.Error(err),
					)
				}
				defer fileClose(cfg, destination, ctx)

				_ = byteCopy(cfg, source, destination, ctx)
				fRand := fInfo.allFilesList[rFile]
				fRand.OriginalFile = &fInfo.allFilesList[rFile]
				fRand.Path = pathNewFile
				fRand.Name = "copy_" + fInfo.allFilesList[rFile].Name
				fInfo.randomFilesList = append(fInfo.randomFilesList, fRand)
			}
		}(i)
	}

	return nil
}
