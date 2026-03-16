package handler

import (
    "errors"
    "net/http"
    "strconv"

    "github.com/teamsphere/server/internal/service"
    "github.com/gin-gonic/gin"
)

type RoomHandler struct {
    roomService    *service.RoomService
    messageService *service.MessageService
    readService    *service.MessageReadService
    auditService   *service.AuditLogService
    settingsService *service.RoomSettingsService
}

func NewRoomHandler(roomService *service.RoomService, messageService *service.MessageService, readService *service.MessageReadService, auditService *service.AuditLogService, settingsService *service.RoomSettingsService) *RoomHandler {
    return &RoomHandler{roomService: roomService, messageService: messageService, readService: readService, auditService: auditService, settingsService: settingsService}
}

// helpers

func parseIDParam(c *gin.Context, name string) (int64, bool) {
    s := c.Param(name)
    id, err := strconv.ParseInt(s, 10, 64)
    if err != nil || id <= 0 {
        Error(c, http.StatusBadRequest, 40001, name+" 蹇呴』鏄鏁存暟")
        return 0, false
    }
    return id, true
}

func handleRoomError(c *gin.Context, err error) {
    switch {
    case errors.Is(err, service.ErrRoomNotFound):
        Error(c, http.StatusNotFound, 40401, "资源不存在")
    case errors.Is(err, service.ErrNotRoomMember):
        Error(c, http.StatusForbidden, 40301, "涓嶆槸鎴块棿鎴愬憳")
    case errors.Is(err, service.ErrNoPermission):
        Error(c, http.StatusForbidden, 40301, "娌℃湁鏉冮檺")
    case errors.Is(err, service.ErrAlreadyMember):
        Error(c, http.StatusConflict, 40902, "资源冲突")
    case errors.Is(err, service.ErrInvitePending):
        Error(c, http.StatusConflict, 40903, "资源冲突")
    case errors.Is(err, service.ErrJoinApprovalRequired):
        Error(c, http.StatusAccepted, 20201, "鍔犲叆璇锋眰宸叉彁浜わ紝绛夊緟瀹℃壒")
    case errors.Is(err, service.ErrJoinInviteOnly):
        Error(c, http.StatusForbidden, 40306, "没有权限")
    case errors.Is(err, service.ErrJoinRequestPending):
        Error(c, http.StatusConflict, 40904, "资源冲突")
    case errors.Is(err, service.ErrRoomReadOnly):
        Error(c, http.StatusForbidden, 40307, "棰戦亾鍙")
    case errors.Is(err, service.ErrRoomSendForbidden):
        Error(c, http.StatusForbidden, 40308, "没有权限")
    case errors.Is(err, service.ErrRoomUploadForbidden):
        Error(c, http.StatusForbidden, 40309, "没有权限")
    case errors.Is(err, service.ErrContentBlocked):
        Error(c, http.StatusForbidden, 40310, "没有权限")
    case errors.Is(err, service.ErrInviteNotFriend):
        Error(c, http.StatusForbidden, 40302, "没有权限")
    case errors.Is(err, service.ErrInviteNotFound):
        Error(c, http.StatusNotFound, 40401, "閭€璇蜂笉瀛樺湪")
    case errors.Is(err, service.ErrCannotLeaveOwner):
        Error(c, http.StatusForbidden, 40301, "缇や富涓嶈兘鐩存帴閫€缇わ紝璇峰厛杞缇や富韬唤")
    case errors.Is(err, service.ErrCannotActOnHigher):
        Error(c, http.StatusForbidden, 40301, "鏃犳硶瀵瑰悓绾ф垨鏇撮珮瑙掕壊鎴愬憳杩涜鎿嶄綔")
    case errors.Is(err, service.ErrRoomNameTaken):
        Error(c, http.StatusConflict, 40901, "资源冲突")
    case errors.Is(err, service.ErrTargetNotMember):
        Error(c, http.StatusNotFound, 40401, "鐩爣鐢ㄦ埛涓嶆槸鎴块棿鎴愬憳")
    case errors.Is(err, service.ErrMessageNotFound):
        Error(c, http.StatusNotFound, 40401, "资源不存在")
    case errors.Is(err, service.ErrRecallTimeout):
        Error(c, http.StatusForbidden, 40304, "宸茶秴杩囨秷鎭彲鎾ゅ洖鏃堕棿")
    case errors.Is(err, service.ErrRecallForbidden):
        Error(c, http.StatusForbidden, 40305, "没有权限")
    case errors.Is(err, service.ErrAlreadyRecalled):
        Error(c, http.StatusConflict, 40901, "娑堟伅宸茶鎾ゅ洖")
    case errors.Is(err, service.ErrEditForbidden):
        Error(c, http.StatusForbidden, 40301, "鍙湁鍙戦€佽€呭彲浠ョ紪杈戣娑堟伅")
    default:
        Error(c, http.StatusInternalServerError, 50001, "服务器内部错误")
    }
}

func (h *RoomHandler) audit(c *gin.Context, action, entityType string, entityID int64, meta any) {
    if h.auditService == nil {
        return
    }
    userID := c.GetInt64("user_id")
    ip := c.ClientIP()
    ua := c.GetHeader("User-Agent")
    _ = h.auditService.Record(c.Request.Context(), userID, action, entityType, entityID, meta, ip, ua)
}
