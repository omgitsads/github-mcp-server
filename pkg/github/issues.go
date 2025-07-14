package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/github/github-mcp-server/pkg/utils"
	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-github/v72/github"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shurcooL/githubv4"
)

// GetIssue creates a tool to get details of a specific issue in a GitHub repository.
func GetIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_issue",
			Description: t("TOOL_GET_ISSUE_DESCRIPTION", "Get details of a specific issue in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_ISSUE_USER_TITLE", "Get issue details"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "issue_number"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_GET_ISSUE_OWNER_DESC", "The owner of the repository"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_GET_ISSUE_REPO_DESC", "The name of the repository"),
					},
					"issue_number": {
						Type:        "number",
						Description: t("TOOL_GET_ISSUE_NUMBER_DESC", "The number of the issue"),
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			issue, resp, err := client.Issues.Get(ctx, owner, repo, issueNumber)
			if err != nil {
				return nil, fmt.Errorf("failed to get issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get issue: %s", string(body))), nil
			}

			r, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal issue: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// AddIssueComment creates a tool to add a comment to an issue.
func AddIssueComment(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "add_issue_comment",
			Description: t("TOOL_ADD_ISSUE_COMMENT_DESCRIPTION", "Add a comment to a specific issue in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_ADD_ISSUE_COMMENT_USER_TITLE", "Add comment to issue"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "issue_number", "body"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_ADD_ISSUE_COMMENT_OWNER_DESC", "Repository owner"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_ADD_ISSUE_COMMENT_REPO_DESC", "Repository name"),
					},
					"issue_number": {
						Type:        "number",
						Description: t("TOOL_ADD_ISSUE_COMMENT_NUMBER_DESC", "Issue number to comment on"),
					},
					"body": {
						Type:        "string",
						Description: t("TOOL_ADD_ISSUE_COMMENT_BODY_DESC", "Comment content"),
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			body, err := RequiredParam[string](request, "body")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			comment := &github.IssueComment{
				Body: github.Ptr(body),
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			createdComment, resp, err := client.Issues.CreateComment(ctx, owner, repo, issueNumber, comment)
			if err != nil {
				return nil, fmt.Errorf("failed to create comment: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to create comment: %s", string(body))), nil
			}

			r, err := json.Marshal(createdComment)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// SearchIssues creates a tool to search for issues.
func SearchIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "search_issues",
			Description: t("TOOL_SEARCH_ISSUES_DESCRIPTION", "Search for issues in GitHub repositories using issues search syntax already scoped to is:issue"),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_SEARCH_ISSUES_USER_TITLE", "Search issues"),
				ReadOnlyHint: true,
			},
			InputSchema: WithPagination(&jsonschema.Schema{
				Type:     "object",
				Required: []string{"query"},
				Properties: map[string]*jsonschema.Schema{
					"query": {
						Type:        "string",
						Description: t("TOOL_SEARCH_ISSUES_QUERY_DESC", "Search query using GitHub issues search syntax"),
					},
					"owner": {
						Type:        "string",
						Description: t("TOOL_SEARCH_ISSUES_OWNER_DESC", "Optional repository owner. If provided with repo, only notifications for this repository are listed."),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_SEARCH_ISSUES_REPO_DESC", "Optional repository name. If provided with owner, only notifications for this repository are listed."),
					},
					"sort": {
						Type:        "string",
						Description: t("TOOL_SEARCH_ISSUES_SORT_DESC", "Sort field by number of matches of categories, defaults to best match"),
						Enum: []any{
							"comments",
							"reactions",
							"reactions-+1",
							"reactions--1",
							"reactions-smile",
							"reactions-thinking_face",
							"reactions-heart",
							"reactions-tada",
							"interactions",
							"created",
							"updated",
						},
					},
					"order": {
						Type:        "string",
						Description: t("TOOL_SEARCH_ISSUES_ORDER_DESC", "Sort order"),
						Enum:        []any{"asc", "desc"},
					},
				},
			}),
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			return searchHandler(ctx, getClient, request, "issue", "failed to search issues")
		}
}

