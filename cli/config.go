package cli

import (
	"github.com/shuntaka9576/oax"
)

func Config(editor string, setting bool, profile bool) error {
	err := oax.CmdConfig(editor, setting, profile)
	if err != nil {
		return err
	}

	return nil
}
