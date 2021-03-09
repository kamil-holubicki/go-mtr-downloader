package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/kamil-holubicki/go-mtr-downloader/pkg/psdownloader"
	"github.com/kamil-holubicki/go-mtr-downloader/pkg/pxcdownloader"
)

type IDownloader interface {
	Download()
}

func paramsValid(cookie *string, job *string, dir *string) bool {
	if _, err := os.Stat(*cookie); err != nil {
		fmt.Println("Cookie file does not exist:", *cookie)
		return false
	}
	if _, err := strconv.Atoi(*job); err != nil {
		fmt.Println("Jenkins job number expected. Provided:", *job)
		return false
	}
	if _, err := os.Stat(*dir); err == nil {
		fmt.Println("Directory", *dir, "already exists")
		return false
	}
	return true
}

func createDownloader(project *string, cookie *string, job *string, dir *string) IDownloader {
	switch {
	case *project == "PS":
		return psdownloader.New(cookie, job)
	case *project == "PXC":
		return pxcdownloader.New(*cookie, *job, *dir)
	default:
		return nil
	}
}
func main() {
	cookieFlag := flag.String("cookie", "", "Jenkins authentication cookie file")
	jobFlag := flag.String("job", "", "Parent Jenkins job number")
	dirFlag := flag.String("outputdir", "./run", "Download output dir")
	projectFlag := flag.String("project", "PS", "Project")
	flag.Parse()

	if !paramsValid(cookieFlag, jobFlag, dirFlag) {
		return
	}

	os.Mkdir(*dirFlag, 0755)
	dloader := createDownloader(projectFlag, cookieFlag, jobFlag, dirFlag)
	dloader.Download()
}