// CreateIssue creates a tool to create a new issue in a GitHub repository.
func CreateIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "create_issue",
			Description: t("TOOL_CREATE_ISSUE_DESCRIPTION", "Create a new issue in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_CREATE_ISSUE_USER_TITLE", "Open new issue"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "title"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_CREATE_ISSUE_OWNER_DESC", "Repository owner"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_CREATE_ISSUE_REPO_DESC", "Repository name"),
					},
					"title": {
						Type:        "string",
						Description: t("TOOL_CREATE_ISSUE_TITLE_DESC", "Issue title"),
					},
					"body": {
						Type:        "string",
						Description: t("TOOL_CREATE_ISSUE_BODY_DESC", "Issue body content"),
					},
					"assignees": {
						Type:        "array",
						Description: t("TOOL_CREATE_ISSUE_ASSIGNEES_DESC", "Usernames to assign to this issue"),
						Items: &jsonschema.Schema{
							Type: "string",
						},
					},
					"labels": {
						Type:        "array",
						Description: t("TOOL_CREATE_ISSUE_LABELS_DESC", "Labels to apply to this issue"),
						Items: &jsonschema.Schema{
							Type: "string",
						},
					},
					"milestone": {
						Type:        "number",
						Description: t("TOOL_CREATE_ISSUE_MILESTONE_DESC", "Milestone number"),
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			title, err := RequiredParam[string](request, "title")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Optional parameters
			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Get assignees
			assignees, err := OptionalStringArrayParam(request, "assignees")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Get labels
			labels, err := OptionalStringArrayParam(request, "labels")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Get optional milestone
			milestone, err := OptionalIntParam(request, "milestone")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			var milestoneNum *int
			if milestone != 0 {
				milestoneNum = &milestone
			}

			// Create the issue request
			issueRequest := &github.IssueRequest{
				Title:     github.Ptr(title),
				Body:      github.Ptr(body),
				Assignees: &assignees,
				Labels:    &labels,
				Milestone: milestoneNum,
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			issue, resp, err := client.Issues.Create(ctx, owner, repo, issueRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to create issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to create issue: %s", string(body))), nil
			}

			r, err := json.Marshal(issue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// ListIssues creates a tool to list and filter repository issues
func ListIssues(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "list_issues",
			Description: t("TOOL_LIST_ISSUES_DESCRIPTION", "List issues in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_LIST_ISSUES_USER_TITLE", "List issues"),
				ReadOnlyHint: true,
			},
			InputSchema: WithPagination(&jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_LIST_ISSUES_OWNER_DESC", "Repository owner"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_LIST_ISSUES_REPO_DESC", "Repository name"),
					},
					"state": {
						Type:        "string",
						Description: t("TOOL_LIST_ISSUES_STATE_DESC", "Filter by state"),
						Enum:        []any{"open", "closed", "all"},
					},
					"labels": {
						Type:        "array",
						Description: t("TOOL_LIST_ISSUES_LABELS_DESC", "Filter by labels"),
						Items: &jsonschema.Schema{
							Type: "string",
						},
					},
					"sort": {
						Type:        "string",
						Description: t("TOOL_LIST_ISSUES_SORT_DESC", "Sort order"),
						Enum:        []any{"created", "updated", "comments"},
					},
					"direction": {
						Type:        "string",
						Description: t("TOOL_LIST_ISSUES_DIRECTION_DESC", "Sort direction"),
						Enum:        []any{"asc", "desc"},
					},
					"since": {
						Type:        "string",
						Description: t("TOOL_LIST_ISSUES_SINCE_DESC", "Filter by date (ISO 8601 timestamp)"),
					},
				},
			}),
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts := &github.IssueListByRepoOptions{}

			// Set optional parameters if provided
			opts.State, err = OptionalParam[string](request, "state")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Get labels
			opts.Labels, err = OptionalStringArrayParam(request, "labels")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts.Sort, err = OptionalParam[string](request, "sort")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts.Direction, err = OptionalParam[string](request, "direction")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			since, err := OptionalParam[string](request, "since")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if since != "" {
				timestamp, err := parseISOTimestamp(since)
				if err != nil {
					return utils.NewToolResultError(fmt.Sprintf("failed to list issues: %s", err.Error())), nil
				}
				opts.Since = timestamp
			}

			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			opts.ListOptions.Page = pagination.page
			opts.ListOptions.PerPage = pagination.perPage

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			issues, resp, err := client.Issues.ListByRepo(ctx, owner, repo, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to list issues: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to list issues: %s", string(body))), nil
			}

			r, err := json.Marshal(issues)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal issues: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// UpdateIssue creates a tool to update an existing issue in a GitHub repository.
