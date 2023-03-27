package oax

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
	"unicode"

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
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	homeDir := usr.HomeDir

	if strings.HasPrefix(*c.FilePath, homeDir) {
		return strings.Replace(*c.FilePath, homeDir, "~", 1), nil
	}

	return *c.FilePath, nil
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
	if c.FilePath == nil {
		filename := time.Now().Format("2006-01-02_15-04-05") + ".toml"
		filePath := filepath.Join(c.ConfigDir, filename)
		c.FilePath = &filePath
	}

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

func (c *ChatLog) LoadFile(filePath string) error {

	if strings.HasPrefix(filePath, "~") {
		usr, err := user.Current()
		if err != nil {
			return err
		}
		homeDir := usr.HomeDir

		path := strings.Replace(*c.FilePath, "~", homeDir, 1)
		c.FilePath = &path
	}
	c.FilePath = &filePath

	return nil
}
