package oax

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml"
)

type SettingTOML struct {
	Setting Setting `toml:"setting"`
}

type Setting struct {
	Editor               string `toml:"editor"`
	ChatLogDir           string `toml:"chatLogDir"`
	InferChatLogFileName string `toml:"inferChatLogFileName"`
}

type Profile struct {
	Name           string `toml:"name"`
	Description    string `toml:"description"`
	ApiKey         string `toml:"apiKey"`
	OrganizationID string `toml:"organizationId"`
	Default        bool   `toml:"default"`
}

type ProfileToml struct {
	Profiles []Profile `toml:"profiles"`
}

type Config struct {
	Profiles []Profile
	Setting  Setting
}

var (
	settingFilePath       string
	profileFilePath       string
	chatLogDirDefaultPath string
)

func init() {
	configDir, err := getConfigDir()
	if err != nil {
		panic(err)
	}

	settingFilePath = filepath.Join(configDir, "settings.toml")
	profileFilePath = filepath.Join(configDir, "profiles.toml")
	chatLogDirDefaultPath = filepath.Join(configDir, "chat-log")
}

func GetConfig() (*Config, error) {
	setting, err := loadSetting()
	if err != nil {
		return nil, fmt.Errorf("error load setting: %w", err)
	}

	profiles, err := loadProfiles()
	if err != nil {
		return nil, fmt.Errorf("error load profile: %w", err)
	}

	err = createIfNotExistChatLogDir(setting.Setting.ChatLogDir)
	if err != nil {
		return nil, fmt.Errorf("error create chat log dir: %w", err)
	}

	if setting.Setting.ChatLogDir == "" {
		setting.Setting.ChatLogDir = chatLogDirDefaultPath
	}

	return &Config{
		Setting:  setting.Setting,
		Profiles: profiles.Profiles,
	}, nil
}

func getConfigDir() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}

	var configDir string

	if runtime.GOOS == "windows" {
		configDir = filepath.Join(usr.HomeDir, "AppData", "Roaming", "oax")
	} else {
		configDir = filepath.Join(usr.HomeDir, ".config", "oax")
	}

	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

func loadSetting() (*SettingTOML, error) {
	_, err := os.Stat(settingFilePath)

	var configTree *toml.Tree

	if os.IsNotExist(err) {
		settingStr := `[setting]
  editor = "vim"`

		err = ioutil.WriteFile(settingFilePath, []byte(settingStr), 0644)
		if err != nil {
			return nil, err
		}

		configTree, err = toml.Load(settingStr)
		if err != nil {
			return nil, err
		}
	} else {
		configTree, err = toml.LoadFile(settingFilePath)
		if err != nil {
			return nil, err
		}
	}

	setting := &SettingTOML{}
	err = configTree.Unmarshal(setting)
	if err != nil {
		return nil, err
	}

	return setting, nil
}

func loadProfiles() (*ProfileToml, error) {
	_, err := os.Stat(profileFilePath)

	var configTree *toml.Tree

	if os.IsNotExist(err) {
		settingStr := `[[profiles]]
name = "personal"
apiKey = "sk-xxxx"
default = true
`

		err = ioutil.WriteFile(profileFilePath, []byte(settingStr), 0644)
		if err != nil {
			return nil, err
		}

		configTree, err = toml.Load(settingStr)
		if err != nil {
			return nil, err
		}
	} else {
		configTree, err = toml.LoadFile(profileFilePath)
		if err != nil {
			return nil, err
		}
	}

	profile := &ProfileToml{}
	err = configTree.Unmarshal(profile)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func createIfNotExistChatLogDir(chatLogDir string) error {
	if chatLogDir == "" {
		chatLogDir = chatLogDirDefaultPath
	}

	if err := createDirIfNotExist(chatLogDir); err != nil {
		return err
	}

	return nil
}

func CmdConfig(editor string, setting bool, profile bool) error {
	var cmd *exec.Cmd
	if setting {
		cmd = exec.Command(editor, settingFilePath)
	} else {
		cmd = exec.Command(editor, profileFilePath)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
