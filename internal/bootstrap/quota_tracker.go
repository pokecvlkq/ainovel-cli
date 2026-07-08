package bootstrap

import (
	"sync"
	"time"
)

// ProviderStatus định nghĩa các trạng thái của provider
type ProviderStatus string

const (
	StatusActive   ProviderStatus = "Active"
	StatusDead     ProviderStatus = "Dead"
	StatusCooldown ProviderStatus = "Cooldown"
)

// ProviderStatusSnapshot chứa thông tin snapshot trạng thái để hiển thị UI
type ProviderStatusSnapshot struct {
	Provider  string
	Status    ProviderStatus
	Reason    string
	ResumeAt  time.Time
	DeadSince time.Time // Thêm để hỗ trợ auto-revival
}

// QuotaTracker theo dõi trạng thái hạn mức và rate limit của các provider
type QuotaTracker struct {
	mu       sync.RWMutex
	statuses map[string]*ProviderStatusSnapshot
}

// NewQuotaTracker khởi tạo QuotaTracker mới
func NewQuotaTracker() *QuotaTracker {
	return &QuotaTracker{
		statuses: make(map[string]*ProviderStatusSnapshot),
	}
}

// GlobalQuotaTracker là instance global để sử dụng trong toàn bộ app
var GlobalQuotaTracker = NewQuotaTracker()

// MarkQuotaExhausted đánh dấu provider đã hết quota (chết hẳn)
func (qt *QuotaTracker) MarkQuotaExhausted(provider string) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	qt.statuses[provider] = &ProviderStatusSnapshot{
		Provider:  provider,
		Status:    StatusDead,
		Reason:    "Out of Funds / Billing Error",
		DeadSince: time.Now(),
	}
}

// MarkActive đánh dấu provider hoạt động bình thường, xoá lỗi
func (qt *QuotaTracker) MarkActive(provider string) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	// Xóa khỏi danh sách lỗi để trở về mặc định là Active
	delete(qt.statuses, provider)
}

// MarkCooldown đánh dấu provider bị rate limit hoặc quá tải, cần nghỉ 60s
func (qt *QuotaTracker) MarkCooldown(provider, reason string) {
	qt.mu.Lock()
	defer qt.mu.Unlock()
	
	if reason == "" {
		reason = "Rate Limit / Overloaded"
	}
	
	qt.statuses[provider] = &ProviderStatusSnapshot{
		Provider: provider,
		Status:   StatusCooldown,
		Reason:   reason,
		ResumeAt: time.Now().Add(180 * time.Second), // 3 phút thay vì 60s
	}
}

// IsAvailable kiểm tra xem provider có sẵn sàng để gọi không
func (qt *QuotaTracker) IsAvailable(provider string) bool {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	
	status, exists := qt.statuses[provider]
	if !exists {
		return true // Mặc định là Active nếu chưa có lỗi gì
	}

	if status.Status == StatusDead {
		// Auto-revival: Nếu chết quá 5 phút, cho phép thử lại
		if !status.DeadSince.IsZero() && time.Since(status.DeadSince) > 5*time.Minute {
			return true
		}
		return false
	}

	if status.Status == StatusCooldown {
		// Nếu đã qua thời gian cooldown, cho phép gọi lại
		if time.Now().After(status.ResumeAt) {
			return true
		}
		return false
	}

	return true
}

// AllStatuses trả về danh sách snapshot trạng thái của tất cả provider
func (qt *QuotaTracker) AllStatuses() []ProviderStatusSnapshot {
	qt.mu.RLock()
	defer qt.mu.RUnlock()
	
	snapshots := make([]ProviderStatusSnapshot, 0, len(qt.statuses))
	for _, status := range qt.statuses {
		snapshots = append(snapshots, *status)
	}
	return snapshots
}
