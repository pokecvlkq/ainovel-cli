package tools

import (
	"testing"

	"github.com/voocel/ainovel-cli/internal/domain"
)

func TestParsePremiseSections(t *testing.T) {
	premise := `# Premise

## Đề tài và giọng điệu
Đông phương huyền huyễn, trưởng thành lạnh lùng.

## Định vị đề tài
Đông phương huyền huyễn thăng cấp lưu, hướng tới độc giả theo đuổi sảng điểm và thúc đẩy quan hệ.

## Xung đột cốt lõi
Nhân vật chính phải đưa ra lựa chọn giữa quy tắc tông môn và lương tri cá nhân.

## Chuyển hướng giữa chừng
Con đường tu luyện cũ mất tác dụng, bắt buộc chuyển sang hệ thống cấm thuật.
`

	sections := parsePremiseSections(premise)
	if sections["Đề tài và giọng điệu"] == "" {
		t.Fatalf("expected Đề tài và giọng điệu section, got %+v", sections)
	}
	if sections["Định vị đề tài"] == "" {
		t.Fatalf("expected Định vị đề tài section, got %+v", sections)
	}
	if sections["Xung đột cốt lõi"] == "" {
		t.Fatalf("expected Xung đột cốt lõi section, got %+v", sections)
	}
	if sections["Chuyển hướng giữa chừng"] == "" {
		t.Fatalf("expected Chuyển hướng giữa chừng alias normalized to Chuyển hướng giữa chừng, got %+v", sections)
	}
}

func TestPremiseStructure(t *testing.T) {
	premise := `## Đề tài và giọng điệu
Thăng cấp lưu, hơi lạnh lùng.

## Định vị đề tài
Thăng cấp lưu

## Xung đột cốt lõi
Xung đột

## Mục tiêu của nhân vật chính
Mục tiêu

## Hướng kết cục
Kết cục

## Vùng cấm khi viết
Vùng cấm

## Điểm thu hút khác biệt
Điểm bán

## Hook khác biệt
Hook

## Cam kết cốt lõi
Cam kết

## Động cơ câu chuyện
Động cơ

## Chuyển hướng giữa chừng
Chuyển hướng
`

	structure := premiseStructure(premise, domain.PlanningTierMid)
	if ready, _ := structure["template_ready"].(bool); !ready {
		t.Fatalf("expected template_ready, got %+v", structure)
	}
	missing, _ := structure["missing"].([]string)
	if len(missing) != 0 {
		t.Fatalf("expected no missing headings, got %+v", missing)
	}
}

func TestPremiseStructureShortAcceptsLegacyHeadingAlias(t *testing.T) {
	premise := `## Đề tài và giọng điệu
Một quyển giải cứu cao áp.

## Định vị đề tài
Mạo hiểm mật độ cao truyện ngắn.

## Xung đột cốt lõi
Nhân vật chính phải cứu con tin trong vòng một đêm.

## Mục tiêu của nhân vật chính
Cứu được con tin và sống sót rời đi.

## Hướng kết cục
Hoàn thành nhiệm vụ nhưng phải trả giá.

## Vùng cấm khi viết
Không mở rộng thành truyện dài nhiều kỳ.

## Điểm thu hút khác biệt
Áp lực thời gian và liên tục lật ngược tình thế.

## Hook khác biệt
Mỗi lần lựa chọn đều rút ngắn thời gian giải cứu.

## Cam kết cốt lõi
Cảm giác cấp bách, sự lựa chọn và lật ngược tình thế.

## Tính tương thích truyện ngắn
Mâu thuẫn cốt lõi và vòng cung nhân vật đều có thể hoàn thành trong một nhiệm vụ duy nhất.
`

	structure := premiseStructure(premise, domain.PlanningTierShort)
	if ready, _ := structure["template_ready"].(bool); !ready {
		t.Fatalf("expected short template_ready, got %+v", structure)
	}
}
