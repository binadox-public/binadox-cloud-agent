package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/binadox-public/binadox-cloud-agent/engine"
	"log"
	"os"
	"os/exec"
)

var (
    sha1ver   string // sha1 revision used to build the program
    buildTime string // when the executable was built
    versionTag string
    inRelease string
)

var (
	workSpace string
)

func parseCmdLineFlags() {
	var (
		flgVersion bool
		flgWorkDir string
		flgWorkspace string
	)

    flag.BoolVar(&flgVersion, "version", false, "if set, print version and exit")
    flag.StringVar(&flgWorkDir, "workdir", "", "path to the application data")
	flag.StringVar(&flgWorkspace, "workspace", "", "Binadox workspace id")
    flag.Parse()
    if flgVersion {
        fmt.Printf("Build on %s from sha1 %s tag %s\n", buildTime, sha1ver, versionTag)
        os.Exit(0)
    }
    if len(flgWorkDir) > 0 {
		engine.SetWorkDir(flgWorkDir)
	}
    workSpace = flgWorkspace
}

func runSelf(ctx *engine.FetcherContext) {
	stats, err := engine.Fetch(ctx)
	if err == nil {
		stats.Version = buildTime
		var bytes []byte
		bytes, err = json.MarshalIndent(stats, "", " ")
		if err == nil {
			jsonTxt := string(bytes)
			fmt.Printf("%v\n", jsonTxt)
		}
	}
}

func main() {
	parseCmdLineFlags()
	err := os.MkdirAll(engine.GetUpdaterDir(), os.ModePerm)
	ctx := engine.InitFetcher(engine.GetCacheDir())

	newExe, _ := engine.FetchRelease(versionTag)
	if len(newExe) > 0 {
		err = engine.SetLatestApplication(&ctx, newExe)
	}

	var currentApp string
	if err == nil {
		currentApp, err = engine.GetLatestApplication(&ctx)
	}

	if err != nil || len(currentApp) == 0 {
		runSelf(&ctx)
	} else {
		cmd := exec.Command(currentApp, "--workdir", engine.GetWorkDir(), "--workspace", workSpace)
		err = cmd.Run()

    	if err != nil {
        	log.Printf("Error executing %s : %v", currentApp, err.Error())
        	runSelf(&ctx)
    	}
	}
}
