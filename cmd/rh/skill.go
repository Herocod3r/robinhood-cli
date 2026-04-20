package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/herocod3r/robinhood-cli/internal/config"
	"github.com/herocod3r/robinhood-cli/internal/robinhood"
	"github.com/herocod3r/robinhood-cli/internal/skill"
	"github.com/spf13/cobra"
)

func init() {
	root := &cobra.Command{
		Use:   "skill",
		Short: "Manage the robinhood-cli skill for Claude Code / Codex",
		Annotations: map[string]string{
			"category":  "meta",
			"stability": "green",
		},
	}
	install := &cobra.Command{
		Use:   "install",
		Short: "Install the skill to ~/.claude/skills or ~/.codex/skills",
		Annotations: map[string]string{
			"category":  "meta",
			"stability": "green",
		},
		RunE: runSkillInstall,
	}
	install.Flags().String("target", "claude-code", "claude-code | codex | both")
	install.Flags().Bool("force", false, "overwrite an existing skill directory")
	install.Flags().String("path", "", "override the destination directory (advanced)")
	root.AddCommand(install)
	register(root)
}

func runSkillInstall(cmd *cobra.Command, args []string) error {
	target, _ := cmd.Flags().GetString("target")
	force, _ := cmd.Flags().GetBool("force")
	override, _ := cmd.Flags().GetString("path")

	targets := []string{target}
	if target == "both" {
		if override != "" {
			return &robinhood.APIError{
				Code:    robinhood.CodeValidation,
				Message: "--target both is incompatible with --path",
				Hint:    "install one target at a time when using --path",
			}
		}
		targets = []string{"claude-code", "codex"}
	}
	for _, t := range targets {
		dest := override
		if dest == "" {
			d, err := defaultSkillDir(t)
			if err != nil {
				return err
			}
			dest = d
		}
		count, err := skill.Install(t, dest, skill.InstallOpts{Force: force})
		if err != nil {
			if strings.Contains(err.Error(), "already exists") {
				return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error(), Hint: "re-run with --force"}
			}
			return &robinhood.APIError{Code: robinhood.CodeValidation, Message: err.Error()}
		}
		fmt.Fprintf(cmd.ErrOrStderr(), "installed %d files to %s\n", count, dest)
	}
	_ = config.AppName // placeholder for future profile-scoped installs
	return nil
}

func defaultSkillDir(target string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch target {
	case "claude-code":
		return filepath.Join(home, ".claude", "skills"), nil
	case "codex":
		return filepath.Join(home, ".codex", "skills"), nil
	default:
		return "", errors.New("unknown target: " + target)
	}
}
