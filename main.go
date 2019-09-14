//go:generate go run vendor/github.com/Al2Klimov/go-gen-source-repos/main.go github.com/Al2Klimov/check_purgeable

package main

import (
	"bytes"
	"flag"
	"fmt"
	_ "github.com/Al2Klimov/go-gen-source-repos"
	. "github.com/Al2Klimov/go-monplug-utils"
	"html"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	os.Exit(ExecuteCheck(onTerminal, checkPurgeable))
}

func onTerminal() (output string) {
	return fmt.Sprintf(
		"For the terms of use, the source code and the authors\n"+
			"see the projects this program is assembled from:\n\n  %s\n",
		strings.Join(GithubcomAl2klimovGo_gen_source_repos, "\n  "),
	)
}

var deinstallLine = regexp.MustCompile(`\A(\S+)\s+deinstall\b`)

func checkPurgeable() (output string, perfdata PerfdataCollection, errs map[string]error) {
	cli := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	var warn, crit OptionalThreshold

	cli.Var(&warn, "warn", "e.g. @~:42")
	cli.Var(&crit, "crit", "e.g. @~:42")

	if errCli := cli.Parse(os.Args[1:]); errCli != nil {
		os.Exit(3)
	}

	cmd := exec.Command("dpkg", "--get-selections")
	var out bytes.Buffer

	cmd.Env = []string{"LC_ALL=C"}
	cmd.Dir = "/"
	cmd.Stdin = nil
	cmd.Stdout = &out
	cmd.Stderr = nil

	if errRun := cmd.Run(); errRun != nil {
		errs = map[string]error{"dpkg --get-selections": errRun}
		return
	}

	var packages []string

	for _, line := range bytes.Split(out.Bytes(), []byte{'\n'}) {
		if match := deinstallLine.FindSubmatch(line); match != nil {
			packages = append(packages, string(match[1]))
		}
	}

	perfdata = PerfdataCollection{Perfdata{
		Label: "packages",
		Value: float64(len(packages)),
		Warn:  warn,
		Crit:  crit,
		Min:   OptionalNumber{true, 0},
	}}

	if len(packages) > 0 {
		var buf bytes.Buffer

		buf.Write([]byte("<p><b>Some packages may be purged</b></p>\n\n<ul>"))

		for _, packag := range packages {
			buf.Write([]byte("<li>"))
			buf.Write([]byte(html.EscapeString(packag)))
			buf.Write([]byte("</li>"))
		}

		buf.Write([]byte("</ul>"))

		output = buf.String()
	} else {
		output = "<p>No packages to purge.</p>"
	}

	return
}
