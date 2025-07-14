package github

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	ghErrors "github.com/github/github-mcp-server/pkg/errors"
	"github.com/github/github-mcp-server/pkg/translations"
	"github.com/github/github-mcp-server/pkg/utils"
	"github.com/google/go-github/v72/github"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func GetDependabotAlert(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_dependabot_alert",
			Description: t("TOOL_GET_DEPENDABOT_ALERT_DESCRIPTION", "Get details of a specific dependabot alert in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_DEPENDABOT_ALERT_USER_TITLE", "Get dependabot alert"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo", "alertNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "The owner of the repository.",
					},
					"repo": {
						Type:        "string",
						Description: "The name of the repository.",
					},
					"alertNumber": {
						Type:        "number",
						Description: "The number of the alert.",
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
			alertNumber, err := RequiredInt(request, "alertNumber")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			alert, resp, err := client.Dependabot.GetRepoAlert(ctx, owner, repo, alertNumber)
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					fmt.Sprintf("failed to get alert with number '%d'", alertNumber),
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
				return utils.NewToolResultError(fmt.Sprintf("failed to get alert: %s", string(body))), nil
			}

			r, err := json.Marshal(alert)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal alert: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}

func ListDependabotAlerts(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "list_dependabot_alerts",
			Description: t("TOOL_LIST_DEPENDABOT_ALERTS_DESCRIPTION", "List dependabot alerts in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_LIST_DEPENDABOT_ALERTS_USER_TITLE", "List dependabot alerts"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Type:     "object",
				Required: []string{"owner", "repo"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: "The owner of the repository.",
					},
					"repo": {
						Type:        "string",
						Description: "The name of the repository.",
					},
					"state": {
						Type:        "string",
						Description: "Filter dependabot alerts by state. Defaults to open",
						Default:     json.RawMessage(`"open"`),
						Enum:        []any{"open", "fixed", "dismissed", "auto_dismissed"},
					},
					"severity": {
						Type:        "string",
						Description: "Filter dependabot alerts by severity",
						Enum:        []any{"low", "medium", "high", "critical"},
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
			severity, err := OptionalParam[string](request, "severity")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}

			alerts, resp, err := client.Dependabot.ListRepoAlerts(ctx, owner, repo, &github.ListAlertsOptions{
				State:    ToStringPtr(state),
				Severity: ToStringPtr(severity),
			})
			if err != nil {
				return ghErrors.NewGitHubAPIErrorResponse(ctx,
					fmt.Sprintf("failed to list alerts for repository '%s/%s'", owner, repo),
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
				return utils.NewToolResultError(fmt.Sprintf("failed to list alerts: %s", string(body))), nil
			}

			r, err := json.Marshal(alerts)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal alerts: %w", err)
			}

			return utils.NewToolResultText(string(r)), nil
		}
}
