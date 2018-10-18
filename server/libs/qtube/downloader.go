package qtube

import (
	"net/http"
)

import (
	"github.com/benderpan/quickshare/server/libs/fileidx"
)

type Downloader interface {
	ServeFile(res http.ResponseWriter, req *http.Request, fileInfo *fileidx.FileInfo) error
}
