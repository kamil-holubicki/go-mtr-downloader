package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/kamil-holubicki/go-mtr-downloader/pkg/httpdownloader"
	"github.com/kamil-holubicki/go-mtr-downloader/pkg/jenkinsdownloader"
)

const PXCJenkinsURL = "https://pxc.cd.percona.com/job/pxc-8.0-param/"
const PSJenkisnURL = "https://ps80.cd.percona.com/job/percona-server-8.0-param/"

type IDownloader interface {
	Download()
}

func paramsValid(cookie *string, job *string, dir *string, projectURL *string,
	project *string) bool {
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
	if *project == "" && *projectURL == "" {
		fmt.Println("Project or project url has to be specified")
		return false
	}
	return true
}

func getBaseUrl(url string) string {
	pos := strings.Index(url, "/job")
	return url[:pos]
}

func createDownloader(cookieFile string, projectURL string, jobNo string, downloadDir string) IDownloader {
	gd := httpdownloader.NewHttpDownloader(cookieFile, getBaseUrl(projectURL), downloadDir)
	return jenkinsdownloader.NewDownloader(gd, projectURL, jobNo)
}

func getProjectURL(project string, projectURL string) string {
	switch strings.ToLower(project) {
	case "pxc":
		return PXCJenkinsURL
	case "ps":
		return PSJenkisnURL
	default:
		return projectURL
	}
}

func main() {
	cookieFlag := flag.String("cookie", "", "Jenkins authentication cookie file")
	jobFlag := flag.String("job", "", "Parent Jenkins job number")
	projectURLFlag := flag.String("projectURL", "", "Jenkins project URL")
	dirFlag := flag.String("outputdir", "./run", "Download output dir")
	projectFlag := flag.String("project", "PS", "Project")
	flag.Parse()

	if !paramsValid(cookieFlag, jobFlag, dirFlag, projectURLFlag, projectFlag) {
		return
	}

	projectURL := getProjectURL(*projectFlag, *projectURLFlag)
	os.Mkdir(*dirFlag, 0755)
	dloader := createDownloader(*cookieFlag, projectURL, *jobFlag, *dirFlag)
	dloader.Download()
}
