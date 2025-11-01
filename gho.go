package main

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
)

func main() {
	url := getGitRemoteUrl()
	url, err := normalizeGitUrl(url)
	if err != nil {
		slog.Error("failed to normalize git URL", "err", err)
	}
	openUrlInBrowser(url)
}

func getGitRemoteUrl() string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		slog.Error("Failed to get git remote URL", "err", err)
		return ""
	}
	return strings.TrimSpace(string(out))
}

func normalizeGitUrl(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	if strings.HasPrefix(raw, "git@") {
		parts := strings.SplitN(raw, ":", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid SSH-style git URL: %s", raw)
		}
		host := strings.TrimPrefix(parts[0], "git@")
		path := strings.TrimSuffix(parts[1], ".git")
		return fmt.Sprintf("https://%s/%s", host, path), nil
	}

	if strings.HasPrefix(raw, "https://") {
		return strings.TrimSuffix(raw, ".git"), nil
	}

	return "", fmt.Errorf("unsupported remote URL format: %s", raw)
}

func openUrlInBrowser(url string) {
	cmd := exec.Command("open", url)
	_, err := cmd.Output()
	if err != nil {
		slog.Error("error executing open command", "err", err)
	}
}
