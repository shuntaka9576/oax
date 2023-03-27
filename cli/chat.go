package cli

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
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
	File           *string
	Template       *oax.ChatTemplate
}

var (
	contentUserDefault = "# Remove this comment and specify content to send to OpenAI API; otherwise, nothing is sent."
)

func Chat(opt *ChatOption) error {
	if opt.File == nil {
		return ChatNew(opt)
	} else {
		return ChatWithFile(opt)
	}
}

func ChatNew(opt *ChatOption) error {
	if opt.Role == "" {
		opt.Role = "user"
	}

	userEmptyMessage := oax.ChatMessage{
		Role:    opt.Role,
		Content: contentUserDefault,
	}

	chatLog := oax.ChatLog{
		ConfigDir:   opt.ChatLogDir,
		ChatLogToml: oax.ChatLogToml{},
	}
	if opt.Template != nil {
		for _, message := range opt.Template.Messages {
			chatLog.AddChatMessage(
				oax.ChatMessage(message),
			)
		}
	}

	chatLog.AddChatMessage(userEmptyMessage)

	err := chatLog.FlushFile()
	if err != nil {
		return err
	}

	doneCh := make(chan struct{})
	errCh := make(chan error)
	messageCh := make(chan openai.Message)

	openaiClient := openai.InitClient(&openai.InitClientOptions{
		APIKey: opt.APIKey,
		ErrCh:  errCh,
	})

	cmd := exec.Command(opt.Editor, *chatLog.FilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
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

	messages := chatLog.CreateOpenAIMessages()

	if isSkip := isSkipRequest(messages); !isSkip {
		go func() {
			openaiClient.ChatCreateCompletion(&openai.ChatCreateCompletionOption{
				Model:     opt.Model,
				Messages:  messages,
				MessageCh: messageCh,
				DoneCh:    doneCh,
				ErrCh:     errCh,
			})
		}()

		var bufFromChatGPT bytes.Buffer
		chatGPTChatMessage := oax.ChatMessage{}
		multiWriter := io.MultiWriter(&bufFromChatGPT, os.Stdout)

	LOOP:
		for {
		INTERACTIVE:
			select {
			case content := <-messageCh:
				if content.Role != "" {
					chatGPTChatMessage.Role = content.Role
				}
				fmt.Fprintf(multiWriter, "%s", content.Content)
			case err := <-errCh:

				if errors.Is(err, openai.ErrorOpenAIUnauthorized) {
					fmt.Fprintf(os.Stderr, "%s. Please check if the API key is correct using `oax config --profiles`.\n", err)
				} else {
					fmt.Fprintf(os.Stderr, "%s", err)
				}

				return err
			case <-doneCh:
				chatGPTChatMessage.Content = bufFromChatGPT.String()
				chatLog.
					AddChatMessage((chatGPTChatMessage))

				err = chatLog.FlushFile()
				if err != nil {
					return err
				}

				reader := bufio.NewReader(os.Stdin)
				for {
					fmt.Print("\n\n")
					fmt.Print("continue (y/n)?: ")

					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("Error reading input: %v\n", err)
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
						chatGPTChatMessage = oax.ChatMessage{}

						cmd := exec.Command(opt.Editor, *chatLog.FilePath)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err = cmd.Run()
						if err != nil {
							log.Fatal(err)
						}

						err = chatLog.LoadLogMessage()
						if err != nil {
							return err
						}

						messages := chatLog.CreateOpenAIMessages()
						if isSkip = isSkipRequest(messages); !isSkip {
							go func() {
								openaiClient.ChatCreateCompletion(&openai.ChatCreateCompletionOption{
									Model:     opt.Model,
									Messages:  messages,
									MessageCh: messageCh,
									DoneCh:    doneCh,
									ErrCh:     errCh,
								})
							}()
						} else {
							break LOOP
						}

						break INTERACTIVE
					}

					if input == "n" || input == "N" {
						break LOOP
					}
				}
			}
		}
		filePathForUser, err := chatLog.FilePathForUser()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "saved: %s\n", filePathForUser)

		return nil
	} else {
		filePathForUser, err := chatLog.FilePathForUser()
		if err != nil {
			return err
		}
		err = chatLog.DeleteFile()
		if err != nil {
			return err
		}

		fmt.Fprintf(os.Stdout, "skip delete file: %s\n", filePathForUser)

		return nil
	}
}

