package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/sergi/go-diff/diffmatchpatch"
	"github.com/voocel/ainovel-cli/assets"
	"github.com/voocel/ainovel-cli/internal/bootstrap"
	"github.com/voocel/ainovel-cli/internal/host"
	"github.com/voocel/ainovel-cli/internal/store"
	"github.com/voocel/ainovel-cli/internal/tools"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx    context.Context
	cfg    bootstrap.Config
	bundle assets.Bundle
	host   *host.Host
	store  *store.Store
	mu     sync.Mutex
}

// NewApp creates a new App application struct
func NewApp(cfg bootstrap.Config, bundle assets.Bundle) *App {
	return &App{
		cfg:    cfg,
		bundle: bundle,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.host != nil {
		a.host.Close()
	}
}

func (a *App) resolveOutputDir(customDir string) string {
	if customDir != "" {
		if abs, err := filepath.Abs(customDir); err == nil {
			return abs
		}
		return customDir
	}

	out := a.cfg.OutputDir
	if out == "" {
		out = filepath.Join("output", "novel")
	}

	if filepath.IsAbs(out) {
		return out
	}

	candidates := []string{
		out,
		filepath.Join("..", out),
		filepath.Join("..", "..", out),
	}

	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(exeDir, out),
			filepath.Join(exeDir, "..", out),
			filepath.Join(exeDir, "..", "..", out),
			filepath.Join(exeDir, "..", "..", "..", out),
		)
	}

	for _, cand := range candidates {
		checkPath := filepath.Join(cand, "meta", "progress.json")
		if _, err := os.Stat(checkPath); err == nil {
			if abs, err := filepath.Abs(cand); err == nil {
				return abs
			}
			return cand
		}
	}

	if abs, err := filepath.Abs(out); err == nil {
		return abs
	}
	return out
}

// SelectProjectDir opens directory dialog to choose novel output dir
func (a *App) SelectProjectDir() (string, error) {
	dir, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Chọn thư mục dự án Novel",
	})
	if err != nil {
		return "", err
	}
	return dir, nil
}

// StartNovel starts a new novel writing session
func (a *App) StartNovel(prompt string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.host != nil {
		a.host.Close()
	}

	targetDir := a.resolveOutputDir("")
	cfg := a.cfg
	cfg.OutputDir = targetDir

	eng, err := host.New(cfg, a.bundle)
	if err != nil {
		return fmt.Errorf("lỗi khởi tạo host: %v", err)
	}
	a.host = eng
	a.store = store.NewStore(eng.Dir())

	eng.AskUser().SetHandler(func(ctx context.Context, questions []tools.Question) (*tools.AskUserResponse, error) {
		return nil, fmt.Errorf("ask-user needs async implementation")
	})

	err = eng.PrepareUserRules(prompt)
	if err != nil {
		return fmt.Errorf("lỗi chuẩn bị quy tắc: %v", err)
	}

	err = eng.StartPrepared(prompt)
	if err != nil {
		return fmt.Errorf("lỗi bắt đầu sáng tác: %v", err)
	}

	go a.listenEvents()
	return nil
}

func (a *App) ResumeNovel(dir string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.host != nil {
		a.host.Close()
	}

	targetDir := a.resolveOutputDir(dir)
	cfg := a.cfg
	cfg.OutputDir = targetDir

	eng, err := host.New(cfg, a.bundle)
	if err != nil {
		return fmt.Errorf("lỗi khởi tạo host (%s): %v", targetDir, err)
	}
	a.host = eng
	a.store = store.NewStore(eng.Dir())

	label, err := eng.Resume()
	if err != nil {
		return fmt.Errorf("lỗi khôi phục: %v", err)
	}
	if label == "" {
		return fmt.Errorf("không tìm thấy dự án dở dang trong thư mục:\n%s", targetDir)
	}

	go a.listenEvents()
	return nil
}

func (a *App) PauseNovel() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.host != nil {
		a.host.Close()
	}
	return nil
}

func (a *App) listenEvents() {
	for {
		select {
		case ev, ok := <-a.host.Events():
			if !ok {
				return
			}
			runtime.EventsEmit(a.ctx, "novel:event", ev)
			runtime.EventsEmit(a.ctx, "novel:snapshot", a.host.Snapshot())
		case delta, ok := <-a.host.Stream():
			if !ok {
				continue
			}
			if delta == host.StreamClearSentinel {
				runtime.EventsEmit(a.ctx, "novel:stream-clear", struct{}{})
			} else {
				runtime.EventsEmit(a.ctx, "novel:stream", delta)
			}
		case <-a.host.Done():
			runtime.EventsEmit(a.ctx, "novel:done", map[string]bool{"complete": true})
			return
		}
	}
}

