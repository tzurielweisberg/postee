package actions

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/tzurielweisberg/postee/v2/formatting"

	"github.com/tzurielweisberg/postee/v2/layout"
)

type execCmd = func(string, ...string) *exec.Cmd

type ExecClient struct {
	ExecCmd    execCmd
	Name       string
	Env        []string
	InputFile  string
	ExecScript string
	Action     []byte
}

func (e *ExecClient) GetName() string {
	return e.Name
}

func (e *ExecClient) Init() error {
	e.ExecCmd = exec.Command
	return nil
}

func (e *ExecClient) Send(m map[string]string) error {
	envVars := os.Environ()
	envVars = append(envVars, e.Env...)
	envVars = append(envVars, fmt.Sprintf("POSTEE_EVENT=%s", m["description"]))

	var cmd *exec.Cmd
	if len(e.InputFile) > 0 {
		cmd = e.ExecCmd("/bin/sh", e.InputFile)
		cmd.Env = append(cmd.Env, envVars...)
	}
	if len(e.ExecScript) > 0 {
		cmd = e.ExecCmd("/bin/sh")
		cmd.Env = append(cmd.Env, envVars...)
		cmd.Stdin = strings.NewReader(e.ExecScript)
	}

	var err error
	if e.Action, err = cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("error while executing script: %w, output: %s", err, string(e.Action))
	}
	log.Println("execution output: ", "len: ", len(e.Action), "out: ", string(e.Action))
	return nil
}

func (e *ExecClient) Terminate() error {
	log.Printf("Exec action %s terminated\n", e.GetName())
	return nil
}

func (e *ExecClient) GetLayoutProvider() layout.LayoutProvider {
	// Todo: This is MOCK. Because Formatting isn't need for Webhook
	// todo: The App should work with `return nil`
	return new(formatting.HtmlProvider)
}
