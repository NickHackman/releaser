package edit

import (
	"bufio"
	_ "embed"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultEditor = "vim"
)

var (
	//go:embed template-instructions.txt
	TemplateInstructions string

	//go:embed manual-edit-instructions.txt
	ManualEditInstructions string
)

func runEditor(fileName string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = defaultEditor
	}

	var cmd *exec.Cmd
	switch editor {
	case "code":
		// VSCode must pass --wait flag otherwise it exits immediately
		cmd = exec.Command(editor, "--wait", fileName)
	default:
		cmd = exec.Command(editor, fileName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	return cmd.Run()
}

func createTempFile(content string) (string, error) {
	tempDir := os.TempDir()
	f, err := ioutil.TempFile(tempDir, "*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}

	if _, err := f.WriteString(content); err != nil {
		return "", fmt.Errorf("failed to write content to temp file %s: %v", f.Name(), err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close file %s: %v", f.Name(), err)
	}

	return f.Name(), nil
}

func Content(content, instructions string) (string, error) {
	fileContent := content + "\n\n" + instructions
	fileName, err := createTempFile(fileContent)
	if err != nil {
		return "", err
	}

	if err := runEditor(fileName); err != nil {
		return "", fmt.Errorf("failed to run editor on file %s: %v", fileName, err)
	}

	editedContentBytes, err := ioutil.ReadFile(fileName)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", fileName, err)
	}

	if err := os.Remove(fileName); err != nil {
		return "", fmt.Errorf("failed to remove temp file %s: %v", fileName, err)
	}

	editedContent := ignoreComments(string(editedContentBytes))
	return editedContent, nil
}

func ignoreComments(content string) string {
	var builder strings.Builder
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "#") {
			continue
		}

		builder.WriteString(line)
		builder.WriteString("\n")
	}

	return strings.TrimSpace(builder.String())
}
