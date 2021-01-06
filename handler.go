package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	"github.com/iter8-tools/iter8-kfserving-handler/v1alpha2"
	"github.com/iter8-tools/iter8-kfserving-handler/v1beta1"
)

// stdin enables dependency injection for console input (stdin)
var stdin io.Reader

// stdout enables dependency injection for console output (stdout)
var stdout io.Writer

// stderr enables dependency injection for console error output (stderr)
var stderr io.Writer

// init initializes stdin/out/err, and logging.
func init() {
	// stdio
	stdin = os.Stdin
	stdout = os.Stdout
	stderr = os.Stderr
	// logging
	logLevel, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.SetOutput(ioutil.Discard)
	} else {
		log.SetOutput(stdout)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})
		log.SetReportCaller(true)
		log.SetLevel(logLevel)
	}
}

// main serves as the entry point for handler CLI.
func main() {
	// h := handler.Builder(stdin, stdout, stderr)
	if len(os.Args) < 2 {
		fmt.Fprintln(stderr, "expected 'start' or 'finish' subcommands")
	} else if os.Args[1] == "start" || os.Args[1] == "finish" {
		/* Everywhere below, if you encounter error, print error and exit */
		exp, _ := experiment.GetExperiment()
		targetRef := exp.GetTarget()
		targetType := target.GetTargetType(targetRef)
		var targ target.Target
		if targetType == target.V1alpha2 {
			targ = v1alpha2.GetTarget(targetRef)
		} else {
			targ = v1beta1.GetTarget(targetRef)
		}
		if os.Args[1] == "start" {
			targ.InitializeTrafficSplit(exp)
			exp.SetVersionInfo(targ.GetVersionInfo(exp))
		} else { // finish
			if exp.IsSingleVersion() {
				os.Exit(0)
			}
			oldBaseline, err := exp.GetBaseline()
			if err != nil {
				// scream
				os.Exit(1)
			}
			recommendedBaseline, err := exp.GetRecommendedBaseline()
			if err == nil {
				targ.SetNewBaseline(recommendedBaseline)
			} else {
				targ.SetNewBaseline(oldBaseline)
			}
		}
		fmt.Fprintln(stdout, "all good")
	} else {
		fmt.Fprintln(stderr, "expected 'start' or 'finish' subcommands")
	}
}
