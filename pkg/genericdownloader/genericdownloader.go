package genericdownloader

type Downloader interface {
	DownloadFile(url string, dest string)
	Download(url string) (string, error)
}
