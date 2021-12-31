package edit

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"gopkg.in/yaml.v3"
)

const (
	defaultEditor = "vim"

	errMsg = "The Error that occurred on previous edit"
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

func createTempFile(content []byte) (string, error) {
	tempDir := os.TempDir()
	f, err := ioutil.TempFile(tempDir, "tagger-*.yml")
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %v", err)
	}

	if _, err := f.Write(content); err != nil {
		return "", fmt.Errorf("failed to write content to temp file %s: %v", f.Name(), err)
	}

	if err := f.Close(); err != nil {
		return "", fmt.Errorf("failed to close file %s: %v", f.Name(), err)
	}

	return f.Name(), nil
}

func Invoke(content *map[string]interface{}, result interface{}) error {
	out, err := yaml.Marshal(content)
	if err != nil {
		return fmt.Errorf("failed to marshal %#v: %v", content, err)
	}

	fileName, err := createTempFile(out)
	if err != nil {
		return err
	}

	if err := runEditor(fileName); err != nil {
		return fmt.Errorf("failed to run editor on file %s: %v", fileName, err)
	}

	changed, err := ioutil.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", fileName, err)
	}

	if err := os.Remove(fileName); err != nil {
		return fmt.Errorf("failed to remove temp file %s: %v", fileName, err)
	}

	if err := yaml.Unmarshal(changed, result); err != nil {
		// display error to user on next invoke
		(*content)["error"] = &yaml.Node{Value: err.Error(), Kind: yaml.ScalarNode, HeadComment: errMsg}

		// failed to unmarshal, retry
		return Invoke(content, result)
	}

	return nil
}
