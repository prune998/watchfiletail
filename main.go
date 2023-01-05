package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/namsral/flag"
	"github.com/nxadm/tail"
	"github.com/radovskyb/watcher"
	log "github.com/sirupsen/logrus"
)

var (
	// version is filled by -ldflags  at compile time
	version        = "no version set"
	displayVersion = flag.Bool("version", false, "Show version and quit")

	logLevel   = flag.String("loglevel", log.WarnLevel.String(), "the log level to display")
	logJSON    = flag.Bool("logjson", true, "log to stdlog using JSON format")
	folderPath = flag.String("folderPath", "/var/log", "folder to watch for log files")
	fileMatch  = flag.String("fileMatch", ".*", "regexp to match on the file names, ex: ^file.*$")
)

func printVersion() {
	log.Printf("Go Version: %s", runtime.Version())
	log.Printf("Go OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	log.Printf("watchfiletail Version: %v", version)
}

func main() {

	flag.Parse()
	if *displayVersion {
		printVersion()
		os.Exit(0)
	}

	// set log level and json format
	myLogLevel, err := log.ParseLevel(*logLevel)
	if err != nil {
		myLogLevel = log.WarnLevel
	}
	log.SetLevel(myLogLevel)

	if *logJSON {
		log.SetFormatter(&log.JSONFormatter{PrettyPrint: true})
	}

	// Create the loggers
	contextLogger := log.WithFields(log.Fields{
		"app": "watchfiletail",
	})

	// termination channel, not really used
	done := make(chan bool)

	// create a file / folder watcher
	w := watcher.New()
	w.SetMaxEvents(1)

	// Only notify create events, already existing files are watched at init phase and deleted files are handled by tail itself
	w.FilterOps(watcher.Create)

	// Only files that match the regular expression during file listings
	// will be watched.
	r := regexp.MustCompile("^file.*$")
	w.AddFilterHook(watcher.RegexFilterHook(r, false))
	contextLogger.Infof("file matcher added with regexp %s", *fileMatch)

	go func() {
		for {
			select {
			case event := <-w.Event:
				contextLogger.Info(event)
				go tailFile(event.Path, tail.Config{Follow: true}, done)
				// t, err := tail.TailFile(event.Path, tail.Config{Follow: true})
				// if err != nil {
				// 	panic(err)
				// }
				// go func() {
				// 	for line := range t.Lines {
				// 		fmt.Println(line.Text)
				// 	}
				// }()

			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				<-done
				return
			}
		}
	}()

	// Watch this folder for changes.
	if err := w.AddRecursive(*folderPath); err != nil {
		contextLogger.Errorf("error watching: %v", err)
		os.Exit(1)
	}

	// Tail config, follow files and start from the end
	config := tail.Config{
		Follow:   true,
		Location: &tail.SeekInfo{Offset: 0, Whence: io.SeekEnd},
	}

	// Print a list of all of the files and folders currently
	// being watched and their paths and start tailing.
	for path, _ := range w.WatchedFiles() {
		contextLogger.Infof("watching %s", path)
		go tailFile(path, config, done)
	}

	// Start the watching process - it'll check for changes every 100ms.
	if err := w.Start(time.Millisecond * 100); err != nil {
		contextLogger.Errorf("error starting watcher: %v", err)
		os.Exit(1)
	}
}

// tail a file print the output to stdout
func tailFile(filename string, config tail.Config, done chan bool) {
	defer func() { done <- true }()
	t, err := tail.TailFile(filename, config)
	if err != nil {
		fmt.Println(err)
		return
	}
	for line := range t.Lines {
		fmt.Println(line.Text)
	}
	err = t.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
