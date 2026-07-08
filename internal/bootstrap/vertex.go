package bootstrap

import (
	"context"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
	"github.com/voocel/agentcore"
)

// VertexModel implements agentcore.ChatModel for Google Cloud Vertex AI.
// Note for future upgrades: 
// 1. To authenticate, you can either run `gcloud auth application-default login` on your machine.
// 2. Or, create a Service Account in GCP, download the JSON key, and set the environment variable:
//    export GOOGLE_APPLICATION_CREDENTIALS="/path/to/your/service-account-file.json"
// The Vertex AI client will automatically pick up the credentials from either method.
type VertexModel struct {
	client    *genai.Client
	modelName string
	projectID string
	location  string
}

func NewVertexModel(ctx context.Context, projectID, location, modelName string) (*VertexModel, error) {
	// If projectID or location are empty, you could also fall back to env vars
	if projectID == "" {
		projectID = os.Getenv("VERTEX_PROJECT_ID")
	}
	if location == "" {
		location = os.Getenv("VERTEX_LOCATION")
		if location == "" {
			location = "us-central1" // Default fallback
		}
	}

	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return nil, fmt.Errorf("failed to create vertex ai client: %v", err)
	}

	return &VertexModel{
		client:    client,
		modelName: modelName,
		projectID: projectID,
		location:  location,
	}, nil
}

// convertMessages converts agentcore.Message to genai.Content
func (v *VertexModel) convertMessages(messages []agentcore.Message) ([]*genai.Content, []genai.Part) {
	var history []*genai.Content
	var lastParts []genai.Part

	for i, msg := range messages {
		var parts []genai.Part
		for _, block := range msg.Content {
			if block.Type == agentcore.ContentText {
				parts = append(parts, genai.Text(block.Text))
			}
			// Add support for images or tool calls here if needed in the future
		}

		if len(parts) == 0 {
			continue // Skip empty messages
		}

		if i == len(messages)-1 {
			lastParts = parts
			break
		}

		role := string(msg.Role)
		if role == string(agentcore.RoleAssistant) {
			role = "model"
		} else if role == string(agentcore.RoleSystem) {
			role = "user" // Vertex doesn't strictly have a 'system' role in history, fallback to user or use SystemInstruction. For now, user.
		}

		history = append(history, &genai.Content{
			Role:  role,
			Parts: parts,
		})
	}
	return history, lastParts
}

func (v *VertexModel) Generate(ctx context.Context, messages []agentcore.Message, tools []agentcore.ToolSpec, opts ...agentcore.CallOption) (*agentcore.LLMResponse, error) {
	model := v.client.GenerativeModel(v.modelName)
	
	// Convert messages
	history, lastParts := v.convertMessages(messages)

	// Call Vertex AI using ChatSession
	session := model.StartChat()
	session.History = history
	resp, err := session.SendMessage(ctx, lastParts...)
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("vertex ai returned no candidates")
	}

	// Extract text response
	var responseText string
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseText += string(txt)
		}
	}

	// Simple mapping back to agentcore.LLMResponse
	return &agentcore.LLMResponse{
		Message: agentcore.Message{
			Role: agentcore.RoleAssistant,
			Content: []agentcore.ContentBlock{
				agentcore.TextBlock(responseText),
			},
		},
	}, nil
}

func (v *VertexModel) GenerateStream(ctx context.Context, messages []agentcore.Message, tools []agentcore.ToolSpec, opts ...agentcore.CallOption) (<-chan agentcore.StreamEvent, error) {
	// For basic fallback, you can leave streaming unimplemented or implement later
	return nil, fmt.Errorf("GenerateStream is not yet implemented for VertexModel")
}

func (v *VertexModel) SupportsTools() bool {
	return false // Set to true once tool support is implemented
}

func (v *VertexModel) Close() {
	if v.client != nil {
		v.client.Close()
	}
}
