package main

import (
	"Sentinel/lib"
	"Sentinel/lib/utils"
	"fmt"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	var (
		httpClient   *http.Client
		err          error
		localVersion string
		repoVersion  string
		sigChan      chan os.Signal
		filePaths    *utils.FilePaths = nil
	)
	args, err := lib.CliParser()
	if err != nil {
		goto exitErr
	}
	if args.Verbose {
		utils.GVerbose = true
	}
	/*
		Create a channel to receive interrupt signals from the OS.
		The goroutine continuously listens for an interrupt signal
		(Ctrl+C) and handles the interruption.
	*/
	sigChan = make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt)
	go func() {
		for range sigChan {
			utils.SentinelExit(utils.SentinelExitParams{
				ExitCode:    0,
				ExitMessage: "\n\nG0oDBy3!",
				ExitError:   nil,
			})
		}
	}()
	/*
		Set up the HTTP client with a default timeout of 5 seconds
		or a custom timeout specified with the -t flag. If the -r flag
		is provided, the standard HTTP client will be ignored, and
		the client will be configured to route through TOR.
	*/
	httpClient, err = lib.HttpClientInit(&args)
	if err != nil {
		goto exitErr
	}
	localVersion = utils.GetCurrentLocalVersion()
	repoVersion = lib.GetCurrentRepoVersion(httpClient)
	utils.VersionCompare(repoVersion, localVersion)
	fmt.Fprintf(utils.GStdout, " ===[ Sentinel, Version: %s ]===\n\n", localVersion)
	utils.GDisplayCount = 0
	/*
		Initialize the output file paths and create the output
		directory if it does not already exist.
	*/
	filePaths, err = utils.FilePathInit(&args)
	if err != nil {
		goto exitErr
	}
	fmt.Fprint(utils.GStdout, "[*] Method: ")
	if len(args.WordlistPath) == 0 {
		// Perform enumeration using external resources
		fmt.Fprintln(utils.GStdout, "PASSIVE")
		lib.PassiveEnum(&args, httpClient, filePaths)
	} else {
		// Perform enumeration using brute force
		fmt.Fprintln(utils.GStdout, "ACTIVE")
		lib.ActiveEnum(&args, httpClient, filePaths)
	}
	/*
		Save the summary (including IPv4, IPv6, ports if requested,
		and subdomains) in JSON format within the output directory.
	*/
	lib.WriteJSON(filePaths.FilePathJSON)
	utils.SentinelExit(utils.SentinelExitParams{
		ExitCode:    0,
		ExitMessage: "",
		ExitError:   nil,
	})
exitErr:
	fmt.Fprintln(utils.GStdout, err)
	utils.GStdout.Flush()
	utils.Glogger.Println(err)
	utils.Glogger.Fatalf("Program execution failed")
}
