package psdownloader

import "fmt"

type Downloader struct {
	cookie string
	job    string
}

func (d *Downloader) Download() {
	fmt.Println("psdownloader::Download()")
}

func New(cookie *string, job *string) *Downloader {
	return &Downloader{cookie: *cookie, job: *job}
}
