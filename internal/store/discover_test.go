package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/voocel/ainovel-cli/internal/domain"
)

func TestDiscoverProjects(t *testing.T) {
	// Create a temporary output directory
	tempDir := t.TempDir()

	// 1. Create a valid project 1 (older)
	proj1Dir := filepath.Join(tempDir, "novel-1")
	meta1Dir := filepath.Join(proj1Dir, "meta")
	os.MkdirAll(meta1Dir, 0755)

	p1 := domain.Progress{
		NovelName:          "Project One",
		CurrentChapter:     10,
		TotalRealWordCount: 15000,
	}
	p1Data, _ := json.Marshal(p1)
	p1Path := filepath.Join(meta1Dir, "progress.json")
	os.WriteFile(p1Path, p1Data, 0644)
	
	// Chtimes to ensure it's older
	olderTime := time.Now().Add(-2 * time.Hour)
	os.Chtimes(p1Path, olderTime, olderTime)

	// 2. Create a valid project 2 (newer)
	proj2Dir := filepath.Join(tempDir, "novel-2")
	meta2Dir := filepath.Join(proj2Dir, "meta")
	os.MkdirAll(meta2Dir, 0755)

	p2 := domain.Progress{
		NovelName:          "Project Two",
		CurrentChapter:     5,
		TotalRealWordCount: 8000,
	}
	p2Data, _ := json.Marshal(p2)
	p2Path := filepath.Join(meta2Dir, "progress.json")
	os.WriteFile(p2Path, p2Data, 0644)

	// 3. Create an invalid project (no progress.json)
	proj3Dir := filepath.Join(tempDir, "novel-empty")
	os.MkdirAll(proj3Dir, 0755)

	// Test DiscoverProjects
	projects, err := DiscoverProjects(tempDir)
	if err != nil {
		t.Fatalf("DiscoverProjects failed: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("Expected 2 projects, got %d", len(projects))
	}

	// Check sorting (newer should be first)
	if projects[0].DirName != "novel-2" {
		t.Errorf("Expected novel-2 to be first (newest), got %s", projects[0].DirName)
	}
	if projects[1].DirName != "novel-1" {
		t.Errorf("Expected novel-1 to be second (oldest), got %s", projects[1].DirName)
	}

	// Check data parsing
	if projects[0].NovelName != "Project Two" || projects[0].ChapterCount != 5 || projects[0].TotalRealWordCount != 8000 {
		t.Errorf("Project 2 data parsed incorrectly: %+v", projects[0])
	}
	if projects[1].NovelName != "Project One" || projects[1].ChapterCount != 10 || projects[1].TotalRealWordCount != 15000 {
		t.Errorf("Project 1 data parsed incorrectly: %+v", projects[1])
	}
}