func UpdateIssue(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "update_issue",
			Description: t("TOOL_UPDATE_ISSUE_DESCRIPTION", "Update an existing issue in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_UPDATE_ISSUE_USER_TITLE", "Edit issue"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "issue_number"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_UPDATE_ISSUE_OWNER_DESC", "Repository owner"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_UPDATE_ISSUE_REPO_DESC", "Repository name"),
					},
					"issue_number": {
						Type:        "number",
						Description: t("TOOL_UPDATE_ISSUE_NUMBER_DESC", "Issue number to update"),
					},
					"title": {
						Type:        "string",
						Description: t("TOOL_UPDATE_ISSUE_TITLE_DESC", "New title"),
					},
					"body": {
						Type:        "string",
						Description: t("TOOL_UPDATE_ISSUE_BODY_DESC", "New description"),
					},
					"state": {
						Type:        "string",
						Description: t("TOOL_UPDATE_ISSUE_STATE_DESC", "New state"),
						Enum:        []any{"open", "closed"},
					},
					"labels": {
						Type:        "array",
						Description: t("TOOL_UPDATE_ISSUE_LABELS_DESC", "New labels"),
						Items: &jsonschema.Schema{
							Type: "string",
						},
					},
					"assignees": {
						Type:        "array",
						Description: t("TOOL_UPDATE_ISSUE_ASSIGNEES_DESC", "New assignees"),
						Items: &jsonschema.Schema{
							Type: "string",
						},
					},
					"milestone": {
						Type:        "number",
						Description: t("TOOL_UPDATE_ISSUE_MILESTONE_DESC", "New milestone number"),
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Create the issue request with only provided fields
			issueRequest := &github.IssueRequest{}

			// Set optional parameters if provided
			title, err := OptionalParam[string](request, "title")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if title != "" {
				issueRequest.Title = github.Ptr(title)
			}

			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if body != "" {
				issueRequest.Body = github.Ptr(body)
			}

			state, err := OptionalParam[string](request, "state")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if state != "" {
				issueRequest.State = github.Ptr(state)
			}

			// Get labels
			labels, err := OptionalStringArrayParam(request, "labels")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if len(labels) > 0 {
				issueRequest.Labels = &labels
			}

			// Get assignees
			assignees, err := OptionalStringArrayParam(request, "assignees")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if len(assignees) > 0 {
				issueRequest.Assignees = &assignees
			}

			milestone, err := OptionalIntParam(request, "milestone")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			if milestone != 0 {
				milestoneNum := milestone
				issueRequest.Milestone = &milestoneNum
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			updatedIssue, resp, err := client.Issues.Edit(ctx, owner, repo, issueNumber, issueRequest)
			if err != nil {
				return nil, fmt.Errorf("failed to update issue: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to update issue: %s", string(body))), nil
			}

			r, err := json.Marshal(updatedIssue)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// GetIssueComments creates a tool to get comments for a GitHub issue.
func GetIssueComments(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_issue_comments",
			Description: t("TOOL_GET_ISSUE_COMMENTS_DESCRIPTION", "Get comments for a specific issue in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_ISSUE_COMMENTS_USER_TITLE", "Get issue comments"),
				ReadOnlyHint: true,
			},
			InputSchema: WithPagination(&jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "issue_number"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_GET_ISSUE_COMMENTS_OWNER_DESC", "Repository owner"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_GET_ISSUE_COMMENTS_REPO_DESC", "Repository name"),
					},
					"issue_number": {
						Type:        "number",
						Description: t("TOOL_GET_ISSUE_COMMENTS_NUMBER_DESC", "Issue number"),
					},
				},
			}),
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			owner, err := RequiredParam[string](request, "owner")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			repo, err := RequiredParam[string](request, "repo")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			issueNumber, err := RequiredInt(request, "issue_number")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts := &github.IssueListCommentsOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			comments, resp, err := client.Issues.ListComments(ctx, owner, repo, issueNumber, opts)
			if err != nil {
				return nil, fmt.Errorf("failed to get issue comments: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get issue comments: %s", string(body))), nil
			}

			r, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// mvpDescription is an MVP idea for generating tool descriptions from structured data in a shared format.
// It is not intended for widespread usage and is not a complete implementation.
type mvpDescription struct {
	summary        string
	outcomes       []string
	referenceLinks []string
}

func (d *mvpDescription) String() string {
	var sb strings.Builder
	sb.WriteString(d.summary)
	if len(d.outcomes) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString("This tool can help with the following outcomes:\n")
		for _, outcome := range d.outcomes {
			sb.WriteString(fmt.Sprintf("- %s\n", outcome))
		}
	}

	if len(d.referenceLinks) > 0 {
		sb.WriteString("\n\n")
		sb.WriteString("More information can be found at:\n")
		for _, link := range d.referenceLinks {
			sb.WriteString(fmt.Sprintf("- %s\n", link))
		}
	}

	return sb.String()
}

