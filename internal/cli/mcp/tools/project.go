package tools

import "fmt"

// resolveProject translates the explicit project sentinel passed by the agent
// into the string value stored in the database.
//
//	"current" → the session's default project (auto-detected on server start)
//	"global"  → empty string (visible to every project)
//	anything else → error
func resolveProject(input, current string) (string, error) {
	switch input {
	case "current":
		if current == "" {
			return "", fmt.Errorf("project=\"current\" requested but no project is detected for this session")
		}
		return current, nil
	case "global":
		return "", nil
	case "":
		return "", fmt.Errorf("project is required; use \"current\" or \"global\"")
	default:
		return "", fmt.Errorf("project must be \"current\" or \"global\" (got %q)", input)
	}
}
