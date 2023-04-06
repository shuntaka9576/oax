package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kong"
	"github.com/shuntaka9576/oax"
	"github.com/shuntaka9576/oax/cli"
)

type Globals struct {
	Profile string          `short:"p" name:"profile" help:"Specify the profile."`
	Version cli.VersionFlag `short:"v" name:"version" help:"Print the version."`
}

var CLI struct {
	Globals
	Config struct {
		Settings bool `help:"Open the settings configuration file."`
		Profiles bool `help:"Open the profiles configuration file."`
	} `cmd:"" help:"Provides a feature to check the OAX configuration settings"`
	Chat struct {
		Model        string  `short:"m" help:"Specify the ID of the model to use gpt-4, gpt-4-0314, gpt-4-32k, gpt-4-32k-0314, gpt-3.5-turbo, gpt-3.5-turbo-0301(default gpt-3.5-turbo)"`
		File         *string `short:"f" help:"Specify the chat history file with the full path."`
		TemplateName string  `short:"t" help:"Specify a chat template name."`
		Continue     bool    `short:"c" help:"Search your past chat history files with fuzzy matching and resume the chat from where you left off. This is an easier way to resume than using the --file option."`
	} `cmd:"" help:"Provides a dialogue function like chat.openai.com."`
}

func main() {
	config, err := oax.GetConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v", err)
	}

	kontext := kong.Parse(&CLI,
		kong.Name("oax"),
		kong.Description("CLI for OpenAI's ChatGPT."),
	)

	var useProfile oax.Profile
	if CLI.Profile == "" {
		for _, profile := range config.Profiles {
			if profile.Default {
				useProfile = profile
			}
		}
	} else {
		for _, profile := range config.Profiles {
			if profile.Name == CLI.Profile {
				useProfile = profile
			}
		}
	}

	if useProfile.Name == "" {
		fmt.Fprintf(os.Stderr, "invalid profile %s. Please check settings using `oax config --profiles`.\n", CLI.Profile)

		os.Exit(1)
	}

	var useTemplate *oax.ChatTemplate
	if CLI.Chat.TemplateName != "" {
		for _, template := range config.Settings.Chat.Templates {
			if CLI.Chat.TemplateName == template.Name {
				useTemplate = &template
			}
		}
	}

	switch kontext.Command() {
	case "config":
		err := cli.Config(config.Settings.Setting.Editor, CLI.Config.Settings, CLI.Config.Profiles)
		if err != nil {
			os.Exit(1)
		}
	case "chat":
		var model string
		defaultModel := "gpt-3.5-turbo"

		if CLI.Chat.Model == "" && config.Settings.Chat.Model == "" {
			model = defaultModel
		} else if CLI.Chat.Model != "" && config.Settings.Chat.Model == "" {
			model = CLI.Chat.Model
		} else if CLI.Chat.Model == "" && config.Settings.Chat.Model != "" {
			model = config.Settings.Chat.Model
		} else if CLI.Chat.Model != "" && config.Settings.Chat.Model != "" {
			model = CLI.Chat.Model
		}

		err := cli.Chat(&cli.ChatOption{
			APIKey:         useProfile.ApiKey,
			OrganizationID: useProfile.OrganizationID,
			Editor:         config.Settings.Setting.Editor,
			Model:          model,
			ChatLogDir:     config.Settings.Setting.ChatLogDir,
			FileNameFormat: config.Settings.Chat.FileNameFormat,
			File:           CLI.Chat.File,
			Template:       useTemplate,
			Continue:       CLI.Chat.Continue,
		})
		if err != nil {
			os.Exit(1)
		}
	}

}
