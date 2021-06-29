package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/adrg/xdg"
	"github.com/binadox-public/binadox-cloud-agent/engine"
	"os"
	"path"
)

var (
    sha1ver   string // sha1 revision used to build the program
    buildTime string // when the executable was built
)

func getWorkDir() string {
	return path.Join(xdg.CacheHome, "binadox-agent")
}

var (
	workDir string
	workSpace string
)

func parseCmdLineFlags() {
	var (
		flgVersion bool
		flgWorkDir string
		flgWorkspace string
	)

    flag.BoolVar(&flgVersion, "version", false, "if set, print version and exit")
    flag.StringVar(&flgWorkDir, "workdir", getWorkDir(), "path to the application data")
	flag.StringVar(&flgWorkspace, "workspace", "", "Binadox workspace id")
    flag.Parse()
    if flgVersion {
        fmt.Printf("Build on %s from sha1 %s\n", buildTime, sha1ver)
        os.Exit(0)
    }
    workDir = flgWorkDir
    workSpace = flgWorkspace
}

func main() {

	parseCmdLineFlags()

	var stats *engine.InstanceInfo
	var err error
	ctx := engine.InitFetcher(workDir)
	stats, err = engine.Fetch(&ctx)
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
