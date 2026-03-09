package utils

import (
	"testing"
)

func TestCountTasks(t *testing.T) {
	content := `## Tasks
- [x] Completed task 1
- [ ] Pending task 2
- [x] Completed task 3
- [ ] Pending task 4
`

	progress := CountTasks(content)
	if progress.Total != 4 {
		t.Errorf("expected 4 total, got %d", progress.Total)
	}
	if progress.Completed != 2 {
		t.Errorf("expected 2 completed, got %d", progress.Completed)
	}
}

func TestCountTasks_NoTasks(t *testing.T) {
	content := `## Section
Just some text, no tasks here.
`

	progress := CountTasks(content)
	if progress.Total != 0 {
		t.Errorf("expected 0 total, got %d", progress.Total)
	}
	if progress.Completed != 0 {
		t.Errorf("expected 0 completed, got %d", progress.Completed)
	}
}

func TestCountTasks_AllComplete(t *testing.T) {
	content := `- [x] Task 1
- [x] Task 2
- [x] Task 3
`

	progress := CountTasks(content)
	if progress.Total != 3 {
		t.Errorf("expected 3 total, got %d", progress.Total)
	}
	if progress.Completed != 3 {
		t.Errorf("expected 3 completed, got %d", progress.Completed)
	}
}
