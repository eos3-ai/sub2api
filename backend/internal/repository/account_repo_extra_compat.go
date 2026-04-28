package repository

import "strings"

func shouldEnqueueSchedulerOutboxForExtraUpdates(updates map[string]any) bool {
	if len(updates) == 0 {
		return false
	}
	for key := range updates {
		if strings.TrimSpace(key) != "" {
			return true
		}
	}
	return false
}
