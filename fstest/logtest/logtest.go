package logtest

import (
	"bytes"
	"io"
	"log"
	"os"
)

func CaptureLogging(print func()) string {
	// keep backup of the real stdout
	old := log.Default().Writer()
	r, w, _ := os.Pipe()
	log.Default().SetOutput(w)

	print()

	outC := make(chan string)

	// send stdout to channel
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// restore
	w.Close()
	log.Default().SetOutput(old)
	return <-outC
}
