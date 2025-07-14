package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-viper/mapstructure/v2"
	"github.com/google/go-github/v72/github"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/shurcooL/githubv4"

	ghErrors "github.com/github/github-mcp-server/pkg/errors"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/github/github-mcp-server/pkg/utils"
)

// GetPullRequest creates a tool to get details of a specific pull request.
func GetPullRequest(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_pull_request",
			Description: t("TOOL_GET_PULL_REQUEST_DESCRIPTION", "Get details of a specific pull request in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_PULL_REQUEST_USER_TITLE", "Get pull request details"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			pr, resp, err := client.PullRequests.Get(ctx, owner, repo, pullNumber)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get pull request",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get pull request: %s", string(body))), nil
			}

			r, err := json.Marshal(pr)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// CreatePullRequest creates a tool to create a new pull request.
func CreatePullRequest(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "create_pull_request",
			Description: t("TOOL_CREATE_PULL_REQUEST_DESCRIPTION", "Create a new pull request in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_CREATE_PULL_REQUEST_USER_TITLE", "Open new pull request"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "title", "head", "base"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"title": {
						Type:        "string",
						Description: "PR title",
					},
					"body": {
						Type:        "string",
						Description: "PR description",
					},
					"head": {
						Type:        "string",
						Description: "Branch containing changes",
					},
					"base": {
						Type:        "string",
						Description: "Branch to merge into",
					},
					"draft": {
						Type:        "boolean",
						Description: "Create as draft PR",
					},
					"maintainer_can_modify": {
						Type:        "boolean",
						Description: "Allow maintainer edits",
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
			head, err := RequiredParam[string](request, "head")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			base, err := RequiredParam[string](request, "base")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			body, err := OptionalParam[string](request, "body")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			draft, err := OptionalParam[bool](request, "draft")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			maintainerCanModify, err := OptionalParam[bool](request, "maintainer_can_modify")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			newPR := &github.NewPullRequest{
				Title: github.Ptr(title),
				Head:  github.Ptr(head),
				Base:  github.Ptr(base),
			}

			if body != "" {
				newPR.Body = github.Ptr(body)
			}

			newPR.Draft = github.Ptr(draft)
			newPR.MaintainerCanModify = github.Ptr(maintainerCanModify)

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			pr, resp, err := client.PullRequests.Create(ctx, owner, repo, newPR)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to create pull request",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to create pull request: %s", string(body))), nil
			}

			r, err := json.Marshal(pr)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// UpdatePullRequest creates a tool to update an existing pull request.
func UpdatePullRequest(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "update_pull_request",
			Description: t("TOOL_UPDATE_PULL_REQUEST_DESCRIPTION", "Update an existing pull request in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_UPDATE_PULL_REQUEST_USER_TITLE", "Edit pull request"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number to update",
					},
					"title": {
						Type:        "string",
						Description: "New title",
					},
					"body": {
						Type:        "string",
						Description: "New description",
					},
					"state": {
						Type:        "string",
						Description: "New state",
						Enum:        []any{"open", "closed"},
					},
					"base": {
						Type:        "string",
						Description: "New base branch name",
					},
					"maintainer_can_modify": {
						Type:        "boolean",
						Description: "Allow maintainer edits",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Build the update struct only with provided fields
			update := &github.PullRequest{}
			updateNeeded := false

			if title, ok, err := OptionalParamOK[string](request, "title"); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			} else if ok {
				update.Title = github.Ptr(title)
				updateNeeded = true
			}

			if body, ok, err := OptionalParamOK[string](request, "body"); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			} else if ok {
				update.Body = github.Ptr(body)
				updateNeeded = true
			}

			if state, ok, err := OptionalParamOK[string](request, "state"); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			} else if ok {
				update.State = github.Ptr(state)
				updateNeeded = true
			}

			if base, ok, err := OptionalParamOK[string](request, "base"); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			} else if ok {
				update.Base = &github.PullRequestBranch{Ref: github.Ptr(base)}
				updateNeeded = true
			}

			if maintainerCanModify, ok, err := OptionalParamOK[bool](request, "maintainer_can_modify"); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			} else if ok {
				update.MaintainerCanModify = github.Ptr(maintainerCanModify)
				updateNeeded = true
			}

			if !updateNeeded {
				return utils.NewToolResultError("No update parameters provided."), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			pr, resp, err := client.PullRequests.Edit(ctx, owner, repo, pullNumber, update)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to update pull request",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to update pull request: %s", string(body))), nil
			}

			r, err := json.Marshal(pr)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// ListPullRequests creates a tool to list and filter repository pull requests.
