package filework

import (
	"fmt"
	"github.com/White-AK111/GB_Go_Level2/Lesson8/config"
	"math/rand"
	"os"
	"sort"
	"time"
)

// DoRandomCopyFiles function for create random cope of files, return error
func DoRandomCopyFiles(cfg *config.Config) error {
	fInfo := filesInfo{}
	fInfo.directoryList = append(fInfo.directoryList, cfg.App.SourcePath)

	err := findAllFiles(cfg, &fInfo)
	if err != nil {
		cfg.App.ErrorLogger.Printf("error on find all files in source path: %s", err)
		return err
	}

	// sort files, files in root directory priority are considered like original
	sort.Slice(fInfo.allFilesList, func(i, j int) bool {
		return fInfo.allFilesList[i].Path < fInfo.allFilesList[j].Path
	})

	// sort directories, root directory is priority
	sort.Slice(fInfo.directoryList, func(i, j int) bool {
		return fInfo.directoryList[i] < fInfo.directoryList[j]
	})

	// copy files
	err = copyFiles(cfg, &fInfo)
	if err != nil {
		cfg.App.ErrorLogger.Printf("error on copy files: %s", err)
		return err
	}

	fmt.Printf("Count created random copy files: %d\n", len(fInfo.randomFilesList))
	fmt.Printf("Total files after random copy: %d\n", len(fInfo.allFilesList)+len(fInfo.randomFilesList))

	return nil
}

// copyFiles function for random copy files
func copyFiles(cfg *config.Config, fInfo *filesInfo) error {
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
			if _, err := os.Stat(pathNewFile); os.IsNotExist(err) {
				source, err := os.Open(fInfo.allFilesList[rFile].Path)
				if err != nil {
					cfg.App.ErrorLogger.Printf("error on open file: %s\n", err)
				}
				defer fileClose(cfg, source)

				destination, err := os.Create(pathNewFile)
				if err != nil {
					cfg.App.ErrorLogger.Printf("error on create file: %s\n", err)
				}
				defer fileClose(cfg, destination)

				_ = byteCopy(cfg, source, destination)
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
