package ws

import (
	"context"

	"github.com/teamsphere/server/internal/model"
)

// buildReplyInfo fetches a room message by ID and returns a ReplyInfo snapshot.
func (h *Hub) buildReplyInfo(ctx context.Context, replyToID int64) *model.ReplyInfo {
	origMsg, err := h.messageRepo.GetByID(ctx, replyToID)
	if err != nil || origMsg == nil {
		return nil
	}
	ri := &model.ReplyInfo{
		ID:        origMsg.ID,
		MsgType:   origMsg.MsgType,
		IsDeleted: origMsg.DeletedAt != nil,
	}
	if origMsg.DeletedAt == nil && len([]rune(origMsg.Content)) > 0 {
		runes := []rune(origMsg.Content)
		if len(runes) > 200 {
			ri.Content = string(runes[:200]) + "..."
		} else {
			ri.Content = origMsg.Content
		}
	}
	if origMsg.UserID != nil {
		u, err := h.userRepo.GetByID(ctx, *origMsg.UserID)
		if err == nil && u != nil {
			ri.User = model.UserInfo{
				ID:        u.ID,
				Username:  u.Username,
				AvatarURL: u.AvatarURL,
			}
		}
	}
	return ri
}

// buildDMReplyInfo fetches a DM by ID and returns a ReplyInfo snapshot.
func (h *Hub) buildDMReplyInfo(ctx context.Context, replyToID int64) *model.ReplyInfo {
	origDM, err := h.messageRepo.GetDMByID(ctx, replyToID)
	if err != nil || origDM == nil {
		return nil
	}
	ri := &model.ReplyInfo{
		ID:        origDM.ID,
		MsgType:   origDM.MsgType,
		IsDeleted: origDM.DeletedAt != nil,
	}
	if origDM.DeletedAt == nil {
		runes := []rune(origDM.Content)
		if len(runes) > 200 {
			ri.Content = string(runes[:200]) + "..."
		} else {
			ri.Content = origDM.Content
		}
	}
	u, err := h.userRepo.GetByID(ctx, origDM.SenderID)
	if err == nil && u != nil {
		ri.User = model.UserInfo{
			ID:        u.ID,
			Username:  u.Username,
			AvatarURL: u.AvatarURL,
		}
	}
	return ri
}
