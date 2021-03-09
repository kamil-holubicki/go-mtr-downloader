//go:generate mockgen -source genericdownloader.go -destination mock/genericdownloader_mock.go -package mock

package genericdownloader

type Downloader interface {
	DownloadFile(url string, dest string)
	Download(url string) (string, error)
}
