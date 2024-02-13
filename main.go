package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

func main() {
	todoCmd := &cobra.Command{
		Use:   "todo",
		Short: "List pending pull requests",
		Long:  "List pending pull requests",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Opening pending PRs in browser...")

			params := []string{
				"is:open",
				"is:pr",
				"archived:false",
				"user:docker",
				"review-requested:Manzanit0",
				"draft:false",
			}

			fmt.Println("🖥  opening pending work in browser...")
			query := strings.Join(params, "+")
			_ = exec.Command("open", fmt.Sprintf("https://github.com/pulls?q=%s", query)).Run()
		},
	}

	prCmd := &cobra.Command{
		Use:   "pr",
		Short: "Create PR for current HEAD",
		Long:  "Create PR for current HEAD",
		Args:  cobra.MinimumNArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("🪄 pushing to remote...")
			b, err := exec.Command("git", "push", "-u", "origin", "HEAD").CombinedOutput()
			if err != nil {
				fmt.Println("💥 failed to push HEAD to remote:", string(b))
				return
			}

			fmt.Println(" compiling commit messages...")
			b, err = exec.Command("git", "log", "origin..HEAD", `--format="## %s%n%b"`).CombinedOutput()
			if err != nil {
				fmt.Println("💥 failed to compile commit messages:", string(b))
				return
			}

			body := fmt.Sprintf("# Description\n\n%s", string(b))

			// TODO: this is a quick-fix for getting the format right; I haven't
			// bothered understanding why the `git log` puts quotes.
			body = strings.ReplaceAll(body, `"`, "")

			fmt.Println("🗞  creating pull request...")
			command := []string{"pr", "create", "--fill", "--assignee", "Manzanit0", "--body", body}
			if isDraft := cmd.Flag("draft").Value.String(); isDraft == "true" {
				command = append(command, "--draft")
			}

			if title := cmd.Flag("title").Value.String(); title != "" {
				command = append(command, "--title", title)
			}

			b, err = exec.Command("gh", command...).CombinedOutput()
			if err != nil {
				fmt.Println("💥 failed to create pull request:", string(b))
				return
			}

			if cmd.Flag("open").Value.String() == "false" {
				return
			}

			fmt.Println("🖥  opening pull request in browser...")
			b, err = exec.Command("gh", "pr", "view", "--web").CombinedOutput()
			if err != nil {
				fmt.Println("💥 failed to open pull request in browser:", string(b))
				return
			}

			fmt.Println(string(b))
		},
	}

	prCmd.PersistentFlags().Bool("draft", false, "The PR will be created as draft")
	prCmd.PersistentFlags().Bool("open", false, "The PR will be opened in the browser")
	prCmd.PersistentFlags().String("title", "", "Title to give to the PR")

	rootCmd := &cobra.Command{Use: "wk"}
	rootCmd.AddCommand(prCmd, todoCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("command failed: %s", err.Error())
	}
}