func AssignCopilotToIssue(getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	description := mvpDescription{
		summary: "Assign Copilot to a specific issue in a GitHub repository.",
		outcomes: []string{
			"a Pull Request created with source code changes to resolve the issue",
		},
		referenceLinks: []string{
			"https://docs.github.com/en/copilot/using-github-copilot/using-copilot-coding-agent-to-work-on-tasks/about-assigning-tasks-to-copilot",
		},
	}

	return &mcp.Tool{
			Name:        "assign_copilot_to_issue",
			Description: t("TOOL_ASSIGN_COPILOT_TO_ISSUE_DESCRIPTION", description.String()),
			Annotations: &mcp.ToolAnnotations{
				Title:          t("TOOL_ASSIGN_COPILOT_TO_ISSUE_USER_TITLE", "Assign Copilot to issue"),
				ReadOnlyHint:   false,
				IdempotentHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "issueNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_ASSIGN_COPILOT_TO_ISSUE_OWNER_DESC", "Repository owner"),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_ASSIGN_COPILOT_TO_ISSUE_REPO_DESC", "Repository name"),
					},
					"issueNumber": {
						Type:        "number",
						Description: t("TOOL_ASSIGN_COPILOT_TO_ISSUE_NUMBER_DESC", "Issue number"),
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner       string
				Repo        string
				IssueNumber int32
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getGQLClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			// Firstly, we try to find the copilot bot in the suggested actors for the repository.
			// Although as I write this, we would expect copilot to be at the top of the list, in future, maybe
			// it will not be on the first page of responses, thus we will keep paginating until we find it.
			type botAssignee struct {
				ID       githubv4.ID
				Login    string
				TypeName string `graphql:"__typename"`
			}

			type suggestedActorsQuery struct {
				Repository struct {
					SuggestedActors struct {
						Nodes []struct {
							Bot botAssignee `graphql:"... on Bot"`
						}
						PageInfo struct {
							HasNextPage bool
							EndCursor   string
						}
					} `graphql:"suggestedActors(first: 100, after: $endCursor, capabilities: CAN_BE_ASSIGNED)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			}

			variables := map[string]any{
				"owner":     githubv4.String(params.Owner),
				"name":      githubv4.String(params.Repo),
				"endCursor": (*githubv4.String)(nil),
			}

			var copilotAssignee *botAssignee
			for {
				var query suggestedActorsQuery
				err := client.Query(ctx, &query, variables)
				if err != nil {
					return nil, err
				}

				// Iterate all the returned nodes looking for the copilot bot, which is supposed to have the
				// same name on each host. We need this in order to get the ID for later assignment.
				for _, node := range query.Repository.SuggestedActors.Nodes {
					if node.Bot.Login == "copilot-swe-agent" {
						copilotAssignee = &node.Bot
						break
					}
				}

				if !query.Repository.SuggestedActors.PageInfo.HasNextPage {
					break
				}
				variables["endCursor"] = githubv4.String(query.Repository.SuggestedActors.PageInfo.EndCursor)
			}

			// If we didn't find the copilot bot, we can't proceed any further.
			if copilotAssignee == nil {
				// The e2e tests depend upon this specific message to skip the test.
				return utils.NewToolResultError("copilot isn't available as an assignee for this issue. Please inform the user to visit https://docs.github.com/en/copilot/using-github-copilot/using-copilot-coding-agent-to-work-on-tasks/about-assigning-tasks-to-copilot for more information."), nil
			}

			// Next let's get the GQL Node ID and current assignees for this issue because the only way to
			// assign copilot is to use replaceActorsForAssignable which requires the full list.
			var getIssueQuery struct {
				Repository struct {
					Issue struct {
						ID        githubv4.ID
						Assignees struct {
							Nodes []struct {
								ID githubv4.ID
							}
						} `graphql:"assignees(first: 100)"`
					} `graphql:"issue(number: $number)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			}

			variables = map[string]any{
				"owner":  githubv4.String(params.Owner),
				"name":   githubv4.String(params.Repo),
				"number": githubv4.Int(params.IssueNumber),
			}

			if err := client.Query(ctx, &getIssueQuery, variables); err != nil {
				return utils.NewToolResultError(fmt.Sprintf("failed to get issue ID: %v", err)), nil
			}

			// Finally, do the assignment. Just for reference, assigning copilot to an issue that it is already
			// assigned to seems to have no impact (which is a good thing).
			var assignCopilotMutation struct {
				ReplaceActorsForAssignable struct {
					Typename string `graphql:"__typename"` // Not required but we need a selector or GQL errors
				} `graphql:"replaceActorsForAssignable(input: $input)"`
			}

			actorIDs := make([]githubv4.ID, len(getIssueQuery.Repository.Issue.Assignees.Nodes)+1)
			for i, node := range getIssueQuery.Repository.Issue.Assignees.Nodes {
				actorIDs[i] = node.ID
			}
			actorIDs[len(getIssueQuery.Repository.Issue.Assignees.Nodes)] = copilotAssignee.ID

			if err := client.Mutate(
				ctx,
				&assignCopilotMutation,
				ReplaceActorsForAssignableInput{
					AssignableID: getIssueQuery.Repository.Issue.ID,
					ActorIDs:     actorIDs,
				},
				nil,
			); err != nil {
				return nil, fmt.Errorf("failed to replace actors for assignable: %w", err)
			}

			return utils.NewToolResultText("successfully assigned copilot to issue"), nil
		}
}

type ReplaceActorsForAssignableInput struct {
	AssignableID githubv4.ID   `json:"assignableId"`
	ActorIDs     []githubv4.ID `json:"actorIds"`
}

// parseISOTimestamp parses an ISO 8601 timestamp string into a time.Time object.
// Returns the parsed time or an error if parsing fails.
// Example formats supported: "2023-01-15T14:30:00Z", "2023-01-15"
func parseISOTimestamp(timestamp string) (time.Time, error) {
	if timestamp == "" {
		return time.Time{}, fmt.Errorf("empty timestamp")
	}

	// Try RFC3339 format (standard ISO 8601 with time)
	t, err := time.Parse(time.RFC3339, timestamp)
	if err == nil {
		return t, nil
	}

	// Try simple date format (YYYY-MM-DD)
	t, err = time.Parse("2006-01-02", timestamp)
	if err == nil {
		return t, nil
	}

	// Return error with supported formats
	return time.Time{}, fmt.Errorf("invalid ISO 8601 timestamp: %s (supported formats: YYYY-MM-DDThh:mm:ssZ or YYYY-MM-DD)", timestamp)
}

func AssignCodingAgentPrompt(t translations.TranslationHelperFunc) (tool *mcp.Prompt, handler mcp.PromptHandler) {
	// return mcp.NewPrompt("AssignCodingAgent",
	// 		mcp.WithPromptDescription(t("PROMPT_ASSIGN_CODING_AGENT_DESCRIPTION", "Assign GitHub Coding Agent to multiple tasks in a GitHub repository.")),
	// 		mcp.WithArgument("repo", mcp.ArgumentDescription("The repository to assign tasks in (owner/repo)."), mcp.RequiredArgument()),
	// 	),
	return &mcp.Prompt{
			Name:        "AssignCodingAgent",
			Description: t("PROMPT_ASSIGN_CODING_AGENT_DESCRIPTION", "Assign GitHub Coding Agent to multiple tasks in a GitHub repository."),
			Arguments: []*mcp.PromptArgument{
				{
					Name:        "repo",
					Description: t("PROMPT_ASSIGN_CODING_AGENT_REPO_DESC", "The repository to assign tasks in (owner/repo)."),
					Required:    true,
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
			repo := request.Arguments["repo"]

			messages := []*mcp.PromptMessage{
				{
					Role:    "system",
					Content: &mcp.TextContent{Text: "You are a personal assistant for GitHub the Copilot GitHub Coding Agent. Your task is to help the user assign tasks to the Coding Agent based on their open GitHub issues. You can use `assign_copilot_to_issue` tool to assign the Coding Agent to issues that are suitable for autonomous work, and `search_issues` tool to find issues that match the user's criteria. You can also use `list_issues` to get a list of issues in the repository."},
				},
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Please go and get a list of the most recent 10 issues from the %s GitHub repository", repo)},
				},
				{
					Role:    "assistant",
					Content: &mcp.TextContent{Text: fmt.Sprintf("Sure! I will get a list of the 10 most recent issues for the repo %s.", repo)},
				},
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "For each issue, please check if it is a clearly defined coding task with acceptance criteria and a low to medium complexity to identify issues that are suitable for an AI Coding Agent to work on. Then assign each of the identified issues to Copilot."},
				},
				{
					Role:    "assistant",
					Content: &mcp.TextContent{Text: "Certainly! Let me carefully check which ones are clearly scoped issues that are good to assign to the coding agent, and I will summarize and assign them now."},
				},
				{
					Role:    "user",
					Content: &mcp.TextContent{Text: "Great, if you are unsure if an issue is good to assign, ask me first, rather than assigning copilot. If you are certain the issue is clear and suitable you can assign it to Copilot without asking."},
				},
			}
			return &mcp.GetPromptResult{
				Messages: messages,
			}, nil
		}
}
