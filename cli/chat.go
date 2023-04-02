package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/shuntaka9576/oax"
	"github.com/shuntaka9576/oax/openai"
)

type ChatOption struct {
	APIKey         string
	OrganizationID string
	Editor         string
	Model          string
	Role           string
	ChatLogDir     string
	FileNameFormat string
	File           *string
	Template       *oax.ChatTemplate
}

var (
	contentUserDefault = "# Remove this comment and specify content to send to OpenAI API; otherwise, nothing is sent."
)

func Chat(opt *ChatOption) error {
	if opt.Role == "" {
		opt.Role = "user"
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Title: ")
	title, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	title = title[:len(title)-1]

	userEmptyMessage := oax.ChatMessage{
		Role:    opt.Role,
		Content: contentUserDefault,
	}

	chatLog := oax.ChatLog{
		ConfigDir:   opt.ChatLogDir,
		ChatLogToml: oax.ChatLogToml{},
	}

	if opt.File == nil {
		chatLog.InitLogFile(title, opt.FileNameFormat)
		chatLog.FlushFile()
	} else {
		if err := chatLog.LoadFile(*opt.File); err != nil {
			return err
		}
		err := chatLog.LoadLogMessage()
		if err != nil {
			return err
		}
	}

	if !isLastEmptyMessage(chatLog.ChatLogToml.Messages) {
		if opt.File == nil && opt.Template != nil {
			for _, message := range opt.Template.Messages {
				chatLog.AddChatMessage(
					oax.ChatMessage(message),
				)
			}
		}

		chatLog.AddChatMessage(userEmptyMessage)
	}

	err = chatLog.FlushFile()
	if err != nil {
		return err
	}

	editor := oax.InitEditor(opt.Editor)

	err = editor.Open(*chatLog.FilePath)
	if err != nil {
		return err
	}

	err = chatLog.LoadLogMessage()
	if err != nil {
		return err
	}

	err = chatLog.FlushFile()
	if err != nil {
		return err
	}

	if len(chatLog.ChatLogToml.Messages) == 1 && isLastEmptyMessage(chatLog.ChatLogToml.Messages) {
		if err := deleteFile(chatLog); err != nil {
			return err
		}

		return nil
	} else if isLastEmptyMessage(chatLog.ChatLogToml.Messages) {
		fmt.Fprintf(os.Stderr, "detected default comment, terminating process.")

		return nil
	}

	openaiClient := openai.InitClient(&openai.InitClientOptions{
		APIKey:         opt.APIKey,
		OrganizationID: opt.OrganizationID,
	})

	bufFromChatGPT := bytes.Buffer{}
	chatGPTChatMessage := oax.ChatMessage{}
	multiWriter := io.MultiWriter(&bufFromChatGPT, os.Stdout)

	ctx := context.Background()

	subScribeChat := func(event *openai.ChatCompletionResponse, err error) error {
		if err != nil {
			if err == io.EOF {
				return nil
			} else {
				if errors.Is(err, openai.ErrorOpenAIUnauthorized) {
					fmt.Fprintf(os.Stderr, "%s. Please check if the API key is correct using `oax config --profiles`.\n", err)
				}
				fmt.Fprintf(os.Stderr, "%s`.\n", err)

				return err
			}

		} else {
			if event.Choices[0].Delta.Role != "" {
				chatGPTChatMessage.Role = event.Choices[0].Delta.Role
			}
			if event.Choices[0].Delta.Content != "" {
				fmt.Fprintf(multiWriter, "%v", event.Choices[0].Delta.Content)
			}

			return nil
		}
	}

LOOP:
	for {
	INTERACTIVE:
		for {
			if isSkip := isLastEmptyMessage(chatLog.ChatLogToml.Messages); !isSkip {
				err := openaiClient.ChatCreateCompletionSubscribeWithContext(ctx, &openai.ChatCreateCompletionOption{
					Messages: chatLog.CreateOpenAIMessages(),
					Model:    opt.Model,
				}, subScribeChat)

				if err != nil {
					return err
				}

				chatGPTChatMessage.Content = bufFromChatGPT.String()
				chatLog.AddChatMessage(chatGPTChatMessage)
				chatLog.FlushFile()
			} else {
				break LOOP
			}

			reader := bufio.NewReader(os.Stdin)
			fmt.Print("\n\n")
			fmt.Print("continue (y/n)?: ")

			input, err := reader.ReadString('\n')
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
				continue
			}

			input = strings.ToLower(strings.TrimSpace(input))
			if input == "y" || input == "Y" {
				chatLog.
					AddChatMessage(userEmptyMessage)

				err = chatLog.FlushFile()
				if err != nil {
					return err
				}
				bufFromChatGPT = bytes.Buffer{}

				err = editor.Open(*chatLog.FilePath)
				if err != nil {
					return err
				}

				err = chatLog.LoadLogMessage()
				if err != nil {
					return err
				}

				break INTERACTIVE
			}

			if input == "n" || input == "N" {
				break LOOP
			}
		}
	}

	if len(chatLog.ChatLogToml.Messages) == 1 && isLastEmptyMessage(chatLog.ChatLogToml.Messages) {
		if err := deleteFile(chatLog); err != nil {
			return err
		}
	} else {
		filePathForUser, err := chatLog.FilePathForUser()
		if err != nil {
			return err
		}

		if opt.File == nil {
			fmt.Fprintf(os.Stderr, "saved: %s\n", filePathForUser)
		} else {
			fmt.Fprintf(os.Stderr, "updated: %s\n", filePathForUser)
		}
	}

	return nil
}

func deleteFile(chatLog oax.ChatLog) error {
	filePathForUser, err := chatLog.FilePathForUser()
	if err != nil {
		return err
	}

	err = chatLog.DeleteFile()
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "skip delete file: %s\n", filePathForUser)

	return nil
}

func isLastEmptyMessage(messages []oax.ChatMessage) bool {
	if len(messages) > 0 {
		lastmsg := messages[len(messages)-1]

		return lastmsg.Content == contentUserDefault
	}

	return false
}
