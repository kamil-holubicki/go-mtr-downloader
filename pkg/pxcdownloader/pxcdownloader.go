package pxcdownloader

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/aki237/nscjar"
	htmlquery "github.com/antchfx/xquery/html"
)

const jobLinkBase = "https://pxc.cd.percona.com/job/pxc-8.0-param/"

type Downloader struct {
	cookie      string
	job         string
	jobLink     string
	downloadDir string
	httpClient  *http.Client
}

func (d *Downloader) configMatrixToLinks(s string) []string {
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

func (d *Downloader) filename(URL string) string {
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

func (d *Downloader) downloadPipelineResults(URL string, filename string) {
	d.downloadSingleFile(URL+"/consoleText", filename)
}

func (d *Downloader) downloadSinglePlatformResults(URL string, wg *sync.WaitGroup) {
	defer wg.Done()
	// get platform console
	u := d.jobLink + "/" + URL + "/console"
	fmt.Println("Get: " + u)
	resp, err := d.httpClient.Get(u)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		b := string(body)
		//fmt.Println("===========================================")
		//fmt.Println(b)
		//fmt.Println("===========================================")

		pipelineURLStart := strings.Index(b, "started.")
		pipelineURLStart = pipelineURLStart + len("started.")
		equal := pipelineURLStart + strings.Index(b[pipelineURLStart:], "=")
		pipelineURLStart = equal + 2

		pipelineURLEnd := pipelineURLStart + strings.Index(b[pipelineURLStart:], ">") - 1
		pipelineURL := b[pipelineURLStart:pipelineURLEnd]
		pipelineURL = "https://pxc.cd.percona.com/" + pipelineURL

		d.downloadPipelineResults(pipelineURL, d.filename(URL))

	}

}

func (d *Downloader) downloadSingleFile(url string, filename string) {
	fmt.Println("Get: " + url)
	resp, err := d.httpClient.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	filepath := d.downloadDir + "/" + filename
	out, err := os.Create(filepath)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		fmt.Println(err)
	}
}

func (d *Downloader) downloadPlatformsResults(links []string) {
	var wg sync.WaitGroup

	for _, link := range links {
		wg.Add(1)
		d.downloadSinglePlatformResults(link, &wg)
	}
	wg.Wait()
}

func (d *Downloader) createMTRResultsLink(pipelineJobLink string) string {
	result := pipelineJobLink + "/consoleText"
	return result
}

func (d *Downloader) Download() {
	fmt.Println("pxcdownloader::Download()")
	jar := nscjar.Parser{}
	f, err := os.Open(d.cookie)
	if err != nil {
		fmt.Println(err)
		return
	}
	cookies, err := jar.Unmarshal(f)
	if err != nil {
		fmt.Println(err)
	}
	//fmt.Println("total cookies: ", len(cookies))
	now := time.Now()
	for i := 0; i < len(cookies); i++ {
		cookies[i].Expires = now.Add(3 * time.Hour)
		//fmt.Println("cookie: ", cookies[i])
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		fmt.Println("Cannot create cookie jar")
		return
	}

	cookieURL, _ := url.Parse("https://pxc.cd.percona.com")
	cookieJar.SetCookies(cookieURL, cookies)
	d.httpClient = &http.Client{
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	//fmt.Println("cookies len:", len(d.httpClient.Jar.Cookies(cookieURL)))
	//fmt.Println(d.httpClient.Jar.Cookies(cookieURL))
	d.jobLink = jobLinkBase + d.job + "/"
	fmt.Println("Job link:", d.jobLink)
	resp, err := d.httpClient.Get(d.jobLink)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer resp.Body.Close()

	if body, err := ioutil.ReadAll(resp.Body); err == nil {
		b := string(body)
		links := d.configMatrixToLinks(b)
		//links = d.createMtrLinks(jobLink, links)
		d.downloadPlatformsResults(links)
	}
}

func New(cookie string, job string, dir string) *Downloader {
	return &Downloader{cookie: cookie, job: job, downloadDir: dir}
}
