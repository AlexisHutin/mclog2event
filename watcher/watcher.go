package watcher

import (
	"log"

	"github.com/hpcloud/tail"
)

// Watch starts tailing the log file and sends new lines to the channel.
func Watch(logFilePath string, lines chan<- string) {
	defer close(lines)

	t, err := tail.TailFile(logFilePath, tail.Config{
		Follow:    true,
		ReOpen:    true,
		MustExist: true,
		Location:  &tail.SeekInfo{Offset: 0, Whence: 2}, // Seek to end of file
	})
	if err != nil {
		log.Fatalf("failed to tail file: %v", err)
	}

	for line := range t.Lines {
		lines <- line.Text
	}
}
