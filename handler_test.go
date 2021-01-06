package main_test

import (
	"bytes"
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

			outbuf, errbuf := bytes.Buffer{}, bytes.Buffer{}
			cmd.Stderr = &errbuf
			cmd.Stdout = &outbuf

			It("should result in an error message", func() {
				cmd.Run()
				Expect(outbuf.String()).Should(ContainSubstring("expected 'start' or 'finish' subcommands"))
			})
		})

	})
})
