package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/gdamore/tcell"
	"github.com/golang/glog"
	"github.com/rivo/tview"
	"github.com/tupyy/lazylogger/internal/conf"
	"github.com/tupyy/lazylogger/internal/gui"
	"github.com/tupyy/lazylogger/internal/log"
)

// build flags
var (
	Version string

	Build string

	configurationFile string

	// app
	app = tview.NewApplication()

	loggerManager *log.LoggerManager
)

func main() {

	fmt.Printf("LazyLogger Version: %s Build: %s\n\n", Version, Build)

	// Read configuration
	flag.StringVar(&configurationFile, "config", "nodata", "JSON configuration file")
	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	flag.Parse()

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			glog.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	config := conf.ReadConfigurationFile(configurationFile)
	glog.Infof("Configuration has %d.", len(config.LoggerConfigurations))

	// create the loggerManager
	glog.Info("Create logger manager")
	loggerManager = log.NewLoggerManager(config.LoggerConfigurations)
	go loggerManager.Run()
	defer loggerManager.Stop()

	gui := gui.NewGui(app, loggerManager)

	// ESC exits
	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc {
			gui.Stop()
			loggerManager.Stop()
			app.Stop()
		}
		gui.HandleEventKey(event)
		return event
	})

	app.SetRoot(gui.Layout(), true)
	gui.Start()
	app.Run()
}
