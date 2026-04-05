package commands

import (
	"os/exec"
	"regexp"
	"strings"
)

var slugRegexp = regexp.MustCompile(`[^a-z0-9-]+`)

func detectProject(dir string) string {
	if hasGit() {
		if name := fromGitRemote(dir); name != "" {
			return slugify(name)
		}
	}
	return slugify(dir)
}

func hasGit() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

func slugify(name string) string {
	name = strings.ToLower(name)
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = slugRegexp.ReplaceAllString(name, "")
	name = strings.Trim(name, "-")
	return name
}

func fromGitRemote(dir string) string {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return extractRepoName(strings.TrimSpace(string(out)))
}

func extractRepoName(url string) string {
	url = strings.TrimSuffix(url, ".git")

	// SSH: git@github.com:user/repo
	if i := strings.LastIndex(url, ":"); i != -1 && !strings.Contains(url, "://") {
		parts := strings.Split(url[i+1:], "/")
		return parts[len(parts)-1]
	}

	// HTTPS: https://github.com/user/repo
	parts := strings.Split(url, "/")
	return parts[len(parts)-1]
}
