package downloader

import "net/http"

type downloader struct {
}

func (d *downloader) CreateHttpClient(URL string, cookie string) *http.Client {

}
