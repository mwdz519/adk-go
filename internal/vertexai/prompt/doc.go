// Copyright 2025 The Go A2A Authors
// SPDX-License-Identifier: Apache-2.0

/*
Package prompts provides comprehensive prompt management functionality for Vertex AI.

This package ports the functionality from Python's vertexai.prompts module to Go,
enabling creation, management, versioning, and deployment of prompt templates
for use with Vertex AI generative models.

# Core Features

The prompts package provides:

1. **Prompt Management**: Create, save, load, delete, and list prompts with cloud storage
2. **Version Control**: Full version history tracking with restore capabilities
3. **Template System**: Advanced template engine with variable substitution
4. **Cloud Integration**: Seamless integration with Vertex AI online prompt resources
5. **Batch Operations**: Efficient bulk operations for managing multiple prompts

# Basic Usage

	// Initialize the prompts service
	ctx := context.Background()
	service, err := prompts.NewService(ctx, "my-project", "us-central1")
	if err != nil {
		log.Fatal(err)
	}
	defer service.Close()

	// Create a prompt with template variables
	prompt := &prompts.Prompt{
		Name:        "greeting-template",
		DisplayName: "Customer Greeting Template",
		Description: "Template for greeting customers",
		Template:    "Hello {customer_name}, welcome to {company_name}!",
		Variables:   []string{"customer_name", "company_name"},
	}

	// Save the prompt to cloud storage
	savedPrompt, err := service.CreatePrompt(ctx, prompt)
	if err != nil {
		log.Fatal(err)
	}

	// Apply template variables
	content, err := service.ApplyTemplate(ctx, savedPrompt.ID, map[string]string{
		"customer_name": "Alice",
		"company_name":  "Acme Corp",
	})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(content) // "Hello Alice, welcome to Acme Corp!"

# Version Management

	// Create a new version of an existing prompt
	updatedPrompt := savedPrompt
	updatedPrompt.Template = "Hi {customer_name}, thanks for choosing {company_name}!"

	newVersion, err := service.CreateVersion(ctx, updatedPrompt)
	if err != nil {
		log.Fatal(err)
	}

	// List all versions
	versions, err := service.ListVersions(ctx, savedPrompt.ID)
	if err != nil {
		log.Fatal(err)
	}

	// Restore a previous version
	err = service.RestoreVersion(ctx, savedPrompt.ID, "1")
	if err != nil {
		log.Fatal(err)
	}

# Advanced Templates

	// Multi-modal prompt with images and complex variables
	prompt := &prompts.Prompt{
		Name:        "image-analysis-template",
		DisplayName: "Image Analysis Template",
		Template: `Analyze this image: {image_url}

Context: {context}
Task: {task}
Format: {output_format}

Please provide analysis in the specified format.`,

		Variables: []string{"image_url", "context", "task", "output_format"},
		Category:  "multimodal",
		Tags:      []string{"image", "analysis", "multimodal"},
	}

# Batch Operations

	// Import multiple prompts from a configuration
	prompts := []*prompts.Prompt{
		{Name: "template1", Template: "Hello {name}"},
		{Name: "template2", Template: "Goodbye {name}"},
	}

	results, err := service.BatchCreatePrompts(ctx, prompts)
	if err != nil {
		log.Fatal(err)
	}

	// Export prompts for backup or migration
	exported, err := service.ExportPrompts(ctx, []string{"template1", "template2"})
	if err != nil {
		log.Fatal(err)
	}

# Search and Filtering

	// Search prompts by various criteria
	results, err := service.SearchPrompts(ctx, &prompts.SearchOptions{
		Query:    "greeting",
		Category: "customer-service",
		Tags:     []string{"personalization"},
		PageSize: 20,
	})
	if err != nil {
		log.Fatal(err)
	}

# Integration with Generative Models

The prompts package integrates seamlessly with Vertex AI generative models:

	// Use prompt with a generative model
	prompt, err := service.GetPrompt(ctx, "greeting-template")
	if err != nil {
		log.Fatal(err)
	}

	// Apply variables and generate content
	content, err := prompt.ApplyVariables(map[string]string{
		"customer_name": "Bob",
		"company_name":  "Tech Solutions",
	})
	if err != nil {
		log.Fatal(err)
	}

	// Use with generative model (pseudo-code - actual implementation depends on model integration)
	response, err := model.GenerateContent(ctx, content)

# Error Handling

The package provides comprehensive error handling with specific error types:

	if err != nil {
		switch {
		case prompts.IsNotFound(err):
			log.Printf("Prompt not found: %v", err)
		case prompts.IsVersionConflict(err):
			log.Printf("Version conflict: %v", err)
		case prompts.IsInvalidTemplate(err):
			log.Printf("Invalid template: %v", err)
		default:
			log.Printf("Unexpected error: %v", err)
		}
	}

# Best Practices

1. **Template Design**: Use clear, descriptive variable names and provide default values where appropriate
2. **Version Management**: Create versions for significant changes and tag them with meaningful descriptions
3. **Organization**: Use categories and tags to organize prompts for easy discovery
4. **Testing**: Validate templates with sample data before deploying to production
5. **Security**: Avoid embedding sensitive information directly in templates; use variables instead

# Performance Considerations

- The service caches frequently accessed prompts to reduce API calls
- Batch operations are more efficient than individual operations for bulk changes
- Template compilation is optimized for repeated variable substitution
- Version history is stored efficiently with incremental diffs

# Compatibility

This package maintains compatibility with:
- Python's vertexai.prompts module functionality
- Vertex AI Studio prompt management
- Google Cloud Console prompt resources
- Standard Go concurrency patterns
*/
package prompt
