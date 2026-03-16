package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/teamsphere/server/internal/repository"
	"github.com/teamsphere/server/internal/service"
	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	messageService *service.MessageService
}

func NewSearchHandler(messageService *service.MessageService) *SearchHandler {
	return &SearchHandler{messageService: messageService}
}

// SearchMessages handles GET /search/messages?q=&room_id=&sender_id=&from=&to=&limit=
func (h *SearchHandler) SearchMessages(c *gin.Context) {
	q := c.Query("q")
	if q == "" {
		Error(c, http.StatusBadRequest, 40001, "q 涓嶈兘涓虹┖")
		return
	}

	roomID, err := parseInt64Query(c, "room_id")
	if err != nil || roomID <= 0 {
		Error(c, http.StatusBadRequest, 40001, "room_id 蹇呴』涓烘鏁存暟")
		return
	}

	senderID, err := parseInt64Query(c, "sender_id")
	if err != nil {
		Error(c, http.StatusBadRequest, 40001, "sender_id 蹇呴』涓烘鏁存暟")
		return
	}
	from, err := parseTimeQuery(c, "from")
	if err != nil {
		Error(c, http.StatusBadRequest, 40001, "from 鏃堕棿鏍煎紡閿欒")
		return
	}
	to, err := parseTimeQuery(c, "to")
	if err != nil {
		Error(c, http.StatusBadRequest, 40001, "to 鏃堕棿鏍煎紡閿欒")
		return
	}

	limit := parseOptionalInt(c.Query("limit"), 50)
	if limit > 100 {
		limit = 100
	}

	userID := c.GetInt64("user_id")
	results, err := h.messageService.SearchMessages(c.Request.Context(), userID, q, roomID, senderID, from, to, limit)
	if err != nil {
		handleRoomError(c, err)
		return
	}
	if results == nil {
		results = []*repository.MessageWithUser{}
	}
	Success(c, results)
}

func parseInt64Query(c *gin.Context, key string) (int64, error) {
	val := c.Query(key)
	if val == "" {
		return 0, nil
	}
	return strconv.ParseInt(val, 10, 64)
}

func parseTimeQuery(c *gin.Context, key string) (*time.Time, error) {
	val := c.Query(key)
	if val == "" {
		return nil, nil
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return &t, nil
	}
	if t, err := time.Parse("2006-01-02", val); err == nil {
		return &t, nil
	}
	return nil, strconv.ErrSyntax
}
