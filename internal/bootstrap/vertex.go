package bootstrap

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync/atomic"

	"cloud.google.com/go/vertexai/genai"
	"github.com/voocel/agentcore"
	"google.golang.org/api/option"
)

// VertexModel triển khai agentcore.ChatModel cho Google Cloud Vertex AI.
//
// === HƯỚNG DẪN NÂNG CẤP (Vertex AI Provider) ===
// 1. Xác thực: Dùng Service Account JSON, truyền qua:
//   - config.json: "api_key": "$VERTEX_ACC_1"  (đọc JSON từ biến môi trường .env)
//   - config.json: "api_key": "credentials/aitnd.json"  (đọc file trực tiếp)
//
// 2. project_id được TỰ ĐỘNG parse từ JSON credential, không cần cấu hình thêm.
// 3. location mặc định "us-central1", có thể ghi đè bằng env VERTEX_LOCATION.
// 4. Hỗ trợ multi-account fallback: mỗi provider Vertex dùng 1 Service Account riêng.
// ================================================
type VertexModel struct {
	client      *genai.Client
	modelName   string
	projectID   string
	location    string
	callCounter int64 // bộ đếm tạo ID duy nhất cho mỗi lần gọi tool
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

// convertSchema converts a generic JSON schema representation to genai.Schema
func convertSchema(v any) *genai.Schema {
	if v == nil {
		return nil
	}
	var m map[string]any
	switch val := v.(type) {
	case map[string]any:
		m = val
	case []byte:
		_ = json.Unmarshal(val, &m)
	case string:
		_ = json.Unmarshal([]byte(val), &m)
	default:
		b, _ := json.Marshal(v)
		_ = json.Unmarshal(b, &m)
	}

	if m == nil {
		return nil
	}

	schema := &genai.Schema{}
	if t, ok := m["type"].(string); ok {
		switch t {
		case "string":
			schema.Type = genai.TypeString
		case "number":
			schema.Type = genai.TypeNumber
		case "integer":
			schema.Type = genai.TypeInteger
		case "boolean":
			schema.Type = genai.TypeBoolean
		case "array":
			schema.Type = genai.TypeArray
		case "object":
			schema.Type = genai.TypeObject
		}
	}
	if d, ok := m["description"].(string); ok {
		schema.Description = d
	}
	if props, ok := m["properties"].(map[string]any); ok {
		schema.Properties = make(map[string]*genai.Schema)
		for k, p := range props {
			schema.Properties[k] = convertSchema(p)
		}
	}
	if reqs, ok := m["required"].([]any); ok {
		for _, r := range reqs {
			if s, ok := r.(string); ok {
				schema.Required = append(schema.Required, s)
			}
		}
	}
	if items, ok := m["items"]; ok {
		schema.Items = convertSchema(items)
	}
	return schema
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

	// Tự động loại bỏ prefix "google/" và suffix ":free"/":paid" nếu user nhập nhầm
	modelName = strings.TrimPrefix(modelName, "google/")
	if idx := strings.Index(modelName, ":"); idx > 0 {
		modelName = modelName[:idx]
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

	var allContents []*genai.Content
	var currentToolContent *genai.Content

	for _, msg := range messages {
		var parts []genai.Part

		role := string(msg.Role)
		if role == string(agentcore.RoleAssistant) {
			role = "model"
		} else if role == string(agentcore.RoleSystem) || role == string(agentcore.RoleTool) {
			role = "user"
		}

		if msg.Role == agentcore.RoleTool {
			toolCallID, _ := msg.Metadata["tool_call_id"].(string)
			// Tách phần _N (counter) ra để lấy lại tên hàm gốc cho Vertex AI
			toolName := toolCallID
			if idx := strings.LastIndex(toolCallID, "_"); idx > 0 {
				suffix := toolCallID[idx+1:]
				if _, err := strconv.Atoi(suffix); err == nil {
					toolName = toolCallID[:idx]
				}
			}
			if len(msg.Content) > 0 {
				var respMap map[string]any
				if err := json.Unmarshal([]byte(msg.Content[0].Text), &respMap); err == nil {
					parts = append(parts, genai.FunctionResponse{
						Name:     toolName,
						Response: respMap,
					})
				} else {
					parts = append(parts, genai.FunctionResponse{
						Name:     toolName,
						Response: map[string]any{"result": msg.Content[0].Text},
					})
				}
			}

			if len(parts) > 0 {
				if currentToolContent != nil {
					currentToolContent.Parts = append(currentToolContent.Parts, parts...)
				} else {
					currentToolContent = &genai.Content{
						Role:  role,
						Parts: parts,
					}
					allContents = append(allContents, currentToolContent)
				}
			}
			continue
		}

		currentToolContent = nil

		for _, block := range msg.Content {
			if block.Type == agentcore.ContentText {
				parts = append(parts, genai.Text(block.Text))
			} else if block.Type == agentcore.ContentToolCall {
				var args map[string]any
				_ = json.Unmarshal(block.ToolCall.Args, &args)
				parts = append(parts, genai.FunctionCall{
					Name: block.ToolCall.Name,
					Args: args,
				})
			}
		}

		if len(parts) > 0 {
			allContents = append(allContents, &genai.Content{
				Role:  role,
				Parts: parts,
			})
		}
	}

	if len(allContents) > 0 {
		lastContent := allContents[len(allContents)-1]
		lastParts = lastContent.Parts
		history = allContents[:len(allContents)-1]
	}

	return history, lastParts
}

func (v *VertexModel) Generate(ctx context.Context, messages []agentcore.Message, tools []agentcore.ToolSpec, opts ...agentcore.CallOption) (*agentcore.LLMResponse, error) {
	model := v.client.GenerativeModel(v.modelName)

	if len(tools) > 0 {
		var decls []*genai.FunctionDeclaration
		for _, t := range tools {
			decls = append(decls, &genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  convertSchema(t.Parameters),
			})
		}
		model.Tools = []*genai.Tool{
			{FunctionDeclarations: decls},
		}
	}

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

	// Extract text response and tool calls
	var responseBlocks []agentcore.ContentBlock
	for _, part := range resp.Candidates[0].Content.Parts {
		if txt, ok := part.(genai.Text); ok {
			responseBlocks = append(responseBlocks, agentcore.TextBlock(string(txt)))
		} else if fc, ok := part.(genai.FunctionCall); ok {
			argsBytes, _ := json.Marshal(fc.Args)
			callID := fc.Name + "_" + strconv.FormatInt(atomic.AddInt64(&v.callCounter, 1), 10)
			responseBlocks = append(responseBlocks, agentcore.ToolCallBlock(agentcore.ToolCall{
				ID:   callID,
				Name: fc.Name,
				Args: argsBytes,
			}))
		}
	}

	// Simple mapping back to agentcore.LLMResponse
	return &agentcore.LLMResponse{
		Message: agentcore.Message{
			Role:    agentcore.RoleAssistant,
			Content: responseBlocks,
		},
	}, nil
}

func (v *VertexModel) GenerateStream(ctx context.Context, messages []agentcore.Message, tools []agentcore.ToolSpec, opts ...agentcore.CallOption) (<-chan agentcore.StreamEvent, error) {
	// For basic fallback, you can leave streaming unimplemented or implement later
	return nil, fmt.Errorf("GenerateStream is not yet implemented for VertexModel")
}

func (v *VertexModel) SupportsTools() bool {
	return true
}

func (v *VertexModel) Close() {
	if v.client != nil {
		v.client.Close()
	}
}
