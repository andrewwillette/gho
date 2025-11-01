package main

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/rs/zerolog/log"
)

func main() {
	url := getGitRemoteUrl()
	url, err := normalizeGitUrl(url)
	if err != nil {
		log.Fatal().AnErr("failed to normalize git URL", err)
	}
	openUrlInBrowser(url)
}

func getGitRemoteUrl() string {
	cmd := exec.Command("git", "config", "--get", "remote.origin.url")
	out, err := cmd.Output()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get git remote URL")
		return ""
	}
	return strings.TrimSpace(string(out))
}

func normalizeGitUrl(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	if strings.HasPrefix(raw, "git@") {
		// Convert git@github.com:owner/repo.git → https://github.com/owner/repo
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
	output, err := cmd.Output()
	if err != nil {
		log.Err(err)
		log.Error().Msgf("url: %s", url)
	}
	if string(output) != "" {
		log.Warn().Msgf("open url returned output: %s", string(output))
	}
}
