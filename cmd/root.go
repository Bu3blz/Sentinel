package cmd

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/PlagueByteSec/sdakit-project/v2/internal/cli"
	utils "github.com/PlagueByteSec/sdakit-project/v2/internal/coreutils"
	"github.com/PlagueByteSec/sdakit-project/v2/internal/logging"
	"github.com/PlagueByteSec/sdakit-project/v2/internal/requests"
	"github.com/PlagueByteSec/sdakit-project/v2/internal/shared"
	"github.com/PlagueByteSec/sdakit-project/v2/internal/streams"
)

func methodManager(args shared.Args, httpClient *http.Client, filePaths *shared.FilePaths) {
	// Manager for subdomain enumeration methods that require and HTTP client
	methods := MethodManagerInit()
	for key, method := range methods {
		switch key {
		case shared.Passive: // Request endpoints (certificate transparency logs etc.)
			if utils.IsPassiveEnumeration(&args) {
				utils.PrintMethod(method.MethodKey)
				method.Action(&args, httpClient, filePaths)
				shared.GIsExec++
			}
		case shared.Active: // Brute-force by evaluating HTTP codes
			if utils.IsActiveEnumeration(&args) {
				utils.PrintMethod(method.MethodKey)
				method.Action(&args, httpClient, filePaths)
				shared.GIsExec++
			}
		case shared.Dns: // Try to resolve a list of subdomains to IP addresses
			if utils.IsDnsEnumeration(&args) {
				utils.PrintMethod(method.MethodKey)
				method.Action(&args, httpClient, filePaths)
				shared.GIsExec++
			}
		case shared.VHost:
			if utils.IsVHostEnumeration(&args) {
				utils.PrintMethod(method.MethodKey)
				method.Action(&args, httpClient, filePaths)
				shared.GIsExec++
			}
		}
	}
	// Manager for commands that require (.txt) lists containing addresses
	extern := ValidsManagerInit()
	for key, method := range extern {
		switch key {
		case shared.RDns: // Resolving addresses from IP list
			if utils.IsRDnsEnumeration(&args) {
				fmt.Fprintln(shared.GStdout, shared.RDns)
				method.Action(&args)
			}
		case shared.Ping: // Ping subdomains from subdomain list
			if utils.IsPingFromFile(&args) {
				fmt.Fprintln(shared.GStdout, shared.Ping)
				method.Action(&args)
			}
		case shared.HeaderAnalysis:
			if utils.IsHttpHeaderAnalysis(&args) {
				fmt.Fprintln(shared.GStdout, shared.HeaderAnalysis)
				method.Action(&args)
			}
		}
	}
}

func Run(args shared.Args) {
	defer logging.GLogger.Stop()
	shared.GVerbose = args.Verbose
	shared.GTargetDomain = args.Domain
	var filePaths *shared.FilePaths = nil
	InterruptListenerStart()
	/*
		Set up the HTTP client with a default timeout of 5 seconds
		or a custom timeout specified with the -t flag. If the -r flag
		is provided, the standard HTTP client will be ignored, and
		the client will be configured to route through TOR.
	*/
	httpClient, err := requests.HttpClientInit(&args)
	if err != nil {
		utils.ProgramExit(utils.ExitParams{
			ExitCode:    -1,
			ExitMessage: "HttpClientInit failed",
			ExitError:   err,
		})
	}
	// Print banner and compare local with repo version
	utils.PrintBanner(httpClient)
	shared.GDisplayCount = 0
	// assign settings to global output switches directly
	shared.GShowAllHeaders = args.ShowAllHeaders
	shared.GDisableAllOutput = args.DisableAllOutput
	// allow redirects if misonfiguration test is enabled
	args.AllowRedirects = args.MisconfTest
	if !args.DisableAllOutput && args.Domain != "" {
		/*
			Initialize the output file paths and create the output
			directory if it does not already exist.
		*/
		filePaths, err = streams.FilePathInit(&args)
		if err != nil {
			utils.ProgramExit(utils.ExitParams{
				ExitCode:    -1,
				ExitMessage: "FilePathInitInit failed",
				ExitError:   err,
			})
		}
	}
	utils.PrintVerbose("[*] HTTP request method: %s\n", args.HttpRequestMethod)
	methodManager(args, httpClient, filePaths)
	if shared.GIsExec == 0 {
		utils.ProgramExit(utils.ExitParams{
			ExitCode:    -1,
			ExitMessage: cli.HelpBanner,
			ExitError:   errors.New("failed to start enum: syntax error"),
		})
	}
	/*
		Save the summary (including IPv4, IPv6, ports if requested,
		and subdomains) in JSON format within the output directory.
	*/
	if !shared.GDisableAllOutput {
		streams.WriteJSON(filePaths.FilePathJSON)
	}
	utils.ProgramExit(utils.ExitParams{
		ExitCode:    0,
		ExitMessage: "",
		ExitError:   nil,
	})
}