func isSkipRequest(messages []openai.Message) bool {
	lastmsg := messages[len(messages)-1]

	return lastmsg.Content == contentUserDefault
}

func ChatWithFile(opt *ChatOption) error {
	if opt.Role == "" {
		opt.Role = "user"
	}

	userEmptyMessage := oax.ChatMessage{
		Role:    opt.Role,
		Content: contentUserDefault,
	}

	chatLog := oax.ChatLog{
		ConfigDir: opt.ChatLogDir,
	}

	err := chatLog.LoadFile(*opt.File)
	if err != nil {
		return err
	}

	err = chatLog.LoadLogMessage()
	if err != nil {
		return err
	}

	isSkip := isSkipRequest(chatLog.CreateOpenAIMessages())

	if !isSkip {
		chatLog.AddChatMessage(userEmptyMessage)
		err := chatLog.FlushFile()
		if err != nil {
			return err
		}
	}

	cmd := exec.Command(opt.Editor, *chatLog.FilePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return err
	}

	err = chatLog.LoadLogMessage()
	if err != nil {
		return err
	}

	messages := chatLog.CreateOpenAIMessages()

	doneCh := make(chan struct{})
	errCh := make(chan error)
	messageCh := make(chan openai.Message)

	openaiClient := openai.InitClient(&openai.InitClientOptions{
		APIKey: opt.APIKey,
		ErrCh:  errCh,
	})

	if isSkip := isSkipRequest(messages); !isSkip {
		go func() {
			openaiClient.ChatCreateCompletion(&openai.ChatCreateCompletionOption{
				Model:     opt.Model,
				Messages:  messages,
				MessageCh: messageCh,
				DoneCh:    doneCh,
				ErrCh:     errCh,
			})
		}()

		var bufFromChatGPT bytes.Buffer
		chatGPTChatMessage := oax.ChatMessage{}
		multiWriter := io.MultiWriter(&bufFromChatGPT, os.Stdout)

	LOOP:
		for {
		INTERACTIVE:
			select {
			case content := <-messageCh:
				if content.Role != "" {
					chatGPTChatMessage.Role = content.Role
				}
				fmt.Fprintf(multiWriter, "%s", content.Content)
			case err := <-errCh:
				return err
			case <-doneCh:
				chatGPTChatMessage.Content = bufFromChatGPT.String()
				chatLog.
					AddChatMessage((chatGPTChatMessage))

				err = chatLog.FlushFile()
				if err != nil {
					return err
				}

				reader := bufio.NewReader(os.Stdin)
				for {
					fmt.Print("\n\n")
					fmt.Print("continue (y/n)?: ")

					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("Error reading input: %v\n", err)
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
						chatGPTChatMessage = oax.ChatMessage{}

						cmd := exec.Command(opt.Editor, *chatLog.FilePath)
						cmd.Stdin = os.Stdin
						cmd.Stdout = os.Stdout
						cmd.Stderr = os.Stderr
						err = cmd.Run()
						if err != nil {
							log.Fatal(err)
						}

						err = chatLog.LoadLogMessage()
						if err != nil {
							return err
						}

						messages := chatLog.CreateOpenAIMessages()
						if isSkip = isSkipRequest(messages); !isSkip {
							go func() {
								openaiClient.ChatCreateCompletion(&openai.ChatCreateCompletionOption{
									Model:     opt.Model,
									Messages:  messages,
									MessageCh: messageCh,
									DoneCh:    doneCh,
									ErrCh:     errCh,
								})
							}()
						} else {
							break LOOP
						}

						break INTERACTIVE
					}

					if input == "n" || input == "N" {
						break LOOP
					}
				}
			}
		}
		filePathForUser, err := chatLog.FilePathForUser()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stdout, "updated: %s\n", filePathForUser)

		return nil
	}

	return nil
}
