package github

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/github/github-mcp-server/pkg/utils"
	"github.com/google/go-github/v72/github"
	"github.com/modelcontextprotocol/go-sdk/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewServer creates a new GitHub MCP server with the specified GH client and logger.

func NewServer(version string, opts *mcp.ServerOptions) *mcp.Server {
	// Add default options
	// defaultOpts := []server.ServerOption{
	// 	server.WithToolCapabilities(true),
	// 	server.WithResourceCapabilities(true, true),
	// 	server.WithLogging(),
	// }
	// opts = append(defaultOpts, opts...)

	// Create a new MCP server
	s := mcp.NewServer(
		&mcp.Implementation{
			Name:    "github-mcp",
			Title:   "GitHub MCP Server",
			Version: version,
		},
		opts,
	)
	return s
}

// OptionalParamOK is a helper function that can be used to fetch a requested parameter from the request.
// It returns the value, a boolean indicating if the parameter was present, and an error if the type is wrong.
func OptionalParamOK[T any](params *mcp.CallToolParamsFor[map[string]any], p string) (value T, ok bool, err error) {
	// Check if the parameter is present in the request
	val, exists := params.Arguments[p]
	if !exists {
		// Not present, return zero value, false, no error
		return
	}

	// Check if the parameter is of the expected type
	value, ok = val.(T)
	if !ok {
		// Present but wrong type
		err = fmt.Errorf("parameter %s is not of type %T, is %T", p, value, val)
		ok = true // Set ok to true because the parameter *was* present, even if wrong type
		return
	}

	// Present and correct type
	ok = true
	return
}

// isAcceptedError checks if the error is an accepted error.
func isAcceptedError(err error) bool {
	var acceptedError *github.AcceptedError
	return errors.As(err, &acceptedError)
}

// RequiredParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request.
// 2. Checks if the parameter is of the expected type.
// 3. Checks if the parameter is not empty, i.e: non-zero value
func RequiredParam[T comparable](params *mcp.CallToolParamsFor[map[string]any], p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := params.Arguments[p]; !ok {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	// Check if the parameter is of the expected type
	val, ok := params.Arguments[p].(T)
	if !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T", p, zero)
	}

	if val == zero {
		return zero, fmt.Errorf("missing required parameter: %s", p)
	}

	return val, nil
}

// RequiredInt is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request.
// 2. Checks if the parameter is of the expected type.
// 3. Checks if the parameter is not empty, i.e: non-zero value
func RequiredInt(params *mcp.CallToolParamsFor[map[string]any], p string) (int, error) {
	v, err := RequiredParam[float64](params, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// OptionalParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, it checks if the parameter is of the expected type and returns it
func OptionalParam[T any](params *mcp.CallToolParamsFor[map[string]any], p string) (T, error) {
	var zero T

	// Check if the parameter is present in the request
	if _, ok := params.Arguments[p]; !ok {
		return zero, nil
	}

	// Check if the parameter is of the expected type
	if _, ok := params.Arguments[p].(T); !ok {
		return zero, fmt.Errorf("parameter %s is not of type %T, is %T", p, zero, params.Arguments[p])
	}

	return params.Arguments[p].(T), nil
}

// OptionalIntParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, it checks if the parameter is of the expected type and returns it
func OptionalIntParam(params *mcp.CallToolParamsFor[map[string]any], p string) (int, error) {
	v, err := OptionalParam[float64](params, p)
	if err != nil {
		return 0, err
	}
	return int(v), nil
}

// OptionalIntParamWithDefault is a helper function that can be used to fetch a requested parameter from the request
// similar to optionalIntParam, but it also takes a default value.
func OptionalIntParamWithDefault(params *mcp.CallToolParamsFor[map[string]any], p string, d int) (int, error) {
	v, err := OptionalIntParam(params, p)
	if err != nil {
		return 0, err
	}
	if v == 0 {
		return d, nil
	}
	return v, nil
}

// OptionalStringArrayParam is a helper function that can be used to fetch a requested parameter from the request.
// It does the following checks:
// 1. Checks if the parameter is present in the request, if not, it returns its zero-value
// 2. If it is present, iterates the elements and checks each is a string
func OptionalStringArrayParam(params *mcp.CallToolParamsFor[map[string]any], p string) ([]string, error) {
	// Check if the parameter is present in the request
	if _, ok := params.Arguments[p]; !ok {
		return []string{}, nil
	}

	switch v := params.Arguments[p].(type) {
	case nil:
		return []string{}, nil
	case []string:
		return v, nil
	case []any:
		strSlice := make([]string, len(v))
		for i, v := range v {
			s, ok := v.(string)
			if !ok {
				return []string{}, fmt.Errorf("parameter %s is not of type string, is %T", p, v)
			}
			strSlice[i] = s
		}
		return strSlice, nil
	default:
		return []string{}, fmt.Errorf("parameter %s could not be coerced to []string, is %T", p, params.Arguments[p])
	}
}

// WithPagination returns a ToolOption that adds "page" and "perPage" parameters to the tool.
// The "page" parameter is optional, min 1.
// The "perPage" parameter is optional, min 1, max 100. If unset, defaults to 30.
// https://docs.github.com/en/rest/using-the-rest-api/using-pagination-in-the-rest-api
func WithPagination(schema *jsonschema.Schema) *jsonschema.Schema {
	schema.Properties["page"] = &jsonschema.Schema{
		Type:        "Number",
		Description: "Page number for pagination (min 1)",
		Minimum:     jsonschema.Ptr(1.0),
	}

	schema.Properties["perPage"] = &jsonschema.Schema{
		Type:        "Number",
		Description: "Results per page for pagination (min 1, max 100)",
		Minimum:     jsonschema.Ptr(1.0),
		Maximum:     jsonschema.Ptr(100.0),
	}

	return schema
}

type PaginationParams struct {
	page    int
	perPage int
}

// OptionalPaginationParams returns the "page" and "perPage" parameters from the request,
// or their default values if not present, "page" default is 1, "perPage" default is 30.
// In future, we may want to make the default values configurable, or even have this
// function returned from `withPagination`, where the defaults are provided alongside
// the min/max values.
func OptionalPaginationParams(params *mcp.CallToolParamsFor[map[string]any]) (PaginationParams, error) {
	page, err := OptionalIntParamWithDefault(params, "page", 1)
	if err != nil {
		return PaginationParams{}, err
	}
	perPage, err := OptionalIntParamWithDefault(params, "perPage", 30)
	if err != nil {
		return PaginationParams{}, err
	}
	return PaginationParams{
		page:    page,
		perPage: perPage,
	}, nil
}

func MarshalledTextResult(v any) *mcp.CallToolResult {
	data, err := json.Marshal(v)
	if err != nil {
		return utils.NewToolResultErrorFromErr("failed to marshal text result to json", err)
	}

	return utils.NewToolResultText(string(data))
}
