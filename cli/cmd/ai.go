package cmd

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/adrg/xdg"
	"github.com/chzyer/readline"
	"github.com/pkg/errors"
	"github.com/replicatedhq/replicated/pkg/kotsclient"
	"github.com/spf13/cobra"
)

var (
	ErrExit                  = fmt.Errorf("exit")
	ErrBundleNotFound        = fmt.Errorf("bundle not found")
	ErrIncompleteLoadCommand = fmt.Errorf("load command requires an argument")
)

const (
	maxCommandHistory = 100
)

func (r *runners) InitAICmd(parent *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ai",
		Hidden:       true,
		Short:        "",
		Long:         ``,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := r.loadAIArgs(); err != nil {
				return err
			}

			if err := r.AIInteractiveShell(); err != nil {
				return err
			}

			return nil
		},
	}
	parent.AddCommand(cmd)

	return cmd
}

func (r *runners) loadAIArgs() error {
	r.ai.model = "zephyr"

	return nil
}

func getHistoryFilePath() (string, error) {
	stateDir := filepath.Join(xdg.StateHome, "replicated")

	// Ensure the directory exists
	if err := os.MkdirAll(stateDir, os.ModePerm); err != nil {
		return "", err
	}

	// Return the full path to the history file
	return filepath.Join(stateDir, "history"), nil
}

func trimHistory(historyFile string, maxLines int) error {
	file, err := os.Open(historyFile)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}

	file.Close()

	err = os.WriteFile(historyFile, []byte(strings.Join(lines, "\n")+"\n"), 0644)
	if err != nil {
		return err
	}

	return nil
}

func (r *runners) AIInteractiveShell() error {
	historyFile, err := getHistoryFilePath()
	if err != nil {
		return err
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          ">>> ",
		HistoryFile:     historyFile,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil { // io.EOF
			break
		}

		if err := trimHistory(historyFile, maxCommandHistory); err != nil {
			log.Fatalf("Error trimming history: %v", err)
		}

		if err := r.aiProcessCommand(line); err != nil {
			if err == ErrExit {
				break
			}
		}
	}

	return nil
}

func (r *runners) aiProcessCommand(cmd string) error {
	if !strings.HasPrefix(cmd, "/") {
		return r.AIGPTCommand(cmd)
	}

	cmdParts := strings.Split(cmd, " ")
	switch cmdParts[0] {
	case "/exit":
		return ErrExit
	case "/help", "/?":
		return r.AIHelp()
	case "/patch":
		// this command supports subcommands
		if len(cmdParts) == 1 {
			return errors.New("patch command requires a subcommand")
		}

		switch cmdParts[1] {
		case "create":
			return r.AIPatchCreate()
		case "ls":
			return r.AIPatchList()
		default:
			return errors.Errorf("unknown patch subcommand %s", cmdParts[1])
		}
	case "/load":
		if len(cmdParts) != 2 {
			return ErrIncompleteLoadCommand
		}
		return r.AILoad(cmdParts[1])
	case "/unload":
		if r.ai.bundle == nil {
			fmt.Println("No bundle loaded")
			return nil
		}
		r.ai.bundle = nil
		return nil
	case "/bundle":
		if r.ai.bundle == nil {
			fmt.Println("No bundle loaded")
			return nil
		}
		return r.AIShowBundle()
	case "/model":
		if len(cmdParts) == 1 {
			fmt.Printf("Current model: %s\n", r.ai.model)
			return nil
		} else {
			if cmdParts[1] == "list" {
				fmt.Println("Available models:")
				fmt.Println("  zephyr")
				return nil
			}
			r.ai.model = cmdParts[1]
			// todo unload the bundle, reload the bundle?
			return nil
		}
	default:
		return r.AIUnknownCommand()
	}
}

func (r *runners) AIShowBundle() error {
	// refresh the bundle from the server
	bundle, err := r.kotsAPI.GetBundle(r.ai.bundle.ID)
	if err != nil {
		return errors.Wrap(err, "get bundle")
	}

	r.ai.bundle = bundle

	fmt.Printf("Bundle status: \n\tID: %s\n\tStatus: %s\n", r.ai.bundle.ID, r.ai.bundle.Status)
	return nil
}

func (r *runners) AIHelp() error {
	fmt.Println(`commands:
  /load          loads an bundle, either local or a replicated release
  /model [name]  shows or changes the model to use
  /apply         generates an actual .patch for the latest prompt
  /bundle 	  	 shows the current bundle status
  /help
  /exit`)

	return nil
}

func (r *runners) AIUnknownCommand() error {
	fmt.Println("unknown command")
	return nil
}

