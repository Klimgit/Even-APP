package domain

import "github.com/google/uuid"

type LexemeUsageEntry struct {
	LessonID     uuid.UUID
	LessonTitle  string
	BlockID      uuid.UUID
	DisplayLabel string
	UsageKind    string
}

type LexemeUsageResult struct {
	LexemeID uuid.UUID
	Usages   []LexemeUsageEntry
}
