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

func GetSecretScanningAlert(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "get_secret_scanning_alert",
			Description: t("TOOL_GET_SECRET_SCANNING_ALERT_DESCRIPTION", "Get details of a specific secret scanning alert in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_SECRET_SCANNING_ALERT_USER_TITLE", "Get secret scanning alert"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Required: []string{"owner", "repo", "alertNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_GET_SECRET_SCANNING_ALERT_OWNER_DESC", "The owner of the repository."),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_GET_SECRET_SCANNING_ALERT_REPO_DESC", "The name of the repository."),
					},
					"alertNumber": {
						Type:        "number",
						Description: t("TOOL_GET_SECRET_SCANNING_ALERT_NUMBER_DESC", "The number of the alert."),
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

			alert, resp, err := client.SecretScanning.GetAlert(ctx, owner, repo, int64(alertNumber))
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

func ListSecretScanningAlerts(getClient GetClientFn, t translations.TranslationHelperFunc) (tool *mcp.Tool, handler mcp.ToolHandler) {
	return &mcp.Tool{
			Name:        "list_secret_scanning_alerts",
			Description: t("TOOL_LIST_SECRET_SCANNING_ALERTS_DESCRIPTION", "List secret scanning alerts in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_LIST_SECRET_SCANNING_ALERTS_USER_TITLE", "List secret scanning alerts"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Required: []string{"owner", "repo"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_LIST_SECRET_SCANNING_ALERTS_OWNER_DESC", "The owner of the repository."),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_LIST_SECRET_SCANNING_ALERTS_REPO_DESC", "The name of the repository."),
					},
					"state": {
						Type:        "string",
						Description: t("TOOL_LIST_SECRET_SCANNING_ALERTS_STATE_DESC", "Filter by state"),
						Enum:        []any{"open", "resolved"},
					},
					"secret_type": {
						Type:        "string",
						Description: t("TOOL_LIST_SECRET_SCANNING_ALERTS_SECRET_TYPE_DESC", "A comma-separated list of secret types to return. All default secret patterns are returned. To return generic patterns, pass the token name(s) in the parameter."),
					},
					"resolution": {
						Type:        "string",
						Description: t("TOOL_LIST_SECRET_SCANNING_ALERTS_RESOLUTION_DESC", "Filter by resolution"),
						Enum:        []any{"false_positive", "wont_fix", "revoked", "pattern_edited", "pattern_deleted", "used_in_tests"},
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
			secretType, err := OptionalParam[string](request, "secret_type")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}
			resolution, err := OptionalParam[string](request, "resolution")
			if err != nil {
				return utils.NewToolResultError(err.Error()), nil
			}

			client, err := getClient(ctx)
			if err != nil {
				return nil, fmt.Errorf("failed to get GitHub client: %w", err)
			}
			alerts, resp, err := client.SecretScanning.ListAlertsForRepo(ctx, owner, repo, &github.SecretScanningAlertListOptions{State: state, SecretType: secretType, Resolution: resolution})
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
