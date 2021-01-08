package main_test

import (
	"bytes"
	"os"
	"os/exec"

	utils "github.com/iter8-tools/iter8ctl/utils"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Handler", func() {

	Describe("CLI", func() {
		// All subtests within this test rely in ./iter8ctl. So go build.
		exec.Command("go", "build", utils.CompletePath("", "handler.go")).Run()

		Context("With no subcommand", func() {
			cmd := exec.Command(utils.CompletePath("", "handler"))
			cmd.Env = append(os.Environ(), "LOG_LEVEL=info")
			outbuf, errbuf := bytes.Buffer{}, bytes.Buffer{}
			cmd.Stderr = &errbuf
			cmd.Stdout = &outbuf

			It("should result in an error message", func() {
				cmd.Run()
				Expect(outbuf.String()).Should(ContainSubstring("expected 'start' or 'finish' subcommands"))
			})
		})

		Context("With invalid subcommand", func() {
			cmd := exec.Command(utils.CompletePath("", "handler"), "invalid")
			cmd.Env = append(os.Environ(), "LOG_LEVEL=info")
			outbuf, errbuf := bytes.Buffer{}, bytes.Buffer{}
			cmd.Stderr = &errbuf
			cmd.Stdout = &outbuf

			It("should result in an error message", func() {
				cmd.Run()
				Expect(outbuf.String()).Should(ContainSubstring("expected 'start' or 'finish' subcommands"))
			})
		})

		Context("With 'start' subcommand and no env variables", func() {
			cmd := exec.Command(utils.CompletePath("", "handler"), "start")
			cmd.Env = append(os.Environ(), "LOG_LEVEL=info")
			outbuf, errbuf := bytes.Buffer{}, bytes.Buffer{}
			cmd.Stderr = &errbuf
			cmd.Stdout = &outbuf

			It("should result in an error message", func() {
				cmd.Run()
				Expect(outbuf.String()).Should(ContainSubstring("environment variables EXPERIMENT_NAME and EXPERIMENT_NAMESPACE need to be set to valid values"))
			})
		})

	})
})
