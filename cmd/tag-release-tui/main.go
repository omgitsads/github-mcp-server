package main

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Application states
type state int

const (
	stateInitial state = iota
	stateValidating
	stateConfirming
	stateExecuting
	stateComplete
	stateError
)

// Model represents the application state
type model struct {
	state          state
	tag            string
	currentBranch  string
	latestTag      string
	errors         []string
	validationStep int
	executionStep  int
	confirmed      bool
	executed       bool
	repoSlug       string
	testMode       bool
	width          int
	height         int
}

// Messages
type validationCompleteMsg struct {
	success bool
	errors  []string
	data    map[string]string
}

type executionStepMsg struct {
	step int
}

type executionCompleteMsg struct {
	success bool
	errors  []string
}

// Styles
var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7C3AED")).
			MarginLeft(2)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280")).
			MarginLeft(2)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#10B981")).
			Bold(true)

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F59E0B")).
			Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#3B82F6")).
			Bold(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#6B7280")).
			Padding(1, 2).
			MarginLeft(2)

	buttonStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#3B82F6")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Padding(0, 2).
			MarginRight(2)

	cancelButtonStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#6B7280")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Padding(0, 2)
)

func initialModel(tag string, testMode bool) model {
	return model{
		state:    stateValidating,
		tag:      tag,
		testMode: testMode,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.EnterAltScreen,
		performValidation(m.tag, "tag-release-charmbracelet", m.testMode),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "y", "Y":
			if m.state == stateConfirming {
				m.confirmed = true
				m.state = stateExecuting
				return m, performExecution(m.tag, m.testMode)
			}
		case "n", "N":
			if m.state == stateConfirming {
				return m, tea.Quit
			}
		case "enter":
			if m.state == stateComplete || m.state == stateError {
				return m, tea.Quit
			}
		}

	case validationCompleteMsg:
		if msg.success {
			m.state = stateConfirming
			if msg.data["currentBranch"] != "" {
				m.currentBranch = msg.data["currentBranch"]
			}
			if msg.data["latestTag"] != "" {
				m.latestTag = msg.data["latestTag"]
			}
			if msg.data["repoSlug"] != "" {
				m.repoSlug = msg.data["repoSlug"]
			}
		} else {
			m.state = stateError
			m.errors = msg.errors
		}
		return m, nil

	case executionStepMsg:
		m.executionStep = msg.step
		return m, nil

	case executionCompleteMsg:
		if msg.success {
			m.state = stateComplete
			m.executed = true
		} else {
			m.state = stateError
			m.errors = msg.errors
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	switch m.state {
	case stateInitial:
		return m.renderInitial()
	case stateValidating:
		return m.renderValidating()
	case stateConfirming:
		return m.renderConfirming()
	case stateExecuting:
		return m.renderExecuting()
	case stateComplete:
		return m.renderComplete()
	case stateError:
		return m.renderError()
	default:
		return "Unknown state"
	}
}

func (m model) renderInitial() string {
	return titleStyle.Render("🏷️  GitHub MCP Server - Tag Release") + "\n\n" +
		subtitleStyle.Render("Initializing...")
}

func (m model) renderValidating() string {
	content := titleStyle.Render("🏷️  GitHub MCP Server - Tag Release") + "\n\n" +
		subtitleStyle.Render("Validating release requirements...") + "\n\n"

	steps := []string{
		"Checking tag format",
		"Verifying current branch",
		"Fetching latest changes",
		"Checking working directory",
		"Validating branch status",
		"Checking tag availability",
	}

	for i, step := range steps {
		if i < m.validationStep {
			content += successStyle.Render("✓ ") + step + "\n"
		} else if i == m.validationStep {
			content += warningStyle.Render("⋯ ") + step + "\n"
		} else {
			content += "  " + step + "\n"
		}
	}

	return content
}

func (m model) renderConfirming() string {
	content := titleStyle.Render("🏷️  GitHub MCP Server - Tag Release")
	if m.testMode {
		content += " " + warningStyle.Render("(TEST MODE)")
	}
	content += "\n\n"

	// Summary box
	summaryContent := fmt.Sprintf("Repository: %s\n", highlightStyle.Render(m.repoSlug))
	summaryContent += fmt.Sprintf("Current branch: %s\n", highlightStyle.Render(m.currentBranch))
	if m.latestTag != "" {
		summaryContent += fmt.Sprintf("Latest release: %s\n", highlightStyle.Render(m.latestTag))
	}
	summaryContent += fmt.Sprintf("New release: %s", highlightStyle.Render(m.tag))

	content += boxStyle.Render(summaryContent) + "\n\n"

	// Steps that will be performed
	if m.testMode {
		content += subtitleStyle.Render("The following actions will be SIMULATED (test mode):") + "\n\n"
	} else {
		content += subtitleStyle.Render("The following actions will be performed:") + "\n\n"
	}

	steps := []string{
		fmt.Sprintf("Create release tag: %s", m.tag),
		fmt.Sprintf("Push tag %s to origin", m.tag),
		"Update 'latest-release' tag",
		"Push 'latest-release' tag to origin",
	}

	for _, step := range steps {
		if m.testMode {
			content += "  • [SIMULATE] " + step + "\n"
		} else {
			content += "  • " + step + "\n"
		}
	}

	if m.testMode {
		content += "\n" + successStyle.Render("✅ TEST MODE: No actual changes will be made.") + "\n\n"
	} else {
		content += "\n" + warningStyle.Render("⚠️  This will create a new release and trigger the release workflow.") + "\n\n"
	}

	// Buttons
	content += buttonStyle.Render("Yes (y)") + " " + cancelButtonStyle.Render("No (n)") + "\n\n"
	if m.testMode {
		content += subtitleStyle.Render("Do you want to proceed with the test simulation?")
	} else {
		content += subtitleStyle.Render("Do you want to proceed with the release?")
	}

	return content
}

func (m model) renderExecuting() string {
	content := titleStyle.Render("🏷️  GitHub MCP Server - Tag Release") + "\n\n" +
		subtitleStyle.Render("Creating release...") + "\n\n"

	steps := []string{
		fmt.Sprintf("Creating tag %s", m.tag),
		fmt.Sprintf("Pushing tag %s to origin", m.tag),
		"Updating 'latest-release' tag",
		"Pushing 'latest-release' tag to origin",
	}

	for i, step := range steps {
		if i < m.executionStep {
			content += successStyle.Render("✓ ") + step + "\n"
		} else if i == m.executionStep {
			content += warningStyle.Render("⋯ ") + step + "\n"
		} else {
			content += "  " + step + "\n"
		}
	}

	return content
}

func (m model) renderComplete() string {
	content := titleStyle.Render("🏷️  GitHub MCP Server - Tag Release")
	if m.testMode {
		content += " " + warningStyle.Render("(TEST MODE)")
	}
	content += "\n\n"

	if m.testMode {
		content += successStyle.Render("✅ Test simulation completed successfully!") + "\n"
		content += successStyle.Render("✅ All validation checks passed") + "\n\n"
		content += subtitleStyle.Render("🧪 Test Results:") + "\n"
		content += "  • Tag format validation: PASSED\n"
		content += "  • Branch validation: PASSED\n"
		content += "  • Working directory: PASSED\n"
		content += "  • Tag availability: PASSED\n\n"
		content += subtitleStyle.Render("To perform the actual release, run without --test flag") + "\n\n"
	} else {
		content += successStyle.Render("✅ Successfully tagged and pushed release "+m.tag) + "\n"
		content += successStyle.Render("✅ 'latest-release' tag has been updated") + "\n\n"

		// Post-release instructions
		content += subtitleStyle.Render("🎉 Release "+m.tag+" has been initiated!") + "\n\n"
		content += subtitleStyle.Render("Next steps:") + "\n"
		steps := []string{
			fmt.Sprintf("📋 Check https://github.com/%s/releases and wait for the draft release", m.repoSlug),
			"✏️  Edit the new release, delete existing notes and click auto-generate button",
			"✨ Add a section at the top calling out the main features",
			"🚀 Publish the release",
			"📢 Post message in #gh-mcp-releases channel in Slack",
		}

		for _, step := range steps {
			content += "  " + step + "\n"
		}

		content += "\n" + subtitleStyle.Render(fmt.Sprintf("📦 Draft Release: https://github.com/%s/releases/tag/%s", m.repoSlug, m.tag)) + "\n\n"
	}

	content += subtitleStyle.Render("Press Enter to exit")
	return content
}

func (m model) renderError() string {
	content := titleStyle.Render("🏷️  GitHub MCP Server - Tag Release") + "\n\n"
	content += errorStyle.Render("❌ Release creation failed") + "\n\n"

	for _, err := range m.errors {
		content += errorStyle.Render("• "+err) + "\n"
	}

	content += "\n" + subtitleStyle.Render("Press Enter to exit")
	return content
}

// Command functions
func performValidation(tag, allowedBranch string, testMode bool) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		errors := []string{}
		data := make(map[string]string)

		// 1. Validate tag format
		tagRegex := regexp.MustCompile(`^v[0-9]+\.[0-9]+\.[0-9]+(-.*)?$`)
		if !tagRegex.MatchString(tag) {
			errors = append(errors, "Tag must be in format vX.Y.Z or vX.Y.Z-suffix (e.g., v1.0.0 or v1.0.0-rc1)")
		}

		// 2. Check current branch
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		output, err := cmd.Output()
		if err != nil {
			errors = append(errors, "Failed to get current branch")
		} else {
			currentBranch := strings.TrimSpace(string(output))
			data["currentBranch"] = currentBranch
			if currentBranch != allowedBranch {
				if testMode {
					// In test mode, just warn but don't fail
					errors = append(errors, fmt.Sprintf("WARNING: Not on '%s' branch (current: '%s'), but continuing in test mode", allowedBranch, currentBranch))
				} else {
					errors = append(errors, fmt.Sprintf("You must be on the '%s' branch to create a release. Current branch is '%s'", allowedBranch, currentBranch))
				}
			}
		}

		// 3. Fetch latest from origin
		cmd = exec.Command("git", "fetch", "origin", allowedBranch)
		if err := cmd.Run(); err != nil {
			if testMode {
				errors = append(errors, fmt.Sprintf("WARNING: Failed to fetch latest changes from origin/%s, but continuing in test mode", allowedBranch))
			} else {
				errors = append(errors, fmt.Sprintf("Failed to fetch latest changes from origin/%s", allowedBranch))
			}
		}

		// 4. Check if working directory is clean
		cmd = exec.Command("git", "diff-index", "--quiet", "HEAD", "--")
		if err := cmd.Run(); err != nil {
			errors = append(errors, "Working directory is not clean. Please commit or stash your changes")
		}

		// 5. Check if main is up-to-date with origin/main
		cmd = exec.Command("git", "rev-parse", "@")
		localSha, err := cmd.Output()
		if err != nil {
			errors = append(errors, "Failed to get local SHA")
		}

		cmd = exec.Command("git", "rev-parse", "@{u}")
		remoteSha, err := cmd.Output()
		if err != nil {
			errors = append(errors, "Failed to get remote SHA")
		}

		if string(localSha) != string(remoteSha) {
			if testMode {
				errors = append(errors, fmt.Sprintf("WARNING: Local '%s' branch is not up-to-date with 'origin/%s', but continuing in test mode", allowedBranch, allowedBranch))
			} else {
				errors = append(errors, fmt.Sprintf("Your local '%s' branch is not up-to-date with 'origin/%s'. Please pull the latest changes", allowedBranch, allowedBranch))
			}
		}

		// 6. Check if tag already exists
		cmd = exec.Command("git", "tag", "-l")
		output, err = cmd.Output()
		if err != nil {
			errors = append(errors, "Failed to list local tags")
		} else {
			tags := strings.Split(strings.TrimSpace(string(output)), "\n")
			for _, existingTag := range tags {
				if existingTag == tag {
					errors = append(errors, fmt.Sprintf("Tag %s already exists locally", tag))
					break
				}
			}
		}

		cmd = exec.Command("git", "ls-remote", "--tags", "origin")
		output, err = cmd.Output()
		if err != nil {
			errors = append(errors, "Failed to check remote tags")
		} else {
			if strings.Contains(string(output), "refs/tags/"+tag) {
				errors = append(errors, fmt.Sprintf("Tag %s already exists on remote 'origin'", tag))
			}
		}

		// Get latest tag
		cmd = exec.Command("git", "tag", "--sort=-version:refname")
		output, err = cmd.Output()
		if err == nil {
			tags := strings.Split(strings.TrimSpace(string(output)), "\n")
			if len(tags) > 0 && tags[0] != "" {
				data["latestTag"] = tags[0]
			}
		}

		// Get repository slug
		cmd = exec.Command("git", "remote", "get-url", "origin")
		output, err = cmd.Output()
		if err == nil {
			repoUrl := strings.TrimSpace(string(output))
			// Extract slug from URL
			slug := repoUrl
			slug = strings.TrimSuffix(slug, ".git")
			if strings.Contains(slug, "github.com/") {
				parts := strings.Split(slug, "github.com/")
				if len(parts) > 1 {
					slug = parts[1]
				}
			}
			slug = strings.TrimPrefix(slug, "git@github.com:")
			data["repoSlug"] = slug
		}

		return validationCompleteMsg{
			success: len(errors) == 0,
			errors:  errors,
			data:    data,
		}
	})
}