// GetSnapshot returns the current state of the host
func (a *App) GetSnapshot() host.UISnapshot {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.host != nil {
		return a.host.Snapshot()
	}

	targetDir := a.resolveOutputDir("")
	st := store.NewStore(targetDir)
	progress, err := st.Progress.Load()
	if err == nil && progress != nil && progress.NovelName != "" {
		snap := host.UISnapshot{
			NovelName:         progress.NovelName,
			Phase:             string(progress.Phase),
			Flow:              string(progress.Flow),
			CurrentChapter:    progress.CurrentChapter,
			TotalChapters:     progress.TotalChapters,
			CompletedCount:    len(progress.CompletedChapters),
			TotalWordCount:    progress.TotalWordCount,
			InProgressChapter: progress.InProgressChapter,
			RuntimeState:      "stopped",
		}

		if usageState, err := st.Usage.Load(); err == nil && usageState != nil {
			snap.TotalInputTokens = usageState.Overall.Input
			snap.TotalOutputTokens = usageState.Overall.Output
			snap.TotalCacheReadTokens = usageState.Overall.CacheRead
			snap.TotalCacheWriteTokens = usageState.Overall.CacheWrite
			snap.TotalCostUSD = usageState.Overall.Cost
			snap.TotalSavedUSD = usageState.Overall.Saved
		}

		// Tải dữ liệu bổ sung để hiển thị khi offline
		if outline, err := st.Outline.LoadOutline(); err == nil && len(outline) > 0 {
			var uiOutline []host.OutlineSnapshot
			for _, node := range outline {
				uiOutline = append(uiOutline, host.OutlineSnapshot{
					Chapter:   node.Chapter,
					Title:     node.Title,
					CoreEvent: node.CoreEvent,
				})
			}
			snap.Outline = uiOutline
		}

		if premise, err := st.Outline.LoadPremise(); err == nil && premise != "" {
			snap.Premise = premise
		}

		if chars, err := st.Characters.Load(); err == nil && len(chars) > 0 {
			var charNames []string
			for _, char := range chars {
				charNames = append(charNames, char.Name)
			}
			snap.Characters = charNames
		}

		return snap
	}

	return host.UISnapshot{}
}

// GetChapterContent reads chapter markdown from store
func (a *App) GetChapterContent(chapterNum int) (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return "", fmt.Errorf("chưa có dự án nào được mở")
	}
	return a.store.Drafts.LoadChapterText(chapterNum)
}

// SaveChapterContent writes chapter markdown to store
func (a *App) SaveChapterContent(chapterNum int, content string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return fmt.Errorf("chưa có dự án nào được mở")
	}
	return a.store.Drafts.SaveFinalChapter(chapterNum, content)
}

type DiffResult struct {
	Diffs []diffmatchpatch.Diff `json:"diffs"`
}

// GetChapterDiff compares draft vs final
func (a *App) GetChapterDiff(chapterNum int) (DiffResult, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return DiffResult{}, fmt.Errorf("chưa có dự án nào được mở")
	}

	draft, err := a.store.Drafts.LoadDraft(chapterNum)
	if err != nil {
		return DiffResult{}, err
	}
	final, err := a.store.Drafts.LoadChapterText(chapterNum)
	if err != nil {
		return DiffResult{}, err
	}

	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(draft, final, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	return DiffResult{Diffs: diffs}, nil
}

// GetConfig reads config
func (a *App) GetConfig() bootstrap.Config {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.cfg
}

// UpdateConfig writes config
func (a *App) UpdateConfig(cfg bootstrap.Config) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.cfg = cfg

	// Save to config file
	path := bootstrap.DefaultConfigPath()
	if path != "" {
		data, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(path, data, 0644); err != nil {
			return err
		}
	}
	return nil
}

type CoCreateReply struct {
	Response string `json:"response"`
}

// CoCreate handles chatting with the AI
func (a *App) CoCreate(message string) (CoCreateReply, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.host == nil {
		return CoCreateReply{}, fmt.Errorf("chưa có dự án nào được mở")
	}
	// Simplified mock, need real implementation of CoCreate from host
	return CoCreateReply{Response: "Tính năng CoCreate đang được phát triển"}, nil
}

func (a *App) AnswerQuestion(answer string) {
	// TODO: Handle ask user bridge response
}
