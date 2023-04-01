package oax

import (
	"os"
	"os/exec"
)

type Editor interface {
	Open(filename string) error
}

type CMDEditor struct {
	cmdName string
}

func InitEditor(cmdName string) Editor {
	editor := CMDEditor{
		cmdName: cmdName,
	}

	return editor
}

func (ce CMDEditor) Open(filePath string) error {
	c := exec.Command(ce.cmdName, filePath)

	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	return c.Run()
}
