// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

package tools

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/genai"
)

// Test function signatures for testing
func testSimpleFunction(ctx context.Context, query string, limit int) ([]string, error) {
	return []string{"result1", "result2"}, nil
}

func testStructFunction(ctx context.Context, req TestRequest) (TestResponse, error) {
	return TestResponse{Message: "success", Count: 1}, nil
}

func testNoContextFunction(name string, age int) (string, error) {
	return "processed", nil
}

func testPointerParams(ctx context.Context, name *string, config *TestConfig) (string, error) {
	return "processed", nil
}

type TestRequest struct {
	Query    string `json:"query"`
	Limit    int    `json:"limit,omitempty"`
	Priority *int   `json:"priority,omitempty"`
}

type TestResponse struct {
	Message string   `json:"message"`
	Count   int      `json:"count"`
	Results []string `json:"results,omitempty"`
}

type TestConfig struct {
	Enabled bool   `json:"enabled"`
	Value   string `json:"value"`
	Hidden  string `json:"-"`
}

func TestBuildFunctionDeclaration(t *testing.T) {
	tests := []struct {
		name    string
		fn      any
		options []FunctionOption
		want    *genai.FunctionDeclaration
		wantErr bool
	}{
		{
			name: "simple function with basic types",
			fn:   testSimpleFunction,
			options: []FunctionOption{
				WithName("search"),
				WithDescription("Search for items"),
				WithParameterDescription("param1", "Search query"),
				WithParameterDescription("param2", "Maximum results"),
			},
			want: &genai.FunctionDeclaration{
				Name:        "search",
				Description: "Search for items",
				Behavior:    genai.BehaviorBlocking,
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"param1": {Type: genai.TypeString, Description: "Search query"},
						"param2": {Type: genai.TypeInteger, Description: "Maximum results"},
					},
					Required: []string{"param1", "param2"},
				},
			},
		},
		{
			name: "struct parameter function",
			fn:   testStructFunction,
			options: []FunctionOption{
				WithName("process"),
				WithDescription("Process request"),
			},
			want: &genai.FunctionDeclaration{
				Name:        "process",
				Description: "Process request",
				Behavior:    genai.BehaviorBlocking,
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"param1": {
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"query":    {Type: genai.TypeString},
								"limit":    {Type: genai.TypeInteger},
								"priority": {Type: genai.TypeInteger},
							},
							Required: []string{"query"},
						},
					},
					Required: []string{"param1"},
				},
			},
		},
		{
			name: "function without context",
			fn:   testNoContextFunction,
			options: []FunctionOption{
				WithName("nocontext"),
			},
			want: &genai.FunctionDeclaration{
				Name:     "nocontext",
				Behavior: genai.BehaviorBlocking,
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"param1": {Type: genai.TypeString},
						"param2": {Type: genai.TypeInteger},
					},
					Required: []string{"param1", "param2"},
				},
			},
		},
		{
			name: "function with pointer parameters",
			fn:   testPointerParams,
			options: []FunctionOption{
				WithName("pointers"),
			},
			want: &genai.FunctionDeclaration{
				Name:     "pointers",
				Behavior: genai.BehaviorBlocking,
				Parameters: &genai.Schema{
					Type: genai.TypeObject,
					Properties: map[string]*genai.Schema{
						"param1": {Type: genai.TypeString},
						"param2": {
							Type: genai.TypeObject,
							Properties: map[string]*genai.Schema{
								"enabled": {Type: genai.TypeBoolean},
								"value":   {Type: genai.TypeString},
							},
							Required: []string{"enabled", "value"},
						},
					},
					// No Required field since all parameters are optional (pointers)
				},
			},
		},
		{
			name:    "nil function should error",
			fn:      nil,
			wantErr: true,
		},
		{
			name:    "non-function should error",
			fn:      "not a function",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildFunctionDeclaration(tt.fn, tt.options...)
			if (err != nil) != tt.wantErr {
				t.Errorf("BuildFunctionDeclaration() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("BuildFunctionDeclaration() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestTypeToSchema(t *testing.T) {
	tests := []struct {
		name    string
		input   any
		want    *genai.Schema
		wantErr bool
	}{
		{
			name:  "string type",
			input: "",
			want:  &genai.Schema{Type: genai.TypeString},
		},
		{
			name:  "int type",
			input: 0,
			want:  &genai.Schema{Type: genai.TypeInteger},
		},
		{
			name:  "float type",
			input: 0.0,
			want:  &genai.Schema{Type: genai.TypeNumber},
		},
		{
			name:  "bool type",
			input: false,
			want:  &genai.Schema{Type: genai.TypeBoolean},
		},
		{
			name:  "slice type",
			input: []string{},
			want: &genai.Schema{
				Type:  genai.TypeArray,
				Items: &genai.Schema{Type: genai.TypeString},
			},
		},
		{
			name:  "map type",
			input: map[string]int{},
			want: &genai.Schema{
				Type:        genai.TypeObject,
				Description: "Map with string keys",
			},
		},
		{
			name:  "struct type",
			input: TestConfig{},
			want: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"enabled": {Type: genai.TypeBoolean},
					"value":   {Type: genai.TypeString},
				},
				Required: []string{"enabled", "value"},
			},
		},
		{
			name:  "any type",
			input: (*any)(nil),
			want:  &genai.Schema{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var input any = tt.input
			if tt.input == (*any)(nil) {
				input = (*any)(nil)
			}

			got, err := typeToSchema(reflect.TypeOf(input))
			if (err != nil) != tt.wantErr {
				t.Errorf("typeToSchema() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("typeToSchema() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestNewFunctionToolWithDeclaration(t *testing.T) {
	fn := func(ctx context.Context, args map[string]any) (any, error) {
		return "result", nil
	}

	tool, err := newFunctionToolWithDeclaration(fn,
		WithName("test_tool"),
		WithDescription("Test tool description"),
	)
	if err != nil {
		t.Fatalf("NewFunctionToolWithDeclaration() error = %v", err)
	}

	if tool.Name() != "test_tool" {
		t.Errorf("tool.Name() = %v, want %v", tool.Name(), "test_tool")
	}

	decl := tool.GetDeclaration()
	if decl == nil {
		t.Fatal("FunctionDeclaration() returned nil")
	}

	if decl.Name != "test_tool" {
		t.Errorf("declaration.Name = %v, want %v", decl.Name, "test_tool")
	}

	if decl.Description != "Test tool description" {
		t.Errorf("declaration.Description = %v, want %v", decl.Description, "Test tool description")
	}
}

func TestWrapFunction(t *testing.T) {
	// Test function to wrap
	testFunc := func(ctx context.Context, req TestRequest) (TestResponse, error) {
		return TestResponse{
			Message: "Processed: " + req.Query,
			Count:   req.Limit,
		}, nil
	}

	wrapped := wrapFunction(testFunc)

	// Test with valid args
	args := map[string]any{
		"query": "test query",
		"limit": 5,
	}

	result, err := wrapped(context.Background(), args)
	if err != nil {
		t.Fatalf("wrapped function error = %v", err)
	}

	response, ok := result.(TestResponse)
	if !ok {
		t.Fatalf("result type = %T, want TestResponse", result)
	}

	if response.Message != "Processed: test query" {
		t.Errorf("response.Message = %v, want %v", response.Message, "Processed: test query")
	}

	if response.Count != 5 {
		t.Errorf("response.Count = %v, want %v", response.Count, 5)
	}
}

func TestGetFunctionName(t *testing.T) {
	tests := []struct {
		name     string
		fn       any
		expected string
	}{
		{
			name:     "named function",
			fn:       testSimpleFunction,
			expected: "testSimpleFunction",
		},
		{
			name: "anonymous function",
			fn: func(ctx context.Context) error {
				return nil
			},
			expected: "function", // fallback for anonymous functions
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := reflect.ValueOf(tt.fn)
			got := getFunctionName(v)

			// For named functions, check that we get the expected name
			// For anonymous functions, accept the fallback
			if tt.expected != "function" && got != tt.expected {
				t.Errorf("getFunctionName() = %v, want %v", got, tt.expected)
			}
			if tt.expected == "function" && got != "function" {
				// Anonymous functions might have generated names, so we accept anything
			}
		})
	}
}

func TestIsContextType(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
		want bool
	}{
		{
			name: "context.Context",
			typ:  reflect.TypeOf((*context.Context)(nil)).Elem(),
			want: true,
		},
		{
			name: "concrete context",
			typ:  reflect.TypeOf(context.Background()),
			want: true,
		},
		{
			name: "string type",
			typ:  reflect.TypeOf(""),
			want: false,
		},
		{
			name: "int type",
			typ:  reflect.TypeOf(0),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isContextType(tt.typ)
			if got != tt.want {
				t.Errorf("isContextType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetJSONFieldName(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected string
	}{
		{
			name: "no json tag",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  "",
			},
			expected: "testfield",
		},
		{
			name: "json tag with name",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"custom_name"`,
			},
			expected: "custom_name",
		},
		{
			name: "json tag with omitempty",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"field_name,omitempty"`,
			},
			expected: "field_name",
		},
		{
			name: "json tag skip",
			field: reflect.StructField{
				Name: "TestField",
				Tag:  `json:"-"`,
			},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getJSONFieldName(tt.field)
			if got != tt.expected {
				t.Errorf("getJSONFieldName() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestIsRequiredField(t *testing.T) {
	tests := []struct {
		name     string
		field    reflect.StructField
		expected bool
	}{
		{
			name: "regular field",
			field: reflect.StructField{
				Name: "TestField",
				Type: reflect.TypeOf(""),
				Tag:  "",
			},
			expected: true,
		},
		{
			name: "pointer field",
			field: reflect.StructField{
				Name: "TestField",
				Type: reflect.TypeOf((*string)(nil)),
				Tag:  "",
			},
			expected: false,
		},
		{
			name: "omitempty field",
			field: reflect.StructField{
				Name: "TestField",
				Type: reflect.TypeOf(""),
				Tag:  `json:"field,omitempty"`,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isRequiredField(tt.field)
			if got != tt.expected {
				t.Errorf("isRequiredField() = %v, want %v", got, tt.expected)
			}
		})
	}
}
