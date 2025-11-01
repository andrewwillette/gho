package main

import (
	"flag"
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/andrewwillette/gocommon"
	"github.com/rs/zerolog/log"
)

type gitUrlRepresentation int

const (
	ssh gitUrlRepresentation = iota
	https
)

func main() {
	flag.Parse()
	gocommon.ConfigureConsoleZerolog()
	url := getUrlFromGitRemote()
	openUrlInBrowser(url)
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

func getUrlFromGitRemote() string {
	cmd := "git remote -v | grep push"
	out, err := exec.Command("bash", "-c", cmd).Output()
	log.Debug().Msgf("git remote -v output: %s", out)
	if err != nil {
		log.Error().Msg("Error executing 'git remote -v' command")
		log.Err(err)
	}
	url := parseUrl(string(out))
	return url
}

type githubOrigin struct {
	domain    string
	repoName  string
	repoOwner string
}

func (gh *githubOrigin) getUrl() string {
	log.Debug().Msg("githubOrigin#getUrl")
	log.Debug().Msgf("%+v", gh)
	return fmt.Sprintf("https://%s/%s/%s", gh.domain, gh.repoOwner, gh.repoName)
}

func (gur gitUrlRepresentation) String() string {
	switch gur {
	case ssh:
		return "ssh"
	case https:
		return "https"
	default:
		return "unsupported git url rep"
	}
}

func getGitUrlRepr(gitRemoteOutput string) gitUrlRepresentation {
	match, _ := regexp.MatchString("https://", gitRemoteOutput)
	if match {
		log.Debug().Msg("git remote is of type https.")
		return https
	} else {
		log.Debug().Msg("git remote is of type ssh.")
		return ssh
	}
}

// parseUrl get url from 'git remote -v' output
func parseUrl(gitRemoteOutput string) string {
	gh := githubOrigin{}
	gitUrlRepr := getGitUrlRepr(gitRemoteOutput)
	switch gitUrlRepr {
	case ssh:
		gh.domain = parseGithubDomainSsh(gitRemoteOutput)
		gh.repoName = parseGithubRepoNameSsh(gitRemoteOutput)
		gh.repoOwner = parseGithubRepoOwnerSsh(gitRemoteOutput)
	case https:
		gh.domain = parseGithubDomainHttps(gitRemoteOutput)
		gh.repoName = parseGithubRepoNameHttps(gitRemoteOutput)
		gh.repoOwner = parseGithubRepoOwnerHttps(gitRemoteOutput)
	default:
		return "invalid gitUrlRepr"
	}
	if gh.domain == "" || gh.repoName == "" || gh.repoOwner == "" {
		log.Error().Msgf("Failed to parse one or more parts of the GitHub origin, domain: %s, repoName: %s, repoOwner: %s", gh.domain, gh.repoName, gh.repoOwner)
		return ""
	}
	return gh.getUrl()
}

func parseGithubDomainSsh(gitSshUrl string) string {
	log.Debug().Msg("parseGithubDomainSsh")
	r := regexp.MustCompile(`.*push`)
	result1 := r.FindString(gitSshUrl)
	log.Debug().Msgf("result1: %s", result1)
	r = regexp.MustCompile(`@([^:].)*`)
	result2 := r.FindString(result1)
	log.Debug().Msgf("result2: %s", result2)
	r = regexp.MustCompile(`[^@].*`)
	result3 := r.FindString(result2)
	log.Debug().Msgf("result3: %s", result3)
	return result3
}

func parseGithubRepoOwnerSsh(gitSshUrl string) string {
	log.Debug().Msg("parseGithubRepoOwnerSsh")
	r := regexp.MustCompile(`:.*/`)
	result1 := r.FindString(gitSshUrl)
	log.Debug().Msgf("result1: %s", result1)
	r = regexp.MustCompile(`[^:][\w|\d|-|\.]*`)
	result2 := r.FindString(result1)
	log.Debug().Msgf("result2: %s", result2)
	return result2
}

func parseGithubRepoNameSsh(gitSshUrl string) string {
	log.Debug().Msgf("parseGithubRepoNameSsh %s", gitSshUrl)
	r := regexp.MustCompile(`/.*\.git`)
	result1 := r.FindString(gitSshUrl)
	log.Debug().Msgf("result1: %s", result1)
	r = regexp.MustCompile(`[^/](\w|\d|-|\.)*`)
	result2 := r.FindString(result1)
	log.Debug().Msgf("result2: %s", result2)
	result3 := strings.TrimSuffix(result2, ".git")
	log.Debug().Msgf("result3: %s", result3)
	return result3
}

func parseGithubDomainHttps(gitHttpsUrl string) string {
	r := regexp.MustCompile(`https://([^/]+)`)
	match := r.FindStringSubmatch(gitHttpsUrl)
	if len(match) > 1 {
		return match[1]
	}
	log.Warn().Msg("Could not parse domain from HTTPS URL")
	return ""
}

func parseGithubRepoOwnerHttps(gitHttpsUrl string) string {
	r := regexp.MustCompile(`https://[^/]+/([^/]+)/`)
	match := r.FindStringSubmatch(gitHttpsUrl)
	if len(match) > 1 {
		return match[1]
	}
	log.Warn().Msg("Could not parse repo owner from HTTPS URL")
	return ""
}

func parseGithubRepoNameHttps(gitHttpsUrl string) string {
	r := regexp.MustCompile(`https://[^/]+/[^/]+/([^ ]+?)(?:\.git)?\s`)
	match := r.FindStringSubmatch(gitHttpsUrl)
	if len(match) > 1 {
		return strings.TrimSuffix(match[1], ".git")
	}
	log.Warn().Msg("Could not parse repo name from HTTPS URL")
	return ""
}
