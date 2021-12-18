package edit

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	defaultEditor = "vim"

	BasicInstructions = `# Quit the file without saving to not modify the results
# Comments are lines starting with a '#' and will be ignored.`

	TemplateInstructions = `# This will overwrite all your manual changes!
#
# Top Level Variables:
# {{ .RepositoryName }}                  hello-world	
# {{ .RepositoryOwner }}                 octocat
# {{ .RepositoryUrl }}                   https://github.com/octocat/hello-world	
# {{ .RepositoryDescription }}           Example description
# {{ .RepositoryDefaultBranch }}         main
# {{ .Commits }}                         Array of commits
#
# Commit:
# {{ .Sha }}                             Unique identifier for commit
# {{ .Url }}                             URL to commit
# {{ .Message }}                         Full commit message (includes newlines)
#
# Author/Committer:
# {{ .AuthorUsername }}                  octocat (GitHub Username)
# {{ .AuthorName }}                      octocat (Commit Name)
# {{ .AuthorEmail }}                     octocat@github.com
# {{ .AuthorDate }} 
# {{ .AuthorUrl }}                       https://github.com/octocat
#
# Templates also include Sprig functions: https://masterminds.github.io/sprig/strings.html
#
# Example:
#
# {{ range .Commits }}
# {{ substr 0 8 .Sha }} committed by {{ .CommitterUsername }} and authored by {{ .AuthorUsername }} {{ .Message }}
# {{ end }}
#
` + BasicInstructions
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
	lines := strings.Split(content, "\n")
	var builder strings.Builder

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}

		builder.WriteString(line)
		builder.WriteRune('\n')
	}

	return strings.TrimSpace(builder.String())
}
