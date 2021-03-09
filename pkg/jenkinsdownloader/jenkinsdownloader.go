package jenkinsdownloader

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"
	"sync"

	htmlquery "github.com/antchfx/xquery/html"
	"github.com/kamil-holubicki/go-mtr-downloader/pkg/genericdownloader"
)

type jenkinsDownloader struct {
	downloader  genericdownloader.Downloader
	cookie      string
	jobNo       string
	downloadDir string
	baseURL     string
	jobLinkBase string
}

func (d *jenkinsDownloader) configMatrixToLinks(s string) []string {
	// find configuration-matrix
	var result []string
	mpBody := []byte(s)
	//fmt.Println(s)
	doc, err := htmlquery.Parse(bytes.NewBuffer(mpBody))
	if err != nil {
		fmt.Println(err)
		return result
	}
	confMatrix := htmlquery.FindOne(doc, "//*/table[@id='configuration-matrix']")

	// confMatrix := xmlquery.FindOne(doc, "//table[@id='configuration-matrix']")
	links := htmlquery.Find(confMatrix, "//*/a[contains(@href,'CMAKE_BUILD_TYPE')]")

	for _, link := range links {
		result = append(result, htmlquery.SelectAttr(link, "href"))
	}
	return result
}

func (d *jenkinsDownloader) filename(URL string) string {
	buildTypeStart := strings.Index(URL, "CMAKE_BUILD_TYPE=")
	buildTypeStart = buildTypeStart + len("CMAKE_BUILD_TYPE=")
	buildTypeEnd := buildTypeStart + strings.Index(URL[buildTypeStart:], ",")
	buildType := URL[buildTypeStart:buildTypeEnd]

	osTypeStart := strings.Index(URL, "DOCKER_OS=")
	osTypeStart = osTypeStart + len("DOCKER_OS=")
	osTypeEnd := osTypeStart + strings.Index(URL[osTypeStart:], "/")
	osType := URL[osTypeStart:osTypeEnd]
	osType, _ = url.QueryUnescape(osType)

	if buildType == "RelWithDebInfo" {
		buildType = "rel"
	} else {
		buildType = "deb"
	}

	return osType + "-" + buildType + ".txt"
}

func (d *jenkinsDownloader) downloadPipelineResults(URL string, filename string) {
	d.downloader.DownloadFile(URL+"/consoleText", filename)
}

func (d *jenkinsDownloader) downloadSinglePlatformResults(URL string, wg *sync.WaitGroup) {
	defer wg.Done()
	// get platform console
	resp, err := d.downloader.Download(d.jobLink() + "/" + URL + "/console")
	if err != nil {
		fmt.Println(err)
		return
	}

	pipelineURLStart := strings.Index(resp, "started.")
	pipelineURLStart = pipelineURLStart + len("started.")
	equal := pipelineURLStart + strings.Index(resp[pipelineURLStart:], "=")
	pipelineURLStart = equal + 2

	pipelineURLEnd := pipelineURLStart + strings.Index(resp[pipelineURLStart:], ">") - 1
	pipelineURL := resp[pipelineURLStart:pipelineURLEnd]
	pipelineURL = d.baseURL + "/" + pipelineURL

	d.downloadPipelineResults(pipelineURL, d.filename(URL))

}

func (d *jenkinsDownloader) downloadPlatformsResults(links []string) {
	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)
		d.downloadSinglePlatformResults(link, &wg)
	}
	wg.Wait()
}

func (d *jenkinsDownloader) createMTRResultsLink(pipelineJobLink string) string {
	result := pipelineJobLink + "/consoleText"
	return result
}

func (d *jenkinsDownloader) jobLink() string {
	return d.jobLinkBase + d.jobNo + "/"
}

func (d *jenkinsDownloader) Download() error {
	fmt.Println("jenkinsdownloader::Download()")

	fmt.Println("Job link:", d.jobLink())
	res, err := d.downloader.Download(d.jobLink())
	if err != nil {
		fmt.Println(err)
		return err
	}

	links := d.configMatrixToLinks(res)
	d.downloadPlatformsResults(links)

	return nil
}

func getBaseUrl(url string) string {
	pos := strings.Index(url, "/job")
	return url[:pos]
}

func NewDownloader(downloader genericdownloader.Downloader, mainJobURL string, jobNo string) *jenkinsDownloader {
	return &jenkinsDownloader{
		downloader:  downloader,
		jobNo:       jobNo,
		baseURL:     getBaseUrl(mainJobURL),
		jobLinkBase: mainJobURL}
}
