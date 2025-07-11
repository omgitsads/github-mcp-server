I would like you to the current file and it's corresponding test file *ONLY*. I need to convert these MCP tool handlers from using mark3labs/mcp-go to using the modelcontextprotocol/go-sdk library. Here's a list of changes that are required for this:

* The import for `github.com/mark3labs/mcp-go/mcp` should be changed to `github.com/modelcontextprotocol/go-sdk/mcp`
* The return signitures for the tool handlers need to change from returning `(tool mcp.Tool, handler server.ToolHandlerFunc)` to returning `(tool *mcp.Tool, handler mcp.ToolHandler)`.
* The function signiture for the tool handler needs to change from `func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)` to `func(ctx context.Context, session *mcp.ServerSession, request *mcp.CallToolParamsFor[map[string]any]) (*mcp.CallToolResult, error)`
* Any calls to `mcp.NewToolResultError` or `mcp.NewToolResultText` should be changed to use the `utils` package, i.e. `utils.NewToolResultText`, `utils.NewToolResultError`

In addition, the tool definition needs to be rewritten. This needs to change from calling the `mcp.NewTool` function, to building a `mcp.Tool` struct directly and returning a pointer to it. For example the following current definition

```
	return mcp.NewTool("get_code_scanning_alert",
			mcp.WithDescription(t("TOOL_GET_CODE_SCANNING_ALERT_DESCRIPTION", "Get details of a specific code scanning alert in a GitHub repository.")),
			mcp.WithToolAnnotation(mcp.ToolAnnotation{
				Title:        t("TOOL_GET_CODE_SCANNING_ALERT_USER_TITLE", "Get code scanning alert"),
				ReadOnlyHint: ToBoolPtr(true),
			}),
			mcp.WithString("owner",
				mcp.Required(),
				mcp.Description("The owner of the repository."),
			),
			mcp.WithString("repo",
				mcp.Required(),
				mcp.Description("The name of the repository."),
			),
			mcp.WithNumber("alertNumber",
				mcp.Required(),
				mcp.Description("The number of the alert."),
			),
		),
    ...
```

This should be rewritten to

```
	return &mcp.Tool{
			Name:        "get_code_scanning_alert",
			Description: t("TOOL_GET_CODE_SCANNING_ALERT_DESCRIPTION", "Get details of a specific code scanning alert in a GitHub repository."),
			Annotations: &mcp.ToolAnnotations{
				Title:        t("TOOL_GET_CODE_SCANNING_ALERT_USER_TITLE", "Get code scanning alert"),
				ReadOnlyHint: true,
			},
			InputSchema: &jsonschema.Schema{
				Required: []string{"owner", "repo", "alertNumber"},
				Properties: map[string]*jsonschema.Schema{
					"owner": {
						Type:        "string",
						Description: t("TOOL_GET_CODE_SCANNING_ALERT_OWNER_DESC", "The owner of the repository."),
					},
					"repo": {
						Type:        "string",
						Description: t("TOOL_GET_CODE_SCANNING_ALERT_REPO_DESC", "The name of the repository."),
					},
					"alertNumber": {
						Type:        "number",
						Description: t("TOOL_GET_CODE_SCANNING_ALERT_NUMBER_DESC", "The number of the alert."),
					},
				},
			},
		},
    ...
```

If a file exists with the same name as the current file, but with a `_test` suffix, this is test file and may require the following changes need to be made:

* The `createMCPRequest` function has changed, it now takes a `context.Context` and returns a `*mcp.ServerSession`. You should create a new `ctx` variable with `context.Background()`
* The `handler` func now takes a `*mcp.ServerSesion`, you should change the method call, and re-use the existing `ctx` for the `context.Context` argument.

If you wish to run the tests, you should run them with `UPDATE_TOOLSNAPS=true` set in the environment, as otherwise the test will fail due to slight changes in the Schema. These failures are expected.