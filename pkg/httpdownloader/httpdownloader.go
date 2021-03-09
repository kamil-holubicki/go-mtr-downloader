package httpdownloader

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"time"

	"github.com/aki237/nscjar"
	"github.com/kamil-holubicki/go-mtr-downloader/pkg/genericdownloader"
)

func NewHttpDownloader(cookieFile string, cookieUrl string, downloadDir string) genericdownloader.Downloader {
	httpClient := createHttpClient(cookieUrl, cookieFile)
	if httpClient == nil {
		fmt.Printf("Problem creating http client")
		return nil
	}
	return &HttpDownloaderImpl{
		httpClient:  httpClient,
		downloadDir: downloadDir}
}

func createHttpClient(baseURL string, cookieFile string) *http.Client {
	jar := nscjar.Parser{}
	f, err := os.Open(cookieFile)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	cookies, err := jar.Unmarshal(f)
	if err != nil {
		fmt.Println(err)
	}

	now := time.Now()
	for i := 0; i < len(cookies); i++ {
		cookies[i].Expires = now.Add(3 * time.Hour)
	}
	cookieJar, err := cookiejar.New(nil)
	if err != nil {
		return nil
	}

	cURL, _ := url.Parse(baseURL)
	cookieJar.SetCookies(cURL, cookies)
	httpClient := &http.Client{
		Jar: cookieJar,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	return httpClient
}

type HttpDownloaderImpl struct {
	httpClient  *http.Client
	downloadDir string
}

func (d *HttpDownloaderImpl) Download(url string) (string, error) {
	fmt.Println("Get: " + url)
	resp, err := d.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (d *HttpDownloaderImpl) DownloadFile(url string, dest string) {
	fmt.Println("Get: " + url + " -> " + dest)
	resp, err := d.httpClient.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	filepath := d.downloadDir + "/" + dest
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
