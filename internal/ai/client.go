package ai

import (
	"os/exec"
)

func Generate(prompt string) (string, error) {

	cmd := exec.Command(
		"llama-cli",
		"-m", "models/phi-3-mini.gguf",
		"-p", prompt,
		"-n", "300",
		"-ngl", "0",
	)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}