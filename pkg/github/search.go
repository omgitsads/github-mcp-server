package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	ghErrors "github.com/github/github-mcp-server/pkg/errors"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/github/github-mcp-server/pkg/utils"
	"github.com/google/go-github/v72/github"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// SearchRepositories creates a tool to search for GitHub repositories.
func SearchRepositories(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "search_repositories",
			Description: t("TOOL_SEARCH_REPOSITORIES_DESCRIPTION", "Search for GitHub repositories"),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_SEARCH_REPOSITORIES_USER_TITLE", "Search repositories"),
				ReadOnlyHint: true,
			},
			InputSchema: WithPagination(&jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"query": {
						Type:        "string",
						Description: "Search query",
					},
				},
				Required: []string{"query"},
			}),
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			query, err := RequiredParam[string](request, "query")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts := &github.SearchOptions{
				ListOptions: github.ListOptions{
					Page:    pagination.page,
					PerPage: pagination.perPage,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			result, resp, err := client.Search.Repositories(ctx, query, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					fmt.Sprintf("failed to search repositories with query '%s'", query),
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to search repositories: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// SearchCode creates a tool to search for code across GitHub repositories.
func SearchCode(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "search_code",
			Description: t("TOOL_SEARCH_CODE_DESCRIPTION", "Search for code across GitHub repositories"),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_SEARCH_CODE_USER_TITLE", "Search code"),
				ReadOnlyHint: true,
			},
			InputSchema: WithPagination(&jsonschema.Schema{
				Type: "object",
				Properties: map[string]*jsonschema.Schema{
					"q": {
						Type:        "string",
						Description: "Search query using GitHub code search syntax",
					},
					"sort": {
						Type:        "string",
						Description: "Sort field ('indexed' only)",
					},
					"order": {
						Type:        "string",
						Description: "Sort order",
						Enum:        []any{"asc", "desc"},
					},
				},
				Required: []string{"q"},
			}),
		},
		func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
			query, err := RequiredParam[string](request, "q")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			sort, err := OptionalParam[string](request, "sort")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			order, err := OptionalParam[string](request, "order")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			pagination, err := OptionalPaginationParams(request)
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			opts := &github.SearchOptions{
				Sort:  sort,
				Order: order,
				ListOptions: github.ListOptions{
					PerPage: pagination.perPage,
					Page:    pagination.page,
				},
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			result, resp, err := client.Search.Code(ctx, query, opts)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					fmt.Sprintf("failed to search code with query '%s'", query),
					resp,
					err,
				), nil
			}
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != 200 {
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("failed to read response body: %w", err)
				}
				return utils.NewToolResultError(fmt.Sprintf("failed to search code: %s", string(body))), nil
			}

			r, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal response: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

// MinimalUser is the output type for user and organization search results.
type MinimalUser struct {
	Login      string       `json:"login"`
	ID         int64        `json:"id,omitempty"`
	ProfileURL string       `json:"profile_url,omitempty"`
	AvatarURL  string       `json:"avatar_url,omitempty"`
	Details    *UserDetails `json:"details,omitempty"` // Optional field for additional user details
}

type MinimalSearchUsersResult struct {
	TotalCount        int           `json:"total_count"`
	IncompleteResults bool          `json:"incomplete_results"`
	Items             []MinimalUser `json:"items"`
}

func userOrOrgHandler(accountType string, getClient GetClientFn) mcp.ToolHandler {
	return func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error) {
		query, err := RequiredParam[string](request, "query")
		if err != nil {
			return utils.NewToolResultError(err.Error()), nil
		}
		sort, err := OptionalParam[string](request, "sort")
		if err != nil {
			return utils.NewToolResultError(err.Error()), nil
		}
		order, err := OptionalParam[string](request, "order")
		if err != nil {
			return utils.NewToolResultError(err.Error()), nil
		}
		pagination, err := OptionalPaginationParams(request)
		if err != nil {
			return utils.NewToolResultError(err.Error()), nil
		}

		opts := &github.SearchOptions{
			Sort:  sort,
			Order: order,
			ListOptions: github.ListOptions{
				PerPage: pagination.perPage,
				Page:    pagination.page,
			},
		}

		client, err := getClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get GitHub client: %w", err)
		}

		searchQuery := "type:" + accountType + " " + query
		result, resp, err := client.Search.Users(ctx, searchQuery, opts)
		if err != nil {
			return ghErrors.NewGitHubAPIErrorResponse(ctx,
				fmt.Sprintf("failed to search %ss with query '%s'", accountType, query),
				resp,
				err,
			), nil
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != 200 {
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("failed to read response body: %w", err)
			}
			return utils.NewToolResultError(fmt.Sprintf("failed to search %ss: %s", accountType, string(body))), nil
		}

		minimalUsers := make([]MinimalUser, 0, len(result.Users))

		for _, user := range result.Users {
			if user.Login != nil {
				mu := MinimalUser{
					Login:      user.GetLogin(),
					ID:         user.GetID(),
					ProfileURL: user.GetHTMLURL(),
					AvatarURL:  user.GetAvatarURL(),
				}
				minimalUsers = append(minimalUsers, mu)
			}
		}
		minimalResp := &MinimalSearchUsersResult{
			TotalCount:        result.GetTotal(),
			IncompleteResults: result.GetIncompleteResults(),
			Items:             minimalUsers,
		}
		if result.Total != nil {
			minimalResp.TotalCount = *result.Total
		}
		if result.IncompleteResults != nil {
			minimalResp.IncompleteResults = *result.IncompleteResults
		}

		r, err := json.Marshal(minimalResp)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal response: %w", err)
		}
		return utils.NewToolResultText(string(r)), nil
	}
}

// SearchUsers creates a tool to search for GitHub users.
func SearchUsers(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
		Name:        "search_users",
		Description: t("TOOL_SEARCH_USERS_DESCRIPTION", "Search for GitHub users exclusively"),
		Annotations: &mcp.ToolAnnotations{
			Title:        t("TOOL_SEARCH_USERS_USER_TITLE", "Search users"),
			ReadOnlyHint: true,
		},
		InputSchema: WithPagination(&jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"query": {
					Type:        "string",
					Description: "Search query using GitHub users search syntax scoped to type:user",
				},
				"sort": {
					Type:        "string",
					Description: "Sort field by category",
					Enum:        []any{"followers", "repositories", "joined"},
				},
				"order": {
					Type:        "string",
					Description: "Sort order",
					Enum:        []any{"asc", "desc"},
				},
			},
			Required: []string{"query"},
		}),
	}, userOrOrgHandler("user", getClient)
}

// SearchOrgs creates a tool to search for GitHub organizations.
func SearchOrgs(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
		Name:        "search_orgs",
		Description: t("TOOL_SEARCH_ORGS_DESCRIPTION", "Search for GitHub organizations exclusively"),
		Annotations: &mcp.ToolAnnotations{
			Title:        t("TOOL_SEARCH_ORGS_USER_TITLE", "Search organizations"),
			ReadOnlyHint: true,
		},
		InputSchema: WithPagination(&jsonschema.Schema{
			Type: "object",
			Properties: map[string]*jsonschema.Schema{
				"query": {
					Type:        "string",
					Description: "Search query using GitHub organizations search syntax scoped to type:org",
				},
				"sort": {
					Type:        "string",
					Description: "Sort field by category",
					Enum:        []any{"followers", "repositories", "joined"},
				},
				"order": {
					Type:        "string",
					Description: "Sort order",
					Enum:        []any{"asc", "desc"},
				},
			},
			Required: []string{"query"},
		}),
	}, userOrOrgHandler("org", getClient)
}