func performExecution(tag string, testMode bool) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		errors := []string{}

		if testMode {
			// In test mode, simulate the steps without actually executing them
			return executionCompleteMsg{success: true, errors: nil}
		}

		// Step 0: Create the tag
		cmd := exec.Command("git", "tag", "-a", tag, "-m", "Release "+tag)
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to create tag %s: %v", tag, err))
			return executionCompleteMsg{success: false, errors: errors}
		}

		// Step 1: Push the tag
		cmd = exec.Command("git", "push", "origin", tag)
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to push tag %s: %v", tag, err))
			return executionCompleteMsg{success: false, errors: errors}
		}

		// Step 2: Update latest-release tag
		cmd = exec.Command("git", "tag", "-f", "latest-release", tag)
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to update latest-release tag: %v", err))
			return executionCompleteMsg{success: false, errors: errors}
		}

		// Step 3: Push latest-release tag
		cmd = exec.Command("git", "push", "origin", "latest-release", "--force")
		if err := cmd.Run(); err != nil {
			errors = append(errors, fmt.Sprintf("Failed to push latest-release tag: %v", err))
			return executionCompleteMsg{success: false, errors: errors}
		}

		return executionCompleteMsg{success: true, errors: nil}
	})
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Error: No tag specified.")
		fmt.Println("Usage: tag-release-tui vX.Y.Z [--test]")
		fmt.Println("  --test: Run in test mode (validation only, no actual changes)")
		os.Exit(1)
	}

	tag := os.Args[1]
	testMode := false

	// Check for test flag
	for i := 2; i < len(os.Args); i++ {
		if os.Args[i] == "--test" || os.Args[i] == "-t" {
			testMode = true
			break
		}
	}

	if testMode {
		fmt.Println("🧪 Running in TEST MODE - no actual changes will be made")
	}

	p := tea.NewProgram(
		initialModel(tag, testMode),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
