package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/voocel/ainovel-cli/internal/domain"
)

// ProjectInfo chứa thông tin tóm tắt của một dự án truyện.
type ProjectInfo struct {
	DirName            string
	NovelName          string
	ChapterCount       int
	TotalRealWordCount int
	LastUpdated        time.Time
}

// DiscoverProjects quét thư mục outputDir tìm các dự án hợp lệ (chứa meta/progress.json)
// và trả về danh sách được sắp xếp theo thời gian cập nhật mới nhất.
func DiscoverProjects(outputDir string) ([]ProjectInfo, error) {
	entries, err := os.ReadDir(outputDir)
	if err != nil {
		if os.IsNotExist(err) {
			return []ProjectInfo{}, nil
		}
		return nil, err
	}

	var projects []ProjectInfo

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		projectDir := filepath.Join(outputDir, entry.Name())
		progressPath := filepath.Join(projectDir, "meta", "progress.json")

		info, err := os.Stat(progressPath)
		if err != nil {
			// Bỏ qua thư mục không có meta/progress.json hợp lệ
			continue
		}

		data, err := os.ReadFile(progressPath)
		if err != nil {
			continue
		}

		var progress domain.Progress
		if err := json.Unmarshal(data, &progress); err != nil {
			continue
		}

		projects = append(projects, ProjectInfo{
			DirName:            entry.Name(),
			NovelName:          progress.NovelName,
			ChapterCount:       progress.CurrentChapter,
			TotalRealWordCount: progress.TotalRealWordCount,
			LastUpdated:        info.ModTime(),
		})
	}

	// Kiểm tra xem outputDir bản thân nó có chứa project cũ không (legacy)
	rootProgressPath := filepath.Join(outputDir, "meta", "progress.json")
	if info, err := os.Stat(rootProgressPath); err == nil {
		if data, err := os.ReadFile(rootProgressPath); err == nil {
			var progress domain.Progress
			if err := json.Unmarshal(data, &progress); err == nil {
				projects = append(projects, ProjectInfo{
					DirName:            ".",
					NovelName:          progress.NovelName,
					ChapterCount:       progress.CurrentChapter,
					TotalRealWordCount: progress.TotalRealWordCount,
					LastUpdated:        info.ModTime(),
				})
			}
		}
	}

	// Sắp xếp giảm dần theo LastUpdated (mới nhất xếp trên)
	sort.Slice(projects, func(i, j int) bool {
		return projects[i].LastUpdated.After(projects[j].LastUpdated)
	})

	return projects, nil
}
