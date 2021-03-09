package jenkinsdownloader_test

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	gdmock "github.com/kamil-holubicki/go-mtr-downloader/pkg/genericdownloader/mock"
	"github.com/kamil-holubicki/go-mtr-downloader/pkg/jenkinsdownloader"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type JenkinsDownloaderSuite struct {
	suite.Suite
	*require.Assertions

	ctrl *gomock.Controller
}

func TestLockerSuite(t *testing.T) {
	suite.Run(t, new(JenkinsDownloaderSuite))
}

// ran before every test
func (s *JenkinsDownloaderSuite) SetupTest() {
	s.Assertions = require.New(s.T())
	s.ctrl = gomock.NewController(s.T())
}

// ran after every test
func (s *JenkinsDownloaderSuite) TearDownTest() {
	s.ctrl.Finish()
}

func (s *JenkinsDownloaderSuite) TestDownloaderFails() {
	m := gdmock.NewMockDownloader(s.ctrl)
	m.EXPECT().Download(gomock.Any()).Return("", errors.New("error")).AnyTimes()

	mainJobURL := "https://one.two.com/job/test"
	jobNo := "10"
	jd := jenkinsdownloader.NewDownloader(m, mainJobURL, jobNo)
	err := jd.Download()

	assert.NotNil(s.T(), err)
}
