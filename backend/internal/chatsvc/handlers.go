package chatsvc

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/bitik/backend/internal/apiresponse"
	"github.com/bitik/backend/internal/notify"
	"github.com/bitik/backend/internal/pgxutil"
	chatstore "github.com/bitik/backend/internal/store/chat"
	notifystore "github.com/bitik/backend/internal/store/notifications"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func parsePage(c *gin.Context) (page, limit, offset int32) {
	page = 1
	limit = 20
	if v, err := strconv.ParseInt(strings.TrimSpace(c.DefaultQuery("page", "1")), 10, 32); err == nil && v > 0 {
		page = int32(v)
	}
	if v, err := strconv.ParseInt(strings.TrimSpace(c.DefaultQuery("limit", "20")), 10, 32); err == nil && v > 0 {
		limit = int32(v)
	}
	if limit > 100 {
		limit = 100
	}
	offset = (page - 1) * limit
	return
}

func pageMeta(page, limit int32, total int64) map[string]any {
	hasNext := int64(page)*int64(limit) < total
	return map[string]any{"page": page, "limit": limit, "total": total, "has_next": hasNext}
}

func (s *Service) HandleListConversations(c *gin.Context) {
	uid, _ := currentUserID(c)
	page, limit, offset := parsePage(c)

	if s.isSeller(c) {
		sid, ok := s.sellerIDForUser(c, uid)
		if !ok {
			apiresponse.Error(c, http.StatusForbidden, "seller_required", "Seller account is required.")
			return
		}
		total, _ := s.chat.CountSellerConversations(c.Request.Context(), pgxutil.UUID(sid))
		rows, err := s.chat.ListSellerConversations(c.Request.Context(), chatstore.ListSellerConversationsParams{SellerID: pgxutil.UUID(sid), Limit: limit, Offset: offset})
		if err != nil {
			apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load conversations.")
			return
		}
		items := make([]gin.H, 0, len(rows))
		for _, r := range rows {
			items = append(items, gin.H{
				"id":              uuidString(r.ID),
				"buyer_id":        uuidString(r.BuyerID),
				"seller_id":       uuidString(r.SellerID),
				"last_message_at": timestamptzValue(r.LastMessageAt),
				"created_at":      r.CreatedAt.Time,
				"buyer": gin.H{
					"name":  r.BuyerName,
					"email": r.BuyerEmail,
				},
				"unread_count": r.UnreadCount,
			})
		}
		apiresponse.Respond(c, http.StatusOK, gin.H{"items": items}, map[string]any{"pagination": pageMeta(page, limit, total)})
		return
	}

	total, _ := s.chat.CountBuyerConversations(c.Request.Context(), pgxutil.UUID(uid))
	rows, err := s.chat.ListBuyerConversations(c.Request.Context(), chatstore.ListBuyerConversationsParams{BuyerID: pgxutil.UUID(uid), Limit: limit, Offset: offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load conversations.")
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, r := range rows {
		items = append(items, gin.H{
			"id":              uuidString(r.ID),
			"buyer_id":        uuidString(r.BuyerID),
			"seller_id":       uuidString(r.SellerID),
			"last_message_at": timestamptzValue(r.LastMessageAt),
			"created_at":      r.CreatedAt.Time,
			"seller": gin.H{
				"shop_name": r.ShopName,
				"slug":      r.Slug,
				"logo_url":  textValue(r.LogoUrl),
			},
			"unread_count": r.UnreadCount,
		})
	}
	apiresponse.Respond(c, http.StatusOK, gin.H{"items": items}, map[string]any{"pagination": pageMeta(page, limit, total)})
}

func (s *Service) HandleCreateConversation(c *gin.Context) {
	uid, _ := currentUserID(c)
	var req struct {
		SellerID string `json:"seller_id"`
		BuyerID  string `json:"buyer_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid conversation request.")
		return
	}

	var buyerID uuid.UUID = uid
	var sellerID uuid.UUID

	if s.isSeller(c) {
		// Seller can create by buyer_id.
		if strings.TrimSpace(req.BuyerID) == "" {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "buyer_id", Message: "is required"}})
			return
		}
		b, err := uuid.Parse(req.BuyerID)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "buyer_id", Message: "must be a valid uuid"}})
			return
		}
		buyerID = b
		sid, ok := s.sellerIDForUser(c, uid)
		if !ok {
			apiresponse.Error(c, http.StatusForbidden, "seller_required", "Seller account is required.")
			return
		}
		sellerID = sid
	} else {
		// Buyer creates by seller_id.
		if strings.TrimSpace(req.SellerID) == "" {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "seller_id", Message: "is required"}})
			return
		}
		sid, err := uuid.Parse(req.SellerID)
		if err != nil {
			apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "seller_id", Message: "must be a valid uuid"}})
			return
		}
		sellerID = sid
	}

	conv, err := s.chat.GetOrCreateConversation(c.Request.Context(), chatstore.GetOrCreateConversationParams{
		BuyerID:  pgxutil.UUID(buyerID),
		SellerID: pgxutil.UUID(sellerID),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "conversation_create_failed", "Could not create conversation.")
		return
	}
	apiresponse.Respond(c, http.StatusCreated, gin.H{
		"id":         uuidString(conv.ID),
		"buyer_id":   uuidString(conv.BuyerID),
		"seller_id":  uuidString(conv.SellerID),
		"created_at": conv.CreatedAt.Time,
	}, nil)
}

func (s *Service) HandleListMessages(c *gin.Context) {
	uid, _ := currentUserID(c)
	cid, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_conversation_id", "Invalid conversation id.")
		return
	}
	if _, err := s.chat.GetConversationForUser(c.Request.Context(), chatstore.GetConversationForUserParams{
		ID:     pgxutil.UUID(cid),
		UserID: pgxutil.UUID(uid),
	}); err != nil {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Not a participant in this conversation.")
		return
	}
	page, limit, offset := parsePage(c)
	rows, err := s.chat.ListMessages(c.Request.Context(), chatstore.ListMessagesParams{ConversationID: pgxutil.UUID(cid), Limit: limit, Offset: offset})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not load messages.")
		return
	}
	items := make([]gin.H, 0, len(rows))
	for _, m := range rows {
		items = append(items, gin.H{
			"id":              uuidString(m.ID),
			"conversation_id": uuidString(m.ConversationID),
			"sender_user_id":  uuidString(m.SenderUserID),
			"message":         textValue(m.Message),
			"attachment_url":  textValue(m.AttachmentUrl),
			"read_at":         timestamptzValue(m.ReadAt),
			"created_at":      m.CreatedAt.Time,
		})
	}
	// v1: no total count endpoint yet; return page metadata without total.
	apiresponse.Respond(c, http.StatusOK, gin.H{"items": items}, map[string]any{"pagination": map[string]any{"page": page, "limit": limit}})
}

func (s *Service) HandleSendMessage(c *gin.Context) {
	uid, _ := currentUserID(c)
	cid, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_conversation_id", "Invalid conversation id.")
		return
	}
	conv, err := s.chat.GetConversationForUser(c.Request.Context(), chatstore.GetConversationForUserParams{
		ID:     pgxutil.UUID(cid),
		UserID: pgxutil.UUID(uid),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Not a participant in this conversation.")
		return
	}
	var req struct {
		Message       string `json:"message"`
		AttachmentURL string `json:"attachment_url"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_request", "Invalid message request.")
		return
	}
	if strings.TrimSpace(req.Message) == "" && strings.TrimSpace(req.AttachmentURL) == "" {
		apiresponse.ValidationError(c, []apiresponse.FieldError{{Field: "message", Message: "message or attachment_url is required"}})
		return
	}
	msg, err := s.chat.CreateMessage(c.Request.Context(), chatstore.CreateMessageParams{
		ConversationID: pgxutil.UUID(cid),
		SenderUserID:   pgxutil.UUID(uid),
		Message:        text(req.Message),
		AttachmentUrl:  text(req.AttachmentURL),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "message_send_failed", "Could not send message.")
		return
	}
	_ = s.chat.TouchConversationLastMessageAt(c.Request.Context(), pgxutil.UUID(cid))

	// Publish event to the other participant (best-effort).
	if s.pub != nil {
		recipient := ""
		if conv.BuyerID == pgxutil.UUID(uid) {
			if v, ok := pgxutil.ToUUID(conv.SellerUserID); ok {
				recipient = v.String()
			}
		} else {
			if v, ok := pgxutil.ToUUID(conv.BuyerID); ok {
				recipient = v.String()
			}
		}
		if recipient != "" {
			recipientID, parseErr := uuid.Parse(recipient)
			if parseErr == nil && s.notifs != nil {
				createdNotif, err := s.notifs.CreateInAppNotification(c.Request.Context(), notifystore.CreateInAppNotificationParams{
					UserID: pgxutil.UUID(recipientID),
					Type:   "chat_message",
					Title:  "New chat message",
					Body:   text("You received a new message."),
					Data: jsonObject(map[string]any{
						"conversation_id": cid.String(),
						"message_id":      uuidString(msg.ID),
						"sender_user_id":  uid.String(),
					}),
				})
				if err == nil {
					s.pub.Publish(c.Request.Context(), notify.Event{
						Type:   notify.EventNotificationCreated,
						UserID: recipient,
						Data: map[string]any{
							"notification_id": uuidString(createdNotif.ID),
							"type":            "chat_message",
							"conversation_id": cid.String(),
						},
					})
				}
			}
			s.pub.Publish(c.Request.Context(), notify.Event{
				Type:   notify.EventChatMessageCreated,
				UserID: recipient,
				Data: map[string]any{
					"conversation_id": cid.String(),
					"message_id":      uuidString(msg.ID),
					"sender_user_id":  uid.String(),
				},
			})
		}
	}

	apiresponse.Respond(c, http.StatusCreated, gin.H{
		"id":              uuidString(msg.ID),
		"conversation_id": uuidString(msg.ConversationID),
		"sender_user_id":  uuidString(msg.SenderUserID),
		"message":         textValue(msg.Message),
		"attachment_url":  textValue(msg.AttachmentUrl),
		"read_at":         timestamptzValue(msg.ReadAt),
		"created_at":      msg.CreatedAt.Time,
	}, nil)
}

func (s *Service) HandleDeleteMessage(c *gin.Context) {
	uid, _ := currentUserID(c)
	cid, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_conversation_id", "Invalid conversation id.")
		return
	}
	mid, err := uuid.Parse(c.Param("message_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_message_id", "Invalid message id.")
		return
	}
	if _, err := s.chat.GetConversationForUser(c.Request.Context(), chatstore.GetConversationForUserParams{
		ID:     pgxutil.UUID(cid),
		UserID: pgxutil.UUID(uid),
	}); err != nil {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Not a participant in this conversation.")
		return
	}
	rows, err := s.chat.DeleteMessageForUser(c.Request.Context(), chatstore.DeleteMessageForUserParams{
		ID:      pgxutil.UUID(mid),
		ID_2:    pgxutil.UUID(cid),
		BuyerID: pgxutil.UUID(uid),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete message.")
		return
	}
	if rows == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Message not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleDeleteConversation(c *gin.Context) {
	uid, _ := currentUserID(c)
	cid, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_conversation_id", "Invalid conversation id.")
		return
	}
	rows, err := s.chat.DeleteConversationForUser(c.Request.Context(), chatstore.DeleteConversationForUserParams{
		ID:      pgxutil.UUID(cid),
		BuyerID: pgxutil.UUID(uid),
	})
	if err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not delete conversation.")
		return
	}
	if rows == 0 {
		apiresponse.Error(c, http.StatusNotFound, "not_found", "Conversation not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

func (s *Service) HandleMarkRead(c *gin.Context) {
	uid, _ := currentUserID(c)
	cid, err := uuid.Parse(c.Param("conversation_id"))
	if err != nil {
		apiresponse.Error(c, http.StatusBadRequest, "invalid_conversation_id", "Invalid conversation id.")
		return
	}
	if _, err := s.chat.GetConversationForUser(c.Request.Context(), chatstore.GetConversationForUserParams{
		ID:     pgxutil.UUID(cid),
		UserID: pgxutil.UUID(uid),
	}); err != nil {
		apiresponse.Error(c, http.StatusForbidden, "forbidden", "Not a participant in this conversation.")
		return
	}
	if err := s.chat.MarkConversationRead(c.Request.Context(), chatstore.MarkConversationReadParams{ConversationID: pgxutil.UUID(cid), SenderUserID: pgxutil.UUID(uid)}); err != nil {
		apiresponse.Error(c, http.StatusInternalServerError, "internal_error", "Could not mark read.")
		return
	}
	c.Status(http.StatusNoContent)
}

// Helpers (local)

func uuidString(id pgtype.UUID) string {
	if v, ok := pgxutil.ToUUID(id); ok {
		return v.String()
	}
	return ""
}

func textValue(t pgtype.Text) any {
	if !t.Valid {
		return nil
	}
	return t.String
}

func timestamptzValue(t pgtype.Timestamptz) any {
	if !t.Valid {
		return nil
	}
	return t.Time
}

func text(raw string) pgtype.Text {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: raw, Valid: true}
}
