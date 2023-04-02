package oax

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"

	"github.com/itchyny/timefmt-go"
	"github.com/pelletier/go-toml"
	"github.com/shuntaka9576/oax/openai"
)

type ChatMessage struct {
	Role    string `toml:"role"`
	Content string `toml:"content"`
}

type ChatLogToml struct {
	Messages []ChatMessage `toml:"messages"`
}

type ChatLog struct {
	ConfigDir   string
	ChatLogToml ChatLogToml
	FilePath    *string
}

func (c *ChatLog) AddChatMessage(chatMessage ChatMessage) *ChatLog {
	c.ChatLogToml.Messages = append(c.ChatLogToml.Messages, chatMessage)

	return c
}

func (c *ChatLog) FilePathForUser() (string, error) {
	valuePath := *c.FilePath
	replaced, err := replaceHomedirWithTilde(valuePath)
	if err != nil {
		return "", err
	}

	return replaced, nil
}

func (c *ChatLog) DeleteFile() error {
	err := os.Remove(*c.FilePath)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatLog) LoadLogMessage() error {
	data, err := ioutil.ReadFile(*c.FilePath)
	if err != nil {
		return err
	}

	var chatLogToml ChatLogToml
	err = toml.Unmarshal(data, &chatLogToml)
	if err != nil {
		return err
	}

	c.ChatLogToml.Messages = chatLogToml.Messages

	for i := range c.ChatLogToml.Messages {
		c.ChatLogToml.Messages[i].Content = strings.TrimRightFunc(c.ChatLogToml.Messages[i].Content, unicode.IsSpace)
	}

	return nil
}

func (c *ChatLog) CreateOpenAIMessages() (messages []openai.Message) {
	for _, message := range c.ChatLogToml.Messages {
		messages = append(messages, openai.Message{
			Content: message.Content,
			Role:    message.Role,
		})
	}

	return messages
}

func (c *ChatLog) FlushFile() error {
	var builder strings.Builder

	for _, message := range c.ChatLogToml.Messages {
		builder.WriteString(fmt.Sprintf(`[[messages]]
  role = "%s"
  content = '''
%s
'''

`, message.Role, message.Content))
	}

	err := ioutil.WriteFile(*c.FilePath, []byte(builder.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (c *ChatLog) InitLogFile(title string, fileNameFormat string) {
	var useFormat string

	if fileNameFormat == "" {
		if title == "" {
			useFormat = "%Y-%m-%d_%H-%M-%S"
		} else {
			useFormat = "%Y-%m-%d_%H-%M-%S_${title}"
		}
	} else {
		useFormat = fileNameFormat
	}

	useFormat = strings.ReplaceAll(useFormat, "${title}", title)

	t := time.Now()
	filename := timefmt.Format(t, useFormat) + ".toml"

	filePath := filepath.Join(c.ConfigDir, filename)
	c.FilePath = &filePath
}

func (c *ChatLog) LoadFile(filePath string) error {
	filePath, err := replaceTildeWithHomedir(filePath)
	if err != nil {
		return err
	}

	c.FilePath = &filePath

	return nil
}
