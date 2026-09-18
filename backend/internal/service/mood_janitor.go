package service

import (
	"time"
)

// StartMoodTrashJanitor 定期把超过七天的回收站情绪记录永久清除。
// 它在启动时立即执行一次，之后每小时执行一次；退出时由 stop 通道结束。
func StartMoodTrashJanitor(s *MoodService, stop <-chan struct{}) {
	run := func() {
		if _, e := s.PurgeExpired(); e != nil {
			s.logger.Error("mood trash janitor purge failed", "error", e)
		}
	}
	run()
	t := time.NewTicker(time.Hour)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			run()
		case <-stop:
			return
		}
	}
}
