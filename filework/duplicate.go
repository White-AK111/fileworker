package filework

import (
	"fmt"
	"github.com/White-AK111/fileworker/config"
	"os"
	"sort"
	"strings"
)

// DoDuplicateFiles function for find duplicate file in source directory files compares by hash, optionality user can delete all duplicate files, return error
func DoDuplicateFiles(cfg *config.Config) error {
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

	// compare files
	for i, file := range fInfo.allFilesList {
		if file.contains(fInfo.allFilesList, i) {
			fmt.Printf("Duplicate file: %s	Original file: %s\n", file.Path, file.OriginalFile.Path)
			fInfo.duplicateFilesList = append(fInfo.duplicateFilesList, file)
		}
	}

	fmt.Printf("Total files: %d\n", len(fInfo.allFilesList))
	fmt.Printf("Duplicate files (without original file): %d\n", len(fInfo.duplicateFilesList))

	// delete files if get a flag, get approval from user
	if cfg.App.FlagDelete || cfg.App.RunInTest {
		if len(fInfo.duplicateFilesList) == 0 {
			fmt.Println("No files for delete!")
		} else {
			var confirm string
			if !cfg.App.RunInTest {
				for strings.ToUpper(confirm) != "Y" && strings.ToUpper(confirm) != "N" {
					fmt.Print("Delete this duplicate files? (Y/N): ")
					_, err = fmt.Fscan(os.Stdin, &confirm)
					if err != nil {
						cfg.App.ErrorLogger.Printf("error on get approval from console: %s\n", err)
						return err
					}
				}
			}
			if strings.ToUpper(confirm) == "Y" || cfg.App.RunInTest {
				err = deleteFiles(cfg, &fInfo)
				if err != nil {
					cfg.App.ErrorLogger.Printf("error on delete files: %s\n", err)
					return err
				}
				fmt.Println("Files deleted!")
			}
		}
	}

	return nil
}

// deleteFiles function delete files, return error
func deleteFiles(cfg *config.Config, fInfo *filesInfo) error {
	wp := newWorkerPool(cfg.App.CountGoroutine)
	defer wp.wg.Wait()

	for _, file := range fInfo.duplicateFilesList {
		wp.wg.Add(1)
		go func(file FileEntity) {
			defer func() {
				wp.mu.Unlock()
				// read to release a slot
				<-wp.semaphoreChan
				fInfo.deleteFilesList = append(fInfo.deleteFilesList, file)
				wp.wg.Done()
			}()
			// block while full
			wp.semaphoreChan <- struct{}{}
			wp.mu.Lock()
			if err := os.Remove(file.Path); err != nil {
				cfg.App.ErrorLogger.Printf("error on delete file %s: %s\n", file.Path, err)
			}
		}(file)
	}

	return nil
}
