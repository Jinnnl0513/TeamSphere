package repository

import (
	"context"
	"errors"

	"github.com/teamsphere/server/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ReactionRepo struct {
	pool *pgxpool.Pool
}

func NewReactionRepo(pool *pgxpool.Pool) *ReactionRepo {
	return &ReactionRepo{pool: pool}
}

func (r *ReactionRepo) Add(ctx context.Context, messageID int64, messageType string, userID int64, emoji string) (bool, error) {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO message_reactions (message_id, message_type, user_id, emoji)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT ON CONSTRAINT message_reactions_unique DO NOTHING`,
		messageID, messageType, userID, emoji,
	)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (r *ReactionRepo) Remove(ctx context.Context, messageID int64, messageType string, userID int64, emoji string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM message_reactions
		 WHERE message_id=$1 AND message_type=$2 AND user_id=$3 AND emoji=$4`,
		messageID, messageType, userID, emoji,
	)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func (r *ReactionRepo) ListByMessage(ctx context.Context, messageID int64, messageType string) ([]*model.ReactionSummary, error) {
	result, err := r.ListByMessages(ctx, []int64{messageID}, messageType)
	if err != nil {
		return nil, err
	}
	if summaries, ok := result[messageID]; ok {
		return summaries, nil
	}
	return []*model.ReactionSummary{}, nil
}

func (r *ReactionRepo) ListByMessages(ctx context.Context, messageIDs []int64, messageType string) (map[int64][]*model.ReactionSummary, error) {
	if len(messageIDs) == 0 {
		return map[int64][]*model.ReactionSummary{}, nil
	}

	rows, err := r.pool.Query(ctx,
		`SELECT
			message_id,
			emoji,
			COUNT(*)               AS count,
			ARRAY_AGG(user_id ORDER BY created_at LIMIT 20) AS user_ids
		 FROM message_reactions
		 WHERE message_id = ANY($1) AND message_type = $2
		 GROUP BY message_id, emoji
		 ORDER BY message_id, count DESC`,
		messageIDs, messageType,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return map[int64][]*model.ReactionSummary{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	result := make(map[int64][]*model.ReactionSummary)
	for rows.Next() {
		var msgID int64
		var emoji string
		var count int
		var userIDs []int64
		if err := rows.Scan(&msgID, &emoji, &count, &userIDs); err != nil {
			return nil, err
		}
		result[msgID] = append(result[msgID], &model.ReactionSummary{
			Emoji:   emoji,
			Count:   count,
			UserIDs: userIDs,
		})
	}
	return result, rows.Err()
}
