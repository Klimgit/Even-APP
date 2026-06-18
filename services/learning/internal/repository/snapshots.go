package repository

import (
	"context"
	"encoding/json"

	"github.com/even-app/even-app/services/learning/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type SnapshotBlockHit struct {
	LessonID uuid.UUID
	CourseID uuid.UUID
	Snapshot json.RawMessage
}

func FindBlockInSnapshots(ctx context.Context, pool *pgxpool.Pool, blockID uuid.UUID) (SnapshotBlockHit, error) {
	const q = `
SELECT lesson_id, course_id, snapshot
FROM published_lesson_snapshots
WHERE EXISTS (
    SELECT 1
    FROM jsonb_array_elements(snapshot->'blocks') AS blk
    WHERE (blk->>'id')::uuid = $1
)
LIMIT 1`
	var hit SnapshotBlockHit
	err := pool.QueryRow(ctx, q, blockID).Scan(&hit.LessonID, &hit.CourseID, &hit.Snapshot)
	return hit, err
}

func BlockFromSnapshotHit(hit SnapshotBlockHit, blockID uuid.UUID) (domain.BlockSnap, error) {
	snap, err := SnapshotFromRow(hit.Snapshot)
	if err != nil {
		return domain.BlockSnap{}, err
	}
	for _, b := range snap.Blocks {
		if b.ID == blockID {
			return b, nil
		}
	}
	return domain.BlockSnap{}, pgx.ErrNoRows
}
