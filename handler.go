package main

import (
	"io"
	"io/ioutil"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/iter8-tools/iter8-kfserving-handler/experiment"
	"github.com/iter8-tools/iter8-kfserving-handler/target"
	"github.com/iter8-tools/iter8-kfserving-handler/v1alpha2"
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

// // K8sClient interface enables getting a k8s client
// type K8sClient interface {
// 	getK8sClient(kubeconfigPath *string) (runtimeclient.Client, error)
// }
// type iter8ctlK8sClient struct{}

// func (k iter8ctlK8sClient) getK8sClient(kubeconfigPath *string) (runtimeclient.Client, error) {
// 	crScheme := runtime.NewScheme()
// 	err := v2alpha1.AddToScheme(crScheme)
// 	if err != nil {
// 		return nil, err
// 	}
// 	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfigPath)
// 	if err != nil {
// 		return nil, err
// 	}
// 	rc, err := runtimeclient.New(config, client.Options{
// 		Scheme: crScheme,
// 	})
// 	if err != nil {
// 		return nil, err
// 	}
// 	return rc, nil
// }

// var k8sClient K8sClient

// stdin enables dependency injection for console input (stdin)
var stdin io.Reader

// stdout enables dependency injection for console output (stdout)
var stdout io.Writer

// stderr enables dependency injection for console error output (stderr)
var stderr io.Writer

// init initializes stdin/out/err, logging, and osExiter.
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
}

// main serves as the entry point for handler CLI.
func main() {
	// h := handler.Builder(stdin, stdout, stderr)
	if len(os.Args) < 2 {
		log.Error("expected 'start' or 'finish' subcommands")
		osExiter.Exit(1)
	} else if os.Args[1] == "start" || os.Args[1] == "finish" {
		exp, err := experiment.GetExperiment()
		if err != nil {
			log.Error("cannot get experiment", err)
			osExiter.Exit(1)
		}
		targetRef := exp.GetTargetRef()
		targetType := target.GetTargetType(targetRef)
		var targ target.Target
		if targetType == target.V1alpha2 {
			targ = v1alpha2.TargetBuilder()
		} else {
			targ = v1beta1.TargetBuilder()
		}
		targ.SetExperiment(exp).GetTarget()

		if os.Args[1] == "start" {
			targ.InitializeTrafficSplit().GetVersionInfo().SetVersionInfoInExperiment()
		} else { // finish
			if exp.IsSingleVersion() {
				osExiter.Exit(0)
			}
			targ.GetOldBaseline().GetNewBaseline().SetNewBaseline()
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
