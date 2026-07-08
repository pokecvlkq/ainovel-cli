package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"cloud.google.com/go/vertexai/genai"
	"github.com/voocel/agentcore"
	"google.golang.org/api/option"
)

// VertexModel triển khai agentcore.ChatModel cho Google Cloud Vertex AI.
//
// === HƯỚNG DẪN NÂNG CẤP (Vertex AI Provider) ===
// 1. Xác thực: Dùng Service Account JSON, truyền qua:
//    - config.json: "api_key": "$VERTEX_ACC_1"  (đọc JSON từ biến môi trường .env)
//    - config.json: "api_key": "credentials/aitnd.json"  (đọc file trực tiếp)
// 2. project_id được TỰ ĐỘNG parse từ JSON credential, không cần cấu hình thêm.
// 3. location mặc định "us-central1", có thể ghi đè bằng env VERTEX_LOCATION.
// 4. Hỗ trợ multi-account fallback: mỗi provider Vertex dùng 1 Service Account riêng.
// ================================================
type VertexModel struct {
	client    *genai.Client
	modelName string
	projectID string
	location  string
}

// resolveCredential xử lý giá trị api_key linh hoạt:
//   - Bắt đầu bằng "$": đọc nội dung từ biến môi trường (ví dụ "$VERTEX_ACC_1")
//   - Bắt đầu bằng "{": coi là chuỗi JSON credential trực tiếp
//   - Còn lại: coi là đường dẫn file JSON
//
// Trả về: (jsonBytes, isJSON, error)
func resolveCredential(raw string) ([]byte, bool, error) {
	if raw == "" {
		return nil, false, nil
	}

	value := raw

	// Bước 1: Nếu bắt đầu bằng $, đọc từ biến môi trường
	if value[0] == '$' {
		envVar := value[1:]
		value = os.Getenv(envVar)
		if value == "" {
			return nil, false, fmt.Errorf("biến môi trường %s không tồn tại hoặc rỗng", envVar)
		}
	}

	// Bước 2: Nếu là chuỗi JSON (bắt đầu bằng {), trả về trực tiếp
	if len(value) > 0 && value[0] == '{' {
		return []byte(value), true, nil
	}

	// Bước 3: Coi là đường dẫn file, đọc nội dung
	data, err := os.ReadFile(value)
	if err != nil {
		return nil, false, fmt.Errorf("không thể đọc file credential %q: %w", value, err)
	}
	return data, true, nil
}

// extractProjectID trích xuất project_id từ JSON credential
func extractProjectID(jsonData []byte) string {
	var sa struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(jsonData, &sa); err == nil && sa.ProjectID != "" {
		return sa.ProjectID
	}
	return ""
}

func NewVertexModel(ctx context.Context, projectID, location, modelName, credentialsFile string) (*VertexModel, error) {
	if location == "" {
		location = os.Getenv("VERTEX_LOCATION")
		if location == "" {
			location = "us-central1"
		}
	}

	var opts []option.ClientOption

	// Xử lý credential và tự động parse project_id
	if credentialsFile != "" {
		jsonData, isJSON, err := resolveCredential(credentialsFile)
		if err != nil {
			return nil, fmt.Errorf("lỗi xử lý credential: %w", err)
		}
		if isJSON && len(jsonData) > 0 {
			opts = append(opts, option.WithCredentialsJSON(jsonData))
			// Tự động lấy project_id từ Service Account JSON nếu chưa có
			if projectID == "" {
				projectID = extractProjectID(jsonData)
			}
		}
	}

	// Fallback cuối cùng cho projectID
	if projectID == "" {
		projectID = os.Getenv("VERTEX_PROJECT_ID")
	}
	if projectID == "" {
		return nil, fmt.Errorf("không xác định được project_id: cần khai báo trong credential JSON hoặc env VERTEX_PROJECT_ID")
	}

	client, err := genai.NewClient(ctx, projectID, location, opts...)
	if err != nil {
		return nil, fmt.Errorf("không thể tạo Vertex AI client: %v", err)
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