func ListPullRequests(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "list_pull_requests",
			Description: t("TOOL_LIST_PULL_REQUESTS_DESCRIPTION", "List pull requests in a GitHub repository. If the user specifies an author, then DO NOT use this tool and use the search_pull_requests tool instead."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_LIST_PULL_REQUESTS_USER_TITLE", "List pull requests"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"state": {
						Type:        "string",
						Description: "Filter by state",
						Enum:        []any{"open", "closed", "all"},
					},
					"head": {
						Type:        "string",
						Description: "Filter by head user/org and branch",
					},
					"base": {
						Type:        "string",
						Description: "Filter by base branch",
					},
					"sort": {
						Type:        "string",
						Description: "Sort by",
						Enum:        []any{"created", "updated", "popularity", "long-running"},
					},
					"direction": {
						Type:        "string",
						Description: "Sort direction",
						Enum:        []any{"asc", "desc"},
					},
					"page": {
						Type:        "number",
						Description: "Page number for pagination",
					},
					"perPage": {
						Type:        "number",
						Description: "Items per page for pagination",
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
			state, err := OptionalParam[string](request, "state")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			head, err := OptionalParam[string](request, "head")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			base, err := OptionalParam[string](request, "base")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			direction, err := OptionalParam[string](request, "direction")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			opts := &github.PullRequestListOptions{
				State:     state,
				Head:      head,
				Base:      base,
				Sort:      sort,
				Direction: direction,
				ListOptions: github.ListOptions{
					PerPage: pagination.perPage,
					Page:    pagination.page,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			prs, resp, err := client.PullRequests.List(ctx, owner, repo, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to list pull requests",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to list pull requests: %s", string(body))), nil
			}

			r, err := json.Marshal(prs)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// MergePullRequest creates a tool to merge a pull request.
func MergePullRequest(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "merge_pull_request",
			Description: t("TOOL_MERGE_PULL_REQUEST_DESCRIPTION", "Merge a pull request in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_MERGE_PULL_REQUEST_USER_TITLE", "Merge pull request"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"commit_title": {
						Type:        "string",
						Description: "Title for merge commit",
					},
					"commit_message": {
						Type:        "string",
						Description: "Extra detail for merge commit",
					},
					"merge_method": {
						Type:        "string",
						Description: "Merge method",
						Enum:        []any{"merge", "squash", "rebase"},
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			commitTitle, err := OptionalParam[string](request, "commit_title")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			commitMessage, err := OptionalParam[string](request, "commit_message")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			mergeMethod, err := OptionalParam[string](request, "merge_method")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			options := &github.PullRequestOptions{
				CommitTitle: commitTitle,
				MergeMethod: mergeMethod,
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			result, resp, err := client.PullRequests.Merge(ctx, owner, repo, pullNumber, commitMessage, options)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to merge pull request",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to merge pull request: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// SearchPullRequests creates a tool to search for pull requests.
func SearchPullRequests(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "search_pull_requests",
			Description: t("TOOL_SEARCH_PULL_REQUESTS_DESCRIPTION", "Search for pull requests in GitHub repositories using issues search syntax already scoped to is:pr"),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_SEARCH_PULL_REQUESTS_USER_TITLE", "Search pull requests"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"query"},
				Properties: map[string]*jsonschema.Schema{
					"query": {
						Type:        "string",
						Description: "Search query using GitHub pull request search syntax",
					},
					"owner": {
						Type:        "string",
						Description: "Optional repository owner. If provided with repo, only notifications for this repository are listed.",
					},
					"repo": {
						Type:        "string",
						Description: "Optional repository name. If provided with owner, only notifications for this repository are listed.",
					},
					"sort": {
						Type:        "string",
						Description: "Sort field by number of matches of categories, defaults to best match",
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
						Description: "Sort order",
						Enum:        []any{"asc", "desc"},
					},
					"page": {
						Type:        "number",
						Description: "Page number for pagination",
					},
					"perPage": {
						Type:        "number",
						Description: "Items per page for pagination",
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			return searchHandler(ctx, getClient, request, "pr", "failed to search pull requests")
		}
}

// GetPullRequestFiles creates a tool to get the list of files changed in a pull request.
func GetPullRequestFiles(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_pull_request_files",
			Description: t("TOOL_GET_PULL_REQUEST_FILES_DESCRIPTION", "Get the files changed in a specific pull request."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_PULL_REQUEST_FILES_USER_TITLE", "Get pull request files"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"page": {
						Type:        "number",
						Description: "Page number for pagination",
					},
					"perPage": {
						Type:        "number",
						Description: "Items per page for pagination",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			opts := &github.ListOptions{
				PerPage: pagination.perPage,
				Page:    pagination.page,
			}
			files, resp, err := client.PullRequests.ListFiles(ctx, owner, repo, pullNumber, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get pull request files",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get pull request files: %s", string(body))), nil
			}

			r, err := json.Marshal(files)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// GetPullRequestStatus creates a tool to get the combined status of all status checks for a pull request.
func GetPullRequestStatus(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_pull_request_status",
			Description: t("TOOL_GET_PULL_REQUEST_STATUS_DESCRIPTION", "Get the status of a specific pull request."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_PULL_REQUEST_STATUS_USER_TITLE", "Get pull request status checks"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			// First get the PR to find the head SHA
			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			pr, resp, err := client.PullRequests.Get(ctx, owner, repo, pullNumber)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get pull request",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get pull request: %s", string(body))), nil
			}

			// Get combined status for the head SHA
			status, resp, err := client.Repositories.GetCombinedStatus(ctx, owner, repo, *pr.Head.SHA, nil)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get combined status",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get combined status: %s", string(body))), nil
			}

			r, err := json.Marshal(status)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// UpdatePullRequestBranch creates a tool to update a pull request branch with the latest changes from the base branch.
func UpdatePullRequestBranch(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "update_pull_request_branch",
			Description: t("TOOL_UPDATE_PULL_REQUEST_BRANCH_DESCRIPTION", "Update the branch of a pull request with the latest changes from the base branch."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_UPDATE_PULL_REQUEST_BRANCH_USER_TITLE", "Update pull request branch"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"expectedHeadSha": {
						Type:        "string",
						Description: "The expected SHA of the pull request's HEAD ref",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			expectedHeadSHA, err := OptionalParam[string](request, "expectedHeadSha")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			opts := &github.PullRequestBranchUpdateOptions{}
			if expectedHeadSHA != "" {
				opts.ExpectedHeadSHA = github.Ptr(expectedHeadSHA)
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			result, resp, err := client.PullRequests.UpdateBranch(ctx, owner, repo, pullNumber, opts)
			if err != nil {
				// Check if it's an acceptedError. An acceptedError indicates that the update is in progress,
				// and it's not a real error.
				if resp != nil && resp.StatusCode == http.StatusAccepted && isAcceptedError(err) {
					return utils.NewToolResultText("Pull request branch update is in progress"), nil
				}
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to update pull request branch",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusAccepted {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to update pull request branch: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// GetPullRequestComments creates a tool to get the review comments on a pull request.
func GetPullRequestComments(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_pull_request_comments",
			Description: t("TOOL_GET_PULL_REQUEST_COMMENTS_DESCRIPTION", "Get comments for a specific pull request."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_PULL_REQUEST_COMMENTS_USER_TITLE", "Get pull request comments"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts := &github.PullRequestListCommentsOptions{
				ListOptions: github.ListOptions{
					PerPage: 100,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			comments, resp, err := client.PullRequests.ListComments(ctx, owner, repo, pullNumber, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get pull request comments",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get pull request comments: %s", string(body))), nil
			}

			r, err := json.Marshal(comments)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// GetPullRequestReviews creates a tool to get the reviews on a pull request.
func GetPullRequestReviews(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {

	return &mcp.Tool{
			Name:        "get_pull_request_reviews",
			Description: t("TOOL_GET_PULL_REQUEST_REVIEWS_DESCRIPTION", "Get reviews for a specific pull request."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_PULL_REQUEST_REVIEWS_USER_TITLE", "Get pull request reviews"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
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
			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			reviews, resp, err := client.PullRequests.ListReviews(ctx, owner, repo, pullNumber, nil)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get pull request reviews",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get pull request reviews: %s", string(body))), nil
			}

			r, err := json.Marshal(reviews)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

func CreateAndSubmitPullRequestReview(getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "create_and_submit_pull_request_review",
			Description: t("TOOL_CREATE_AND_SUBMIT_PULL_REQUEST_REVIEW_DESCRIPTION", "Create and submit a review for a pull request without review comments."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_CREATE_AND_SUBMIT_PULL_REQUEST_REVIEW_USER_TITLE", "Create and submit a pull request review without comments"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber", "body", "event"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"body": {
						Type:        "string",
						Description: "Review comment text",
					},
					"event": {
						Type:        "string",
						Description: "Review action to perform",
						Enum:        []any{"APPROVE", "REQUEST_CHANGES", "COMMENT"},
					},
					"commitID": {
						Type:        "string",
						Description: "SHA of commit to review",
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner      string
				Repo       string
				PullNumber int32
				Body       string
				Event      string
				CommitID   *string
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Given our owner, repo and PR number, lookup the GQL ID of the PR.
			client, err := getGQLClient(ctx)
			if err != nil {
				return utils.NewToolResultError(fmt.Sprintf("failed to get GitHub GQL client: %v", err)), nil
			}

			var getPullRequestQuery struct {
				Repository struct {
					PullRequest struct {
						ID githubv4.ID
					} `graphql:"pullRequest(number: $prNum)"`
				} `graphql:"repository(owner: $owner, name: $repo)"`
			}
			if err := client.Query(ctx, &getPullRequestQuery, map[string]any{
				"owner": githubv4.String(params.Owner),
				"repo":  githubv4.String(params.Repo),
				"prNum": githubv4.Int(params.PullNumber),
			}); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get pull request",
					err,
				), nil
			}

			// Now we have the GQL ID, we can create a review
			var addPullRequestReviewMutation struct {
				AddPullRequestReview struct {
					PullRequestReview struct {
						ID githubv4.ID // We don't need this, but a selector is required or GQL complains.
					}
				} `graphql:"addPullRequestReview(input: $input)"`
			}

			if err := client.Mutate(
				ctx,
				&addPullRequestReviewMutation,
				githubv4.AddPullRequestReviewInput{
					PullRequestID: getPullRequestQuery.Repository.PullRequest.ID,
					Body:          githubv4.NewString(githubv4.String(params.Body)),
					Event:         newGQLStringlike[githubv4.PullRequestReviewEvent](params.Event),
					CommitOID:     newGQLStringlikePtr[githubv4.GitObjectID](params.CommitID),
				},
				nil,
			); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Return nothing interesting, just indicate success for the time being.
			// In future, we may want to return the review ID, but for the moment, we're not leaking
			// API implementation details to the LLM.
			return utils.NewToolResultText("pull request review submitted successfully"), nil
		}
}

// CreatePendingPullRequestReview creates a tool to create a pending review on a pull request.
func CreatePendingPullRequestReview(getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "create_pending_pull_request_review",
			Description: t("TOOL_CREATE_PENDING_PULL_REQUEST_REVIEW_DESCRIPTION", "Create a pending review for a pull request. Call this first before attempting to add comments to a pending review, and ultimately submitting it. A pending pull request review means a pull request review, it is pending because you create it first and submit it later, and the PR author will not see it until it is submitted."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_CREATE_PENDING_PULL_REQUEST_REVIEW_USER_TITLE", "Create pending pull request review"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"commitID": {
						Type:        "string",
						Description: "SHA of commit to review, optional, if not provided, the latest commit will be used.",
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner      string
				Repo       string
				PullNumber int32
				CommitID   *string
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Given our owner, repo and PR number, lookup the GQL ID of the PR.
			client, err := getGQLClient(ctx)
			if err != nil {
				return utils.NewToolResultError(fmt.Sprintf("failed to get GitHub GQL client: %v", err)), nil
			}

			var getPullRequestQuery struct {
				Repository struct {
					PullRequest struct {
						ID githubv4.ID
					} `graphql:"pullRequest(number: $prNum)"`
				} `graphql:"repository(owner: $owner, name: $repo)"`
			}
			if err := client.Query(ctx, &getPullRequestQuery, map[string]any{
				"owner": githubv4.String(params.Owner),
				"repo":  githubv4.String(params.Repo),
				"prNum": githubv4.Int(params.PullNumber),
			}); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get pull request",
					err,
				), nil
			}

			// Now we have the GQL ID, we can create a pending review
			var addPullRequestReviewMutation struct {
				AddPullRequestReview struct {
					PullRequestReview struct {
						ID githubv4.ID // We don't need this, but a selector is required or GQL complains.
					}
				} `graphql:"addPullRequestReview(input: $input)"`
			}

			if err := client.Mutate(
				ctx,
				&addPullRequestReviewMutation,
				githubv4.AddPullRequestReviewInput{
					PullRequestID: getPullRequestQuery.Repository.PullRequest.ID,
					CommitOID:     newGQLStringlikePtr[githubv4.GitObjectID](params.CommitID),
				},
				nil,
			); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Return nothing interesting, just indicate success for the time being.
			// In future, we may want to return the review ID, but for the moment, we're not leaking
			// API implementation details to the LLM.
			return utils.NewToolResultText("pending pull request created"), nil
		}
}

// AddPullRequestReviewCommentToPendingReview creates a tool to add a comment to a pull request review.
func AddPullRequestReviewCommentToPendingReview(getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "add_pull_request_review_comment_to_pending_review",
			Description: t("TOOL_ADD_PULL_REQUEST_REVIEW_COMMENT_TO_PENDING_REVIEW_DESCRIPTION", "Add a comment to the requester's latest pending pull request review, a pending review needs to already exist to call this (check with the user if not sure)."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_ADD_PULL_REQUEST_REVIEW_COMMENT_TO_PENDING_REVIEW_USER_TITLE", "Add comment to the requester's latest pending pull request review"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber", "path", "body", "subjectType"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"path": {
						Type:        "string",
						Description: "The relative path to the file that necessitates a comment",
					},
					"body": {
						Type:        "string",
						Description: "The text of the review comment",
					},
					"subjectType": {
						Type:        "string",
						Description: "The level at which the comment is targeted",
						Enum:        []any{"FILE", "LINE"},
					},
					"line": {
						Type:        "number",
						Description: "The line of the blob in the pull request diff that the comment applies to. For multi-line comments, the last line of the range",
					},
					"side": {
						Type:        "string",
						Description: "The side of the diff to comment on. LEFT indicates the previous state, RIGHT indicates the new state",
						Enum:        []any{"LEFT", "RIGHT"},
					},
					"startLine": {
						Type:        "number",
						Description: "For multi-line comments, the first line of the range that the comment applies to",
					},
					"startSide": {
						Type:        "string",
						Description: "For multi-line comments, the starting side of the diff that the comment applies to. LEFT indicates the previous state, RIGHT indicates the new state",
						Enum:        []any{"LEFT", "RIGHT"},
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner       string
				Repo        string
				PullNumber  int32
				Path        string
				Body        string
				SubjectType string
				Line        *int32
				Side        *string
				StartLine   *int32
				StartSide   *string
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getGQLClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub GQL client: %w", err)
			}

			// First we'll get the current user
			var getViewerQuery struct {
				Viewer struct {
					Login githubv4.String
				}
			}

			if err := client.Query(ctx, &getViewerQuery, nil); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get current user",
					err,
				), nil
			}

			var getLatestReviewForViewerQuery struct {
				Repository struct {
					PullRequest struct {
						Reviews struct {
							Nodes []struct {
								ID    githubv4.ID
								State githubv4.PullRequestReviewState
								URL   githubv4.URI
							}
						} `graphql:"reviews(first: 1, author: $author)"`
					} `graphql:"pullRequest(number: $prNum)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			}

			vars := map[string]any{
				"author": githubv4.String(getViewerQuery.Viewer.Login),
				"owner":  githubv4.String(params.Owner),
				"name":   githubv4.String(params.Repo),
				"prNum":  githubv4.Int(params.PullNumber),
			}

			if err := client.Query(context.Background(), &getLatestReviewForViewerQuery, vars); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get latest review for current user",
					err,
				), nil
			}

			// Validate there is one review and the state is pending
			if len(getLatestReviewForViewerQuery.Repository.PullRequest.Reviews.Nodes) == 0 {
				return utils.NewToolResultError("No pending review found for the viewer"), nil
			}

			review := getLatestReviewForViewerQuery.Repository.PullRequest.Reviews.Nodes[0]
			if review.State != githubv4.PullRequestReviewStatePending {
				errText := fmt.Sprintf("The latest review, found at %s is not pending", review.URL)
				return utils.NewToolResultError(errText), nil
			}

			// Then we can create a new review thread comment on the review.
			var addPullRequestReviewThreadMutation struct {
				AddPullRequestReviewThread struct {
					Thread struct {
						ID githubv4.ID // We don't need this, but a selector is required or GQL complains.
					}
				} `graphql:"addPullRequestReviewThread(input: $input)"`
			}

			if err := client.Mutate(
				ctx,
				&addPullRequestReviewThreadMutation,
				githubv4.AddPullRequestReviewThreadInput{
					Path:                githubv4.String(params.Path),
					Body:                githubv4.String(params.Body),
					SubjectType:         newGQLStringlikePtr[githubv4.PullRequestReviewThreadSubjectType](&params.SubjectType),
					Line:                newGQLIntPtr(params.Line),
					Side:                newGQLStringlikePtr[githubv4.DiffSide](params.Side),
					StartLine:           newGQLIntPtr(params.StartLine),
					StartSide:           newGQLStringlikePtr[githubv4.DiffSide](params.StartSide),
					PullRequestReviewID: &review.ID,
				},
				nil,
			); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Return nothing interesting, just indicate success for the time being.
			// In future, we may want to return the review ID, but for the moment, we're not leaking
			// API implementation details to the LLM.
			return utils.NewToolResultText("pull request review comment successfully added to pending review"), nil
		}
}

// SubmitPendingPullRequestReview creates a tool to submit a pull request review.
func SubmitPendingPullRequestReview(getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "submit_pending_pull_request_review",
			Description: t("TOOL_SUBMIT_PENDING_PULL_REQUEST_REVIEW_DESCRIPTION", "Submit the requester's latest pending pull request review, normally this is a final step after creating a pending review, adding comments first, unless you know that the user already did the first two steps, you should check before calling this."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_SUBMIT_PENDING_PULL_REQUEST_REVIEW_USER_TITLE", "Submit the requester's latest pending pull request review"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber", "event"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
					"event": {
						Type:        "string",
						Description: "The event to perform",
						Enum:        []any{"APPROVE", "REQUEST_CHANGES", "COMMENT"},
					},
					"body": {
						Type:        "string",
						Description: "The text of the review comment, optional, if not provided, no body will be set.",
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner      string
				Repo       string
				PullNumber int32
				Event      string
				Body       *string
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getGQLClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub GQL client: %w", err)
			}

			// First we'll get the current user
			var getViewerQuery struct {
				Viewer struct {
					Login githubv4.String
				}
			}

			if err := client.Query(ctx, &getViewerQuery, nil); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get current user",
					err,
				), nil
			}

			var getLatestReviewForViewerQuery struct {
				Repository struct {
					PullRequest struct {
						Reviews struct {
							Nodes []struct {
								ID    githubv4.ID
								State githubv4.PullRequestReviewState
								URL   githubv4.URI
							}
						} `graphql:"reviews(first: 1, author: $author)"`
					} `graphql:"pullRequest(number: $prNum)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			}

			vars := map[string]any{
				"author": githubv4.String(getViewerQuery.Viewer.Login),
				"owner":  githubv4.String(params.Owner),
				"name":   githubv4.String(params.Repo),
				"prNum":  githubv4.Int(params.PullNumber),
			}

			if err := client.Query(context.Background(), &getLatestReviewForViewerQuery, vars); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get latest review for current user",
					err,
				), nil
			}

			// Validate there is one review and the state is pending
			if len(getLatestReviewForViewerQuery.Repository.PullRequest.Reviews.Nodes) == 0 {
				return utils.NewToolResultError("No pending review found for the viewer"), nil
			}

			review := getLatestReviewForViewerQuery.Repository.PullRequest.Reviews.Nodes[0]
			if review.State != githubv4.PullRequestReviewStatePending {
				errText := fmt.Sprintf("The latest review, found at %s is not pending", review.URL)
				return utils.NewToolResultError(errText), nil
			}

			// Prepare the mutation
			var submitPullRequestReviewMutation struct {
				SubmitPullRequestReview struct {
					PullRequestReview struct {
						ID githubv4.ID // We don't need this, but a selector is required or GQL complains.
					}
				} `graphql:"submitPullRequestReview(input: $input)"`
			}

			if err := client.Mutate(
				ctx,
				&submitPullRequestReviewMutation,
				githubv4.SubmitPullRequestReviewInput{
					PullRequestReviewID: &review.ID,
					Event:               githubv4.PullRequestReviewEvent(params.Event),
					Body:                newGQLStringlikePtr[githubv4.String](params.Body),
				},
				nil,
			); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to submit pull request review",
					err,
				), nil
			}

			// Return nothing interesting, just indicate success for the time being.
			// In future, we may want to return the review ID, but for the moment, we're not leaking
			// API implementation details to the LLM.
			return utils.NewToolResultText("pending pull request review successfully submitted"), nil
		}
}

func DeletePendingPullRequestReview(getGQLClient GetGQLClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "delete_pending_pull_request_review",
			Description: t("TOOL_DELETE_PENDING_PULL_REQUEST_REVIEW_DESCRIPTION", "Delete the requester's latest pending pull request review. Use this after the user decides not to submit a pending review, if you don't know if they already created one then check first."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_DELETE_PENDING_PULL_REQUEST_REVIEW_USER_TITLE", "Delete the requester's latest pending pull request review"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner      string
				Repo       string
				PullNumber int32
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getGQLClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub GQL client: %w", err)
			}

			// First we'll get the current user
			var getViewerQuery struct {
				Viewer struct {
					Login githubv4.String
				}
			}

			if err := client.Query(ctx, &getViewerQuery, nil); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get current user",
					err,
				), nil
			}

			var getLatestReviewForViewerQuery struct {
				Repository struct {
					PullRequest struct {
						Reviews struct {
							Nodes []struct {
								ID    githubv4.ID
								State githubv4.PullRequestReviewState
								URL   githubv4.URI
							}
						} `graphql:"reviews(first: 1, author: $author)"`
					} `graphql:"pullRequest(number: $prNum)"`
				} `graphql:"repository(owner: $owner, name: $name)"`
			}

			vars := map[string]any{
				"author": githubv4.String(getViewerQuery.Viewer.Login),
				"owner":  githubv4.String(params.Owner),
				"name":   githubv4.String(params.Repo),
				"prNum":  githubv4.Int(params.PullNumber),
			}

			if err := client.Query(context.Background(), &getLatestReviewForViewerQuery, vars); err != nil {
				return ghErrors.NewGitHubGraphQLErrorResponse(ctx,
					"failed to get latest review for current user",
					err,
				), nil
			}

			// Validate there is one review and the state is pending
			if len(getLatestReviewForViewerQuery.Repository.PullRequest.Reviews.Nodes) == 0 {
				return utils.NewToolResultError("No pending review found for the viewer"), nil
			}

			review := getLatestReviewForViewerQuery.Repository.PullRequest.Reviews.Nodes[0]
			if review.State != githubv4.PullRequestReviewStatePending {
				errText := fmt.Sprintf("The latest review, found at %s is not pending", review.URL)
				return utils.NewToolResultError(errText), nil
			}

			// Prepare the mutation
			var deletePullRequestReviewMutation struct {
				DeletePullRequestReview struct {
					PullRequestReview struct {
						ID githubv4.ID // We don't need this, but a selector is required or GQL complains.
					}
				} `graphql:"deletePullRequestReview(input: $input)"`
			}

			if err := client.Mutate(
				ctx,
				&deletePullRequestReviewMutation,
				githubv4.DeletePullRequestReviewInput{
					PullRequestReviewID: &review.ID,
				},
				nil,
			); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			// Return nothing interesting, just indicate success for the time being.
			// In future, we may want to return the review ID, but for the moment, we're not leaking
			// API implementation details to the LLM.
			return utils.NewToolResultText("pending pull request review successfully deleted"), nil
		}
}

