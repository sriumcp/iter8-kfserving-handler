package main

import (
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/k8sclient"
	"github.com/iter8-tools/iter8-kfserving-handler/v1beta1"
)

// OSExiter interface enables exiting the current program.
type OSExiter interface {
	Exit(code int)
}
type iter8OS struct{}

func (m *iter8OS) Exit(code int) {
	os.Exit(code)
}

var osExiter OSExiter

var k8s k8sclient.K8s

// stdin enables dependency injection for console input (stdin)
var stdin io.Reader

// stdout enables dependency injection for console output (stdout)
var stdout io.Writer

// stderr enables dependency injection for console error output (stderr)
var stderr io.Writer

// init initializes stdin/out/err, logging, osExiter, and k8s.
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
	// osExiter
	osExiter = &iter8OS{}
	// k8s
	k8s = &k8sclient.Iter8K8s{}
}

// main serves as the entry point for handler CLI.
func main() {
	// h := handler.Builder(stdin, stdout, stderr)
	if len(os.Args) < 2 {
		log.Error("expected 'start' or 'finish' subcommands")
		osExiter.Exit(1)
	} else if os.Args[1] == "start" || os.Args[1] == "finish" {
		client, err := k8s.GetClient()
		exp, err := experiment.GetExperiment(client)
		if err != nil {
			log.Error("cannot get experiment", err)
			osExiter.Exit(1)
		}
		targetRef := exp.GetTargetRef()
		var targ = v1beta1.TargetBuilder()
		targ.SetK8sClient(client).SetExperiment(exp).Fetch(targetRef)
		if os.Args[1] == "start" { // handle start
			targ.InitializeTrafficSplit().SetVersionInfoInExperiment()
		} else { // handle finish
			if exp.IsSingleVersion() {
				osExiter.Exit(0)
			}
			targ.SetNewBaseline()
		}
		if targ.Error() != nil {
			log.Error(targ.Error())
			osExiter.Exit(1)
		}
	} else {
		log.Error("expected 'start' or 'finish' subcommands")
		osExiter.Exit(1)
	}
}
