package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/binadox-public/binadox-cloud-agent/engine"
	"log"
	"net/http"
	"os"
	"os/exec"
)

var (
    sha1ver   string // sha1 revision used to build the program
    buildTime string // when the executable was built
    versionTag string
)

var (
	securityToken string
	serverlUrl string
	dryRun bool
)

func parseCmdLineFlags() {
	var (
		flgVersion            bool
		flgWorkDir            string
		flgToken              string
		flgGenerateSignatures bool
		flgSign               bool
		flgInFile             string
		flgOFile              string
		flgPriv               string
		flgUrl                string
		flgDryRun             bool
	)

    flag.BoolVar(&flgVersion, "version", false, "if set, print version and exit")
    flag.StringVar(&flgWorkDir, "workdir", "", "path to the application data")
	flag.StringVar(&flgToken, "token", "", "Binadox security token")
	flag.StringVar(&flgUrl, "url", "", "Binadox endpoint url")
	flag.BoolVar(&flgDryRun, "dry-run", false, "if set, just output revealed data")

	flag.BoolVar(&flgGenerateSignatures, "generate-signatures", false, "generate keys and and exit")
	flag.BoolVar(&flgSign, "zip", false, "generate distribution zip")
	flag.StringVar(&flgInFile, "in", "", "input file")
	flag.StringVar(&flgOFile, "out", "", "output file")
	flag.StringVar(&flgPriv, "priv", "", "private key")

    flag.Parse()
    if flgVersion {
        fmt.Printf("Build on %s from sha1 %s tag %s\n", buildTime, sha1ver, versionTag)
        os.Exit(0)
    }
    if len(flgWorkDir) > 0 {
		engine.SetWorkDir(flgWorkDir)
	}

	if flgGenerateSignatures {
		err := engine.PrintSignKeys()
		if err != nil {
			fmt.Printf("%s\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if flgSign {
		if len(flgInFile) == 0 || len(flgOFile) == 0 || len(flgPriv) == 0 {
			fmt.Printf("Required args missing")
			os.Exit(1)
		}
		if errZip := engine.CreateDistZip(flgInFile, flgOFile, flgPriv); errZip != nil {
			fmt.Printf("Error: %s\n", errZip.Error())
			os.Exit(1)
		}
		if err := engine.VerifyZip(flgOFile); err != nil {
			fmt.Printf("Error: %s\n", err.Error())
			os.Exit(1)
		} else {
			fmt.Printf("OK\n")
		}
		os.Exit(0)
	}
	dryRun = flgDryRun
	if !dryRun {
		securityToken = flgToken
		if len(securityToken) == 0 {
			fmt.Printf("--token argument is missing")
			os.Exit(1)
		}
		serverlUrl = flgUrl
		if len(serverlUrl) == 0 {
			fmt.Printf("--url argument is missing")
			os.Exit(1)
		}
	}
}

func sendData(body []byte) {
	if dryRun {
		return
	}
	req, err := http.NewRequest("POST", serverlUrl, bytes.NewBuffer(body))
	if err != nil {
		log.Fatalf("Error %s", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("MonitoringToken %s", securityToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error %s", err)
	}
	defer resp.Body.Close()
}

func runSelf(ctx *engine.FetcherContext, ver string) {
	stats, err := engine.Fetch(ctx, ver)
	if stats == nil {
		return
	}
	if err == nil {
		var bytesData []byte
		bytesData, err = json.MarshalIndent(stats, "", " ")
		if err == nil {
			sendData(bytesData)
			jsonTxt := string(bytesData)
			if dryRun {
				fmt.Printf("%v\n", jsonTxt)
			}
		}
	}
}

func main() {
	parseCmdLineFlags()
	err := os.MkdirAll(engine.GetUpdaterDir(), os.ModePerm)
	ctx := engine.InitFetcher(engine.GetCacheDir())



	var myVer string
	if len(versionTag) == 0 {
		myVer = "v0.0.0"
	} else {
		myVer = versionTag
	}

	if dryRun {
		runSelf(&ctx, myVer)
		os.Exit(0)
	}

	newExe, _ := engine.FetchRelease(myVer)
	if len(newExe) > 0 {
		err = engine.SetLatestApplication(&ctx, newExe)
	}

	var currentApp string
	if err == nil {
		currentApp, err = engine.GetLatestApplication(&ctx)
	}

	if err != nil || len(currentApp) == 0 {
		runSelf(&ctx, myVer)
	} else {
		cmd := exec.Command(currentApp, "--workdir", engine.GetWorkDir(), "--token", securityToken, "--url", serverlUrl)
		err = cmd.Run()

    	if err != nil {
        	log.Printf("Error executing %s : %v", currentApp, err.Error())
        	runSelf(&ctx, myVer)
    	}
	}
}