func GetPullRequestDiff(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_pull_request_diff",
			Description: t("TOOL_GET_PULL_REQUEST_DIFF_DESCRIPTION", "Get the diff of a pull request."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_PULL_REQUEST_DIFF_USER_TITLE", "Get pull request diff"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
					},
				},
			},
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			var params struct {
				Owner      string
				Repo       string
				PullNumber int32
			}
			if err := mapstructure.Decode(request.Arguments, &params); err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return utils.NewToolResultError(fmt.Sprintf("failed to get GitHub client: %v", err)), nil
			}

			raw, resp, err := client.PullRequests.GetRaw(
				ctx,
				params.Owner,
				params.Repo,
				int(params.PullNumber),
				github.RawOptions{Type: github.Diff},
			)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to get pull request diff",
					resp,
					err,
				), nil
			}

			if resp.StatusCode != http.StatusOK {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to get pull request diff: %s", string(body))), nil
			}

			defer func() { _ = resp.Body.Close() }()

			// Return the raw response
			return utils.NewToolResultText(string(raw)), nil
		}
}

// RequestCopilotReview creates a tool to request a Copilot review for a pull request.
// Note that this tool will not work on GHES where this feature is unsupported. In future, we should not expose this
// tool if the configured host does not support it.
func RequestCopilotReview(getClient GetClientFn, t translations.TranslationHelperFunc) (*mcp.Tool, mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "request_copilot_review",
			Description: t("TOOL_REQUEST_COPILOT_REVIEW_DESCRIPTION", "Request a GitHub Copilot code review for a pull request. Use this for automated feedback on pull requests, usually before requesting a human reviewer."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_REQUEST_COPILOT_REVIEW_USER_TITLE", "Request Copilot review"),
				ReadOnlyHint: false,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "pullNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "Repository owner",
					},
					"repo": {
						Type:        "string",
						Description: "Repository name",
					},
					"pullNumber": {
						Type:        "number",
						Description: "Pull request number",
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

			pullNumber, err := RequiredInt(request, "pullNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			_, resp, err := client.PullRequests.RequestReviewers(
				ctx,
				owner,
				repo,
				pullNumber,
				github.ReviewersRequest{
					// The login name of the copilot reviewer bot
					Reviewers: []string{"copilot-pull-request-reviewer[bot]"},
				},
			)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					"failed to request copilot review",
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusCreated {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to request copilot review: %s", string(body))), nil
			}

			// Return nothing on success, as there's not much value in returning the Pull Request itself
			return utils.NewToolResultText(""), nil
		}
}

// newGQLString like takes something that approximates a string (of which there are many types in shurcooL/githubv4)
// and constructs a pointer to it, or nil if the string is empty. This is extremely useful because when we parse
// params from the MCP request, we need to convert them to types that are pointers of type def strings and it's
// not possible to take a pointer of an anonymous value e.g. &githubv4.String("foo").
func newGQLStringlike[T ~string](s string) *T {
	if s == "" {
		return nil
	}
	stringlike := T(s)
	return &stringlike
}

func newGQLStringlikePtr[T ~string](s *string) *T {
	if s == nil {
		return nil
	}
	stringlike := T(*s)
	return &stringlike
}

func newGQLIntPtr(i *int32) *githubv4.Int {
	if i == nil {
		return nil
	}
	gi := githubv4.Int(*i)
	return &gi
}