func (r *runners) AILoad(path string) error {
	if stat, err := os.Stat(path); err == nil {
		if stat.IsDir() {
			absolutePath, err := filepath.Abs(path)
			if err != nil {
				return errors.Wrap(err, "get absolute path")
			}

			fmt.Printf("loading local bundle from %s\n", absolutePath)

			archive, err := tarGzDir(path)
			if err != nil {
				return errors.Wrap(err, "tar dir")
			}

			bundle, err := r.kotsAPI.LoadBundle(r.ai.model, archive)
			if err != nil {
				if errors.Cause(err) == kotsclient.ErrAINotEntitled {
					fmt.Printf("You are not entitled to use this feature\n")
					return nil
				}

				return errors.Wrap(err, "load bundle")
			}

			r.ai.bundle = bundle

			fmt.Printf("Bundle is loading, check the status with /bundle command\n")
			return nil
		}
	}

	var appSlug string
	var channelSlug string
	var releaseSequence int64
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 2 {
		// maybe a replicated release
		appSlug = pathParts[0]
		channelSlug = pathParts[1]
	} else if len(pathParts) == 3 && pathParts[1] == "release" {
		appSlug = pathParts[0]
		r, err := strconv.Atoi(pathParts[2])
		if err == nil {
			releaseSequence = int64(r)
		}
	}

	if appSlug == "" {
		return ErrBundleNotFound
	}

	// try to load the bundle from replicated
	fmt.Printf("loading bundle %s / %s / %d\n", appSlug, channelSlug, releaseSequence)
	return nil
}

func (r *runners) AIPatchList() error {
	if r.ai.bundle == nil {
		fmt.Println("No bundle loaded")
		return nil
	}

	patches, err := r.kotsAPI.PatchListBundle(r.ai.bundle.ID)
	if err != nil {
		return errors.Wrap(err, "list bundle patches")
	}

	fmt.Println("Available patches:")
	for _, patch := range patches {
		fmt.Printf("  %s\n", patch.ID)
	}

	return nil
}

func (r *runners) AIPatchCreate() error {
	if r.ai.bundle == nil {
		fmt.Println("No bundle loaded")
		return nil
	}

	if r.ai.previousPrompt == "" || r.ai.previousResponse == "" {
		fmt.Println("No pending message to create a patch from")
		return nil
	}

	fmt.Println("Generating patch...")

	responseCh := make(chan string, 1)
	errCh := make(chan error)

	go func() {
		err := r.kotsAPI.BundleCreatePatch(r.ai.bundle.ID, r.ai.model, r.ai.previousPrompt, r.ai.previousResponse, responseCh)
		if err != nil {
			errCh <- err
			return
		}
	}()

	completeResponse := ""

	for {
		select {
		case response, ok := <-responseCh:
			if !ok {
				fmt.Println("Patch generated. To view app patches, run `/patch ls`")
				fmt.Println(completeResponse)
				return nil
			}

			completeResponse += response
		case err := <-errCh:
			return err
		}
	}
}

func (r *runners) AIGPTCommand(cmd string) error {
	if r.ai.bundle == nil {
		fmt.Println("No bundle loaded")
		return nil
	}

	r.ai.previousPrompt = ""
	r.ai.previousResponse = ""

	responseCh := make(chan string, 1)
	errCh := make(chan error)

	go func() {
		err := r.kotsAPI.BundlePrompt(r.ai.bundle.ID, r.ai.model, cmd, responseCh)
		if err != nil {
			errCh <- err
			return
		}
	}()

	completeResponse := ""
	for {
		select {
		case response, ok := <-responseCh:
			if !ok {
				r.ai.previousPrompt = cmd
				r.ai.previousResponse = completeResponse
				fmt.Println("\n\nUse the `/patch create` command to generate a patch based on this response.")
				return nil
			}

			completeResponse += response
			fmt.Print(response)
		case err := <-errCh:
			return err
		}
	}
}

func tarGzDir(srcDir string) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	gzWriter := gzip.NewWriter(&buf)
	tarWriter := tar.NewWriter(gzWriter)

	defer gzWriter.Close()
	defer tarWriter.Close()

	err := filepath.Walk(srcDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// Adjust the header name to be relative to the source directory
		relPath, err := filepath.Rel(srcDir, file)
		if err != nil {
			return err
		}
		header.Name = relPath

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if !fi.Mode().IsRegular() { // Skip non-regular files (e.g. directories)
			return nil
		}

		fileData, err := os.Open(file)
		if err != nil {
			return err
		}
		defer fileData.Close()

		if _, err := io.Copy(tarWriter, fileData); err != nil {
			return err
		}

		return nil
	})

	return &buf, err
}
