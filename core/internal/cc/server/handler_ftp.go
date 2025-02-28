package server

import (
	"net/http"
	"strconv"
	"time"

	"github.com/jm33-m0/emp3r0r/core/lib/logging"
	"github.com/jm33-m0/emp3r0r/core/lib/util"
	"github.com/schollz/progressbar/v3"
)

// progressMonitor updates the progress bar.
func progressMonitor(bar *progressbar.ProgressBar, filewrite, targetFile string, targetSize int64) {
	if targetSize == 0 {
		logging.Warningf("progressMonitor: targetSize is 0")
		return
	}
	for {
		var nowSize int64
		if util.IsFileExist(filewrite) {
			nowSize = util.FileSize(filewrite)
		} else {
			nowSize = util.FileSize(targetFile)
		}
		bar.Set64(nowSize)
		state := bar.State()
		logging.Infof("%s: %.2f%% (%d of %d bytes) at %.2fKB/s, %ds passed, %ds left",
			strconv.Quote(targetFile),
			state.CurrentPercent*100, nowSize, targetSize,
			state.KBsPerSecond, int(state.SecondsSince), int(state.SecondsLeft))
		if nowSize >= targetSize || state.CurrentPercent >= 1 {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

// handleFTPTransfer processes file transfer requests.
func handleFTPTransfer(wrt http.ResponseWriter, req *http.Request) {
}
