package middlewares

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"crm-backend/models"
)

// sensitiveJSONKeys are redacted from any request body before it's stored —
// never write a raw password/token/secret into the audit log itself.
var sensitiveJSONKeys = map[string]bool{
	"password":           true,
	"old_password":       true,
	"new_password":       true,
	"confirm_password":   true,
	"token":              true,
	"telegram_bot_token": true,
	"secret":             true,
}

const maxAuditBodyLen = 4000 // characters — a defensive cap, not a hard product requirement

// redactSensitiveFields walks a shallow JSON object and replaces any
// sensitive key's value with "***". Deliberately shallow (not recursive
// into nested objects/arrays) — good enough for this app's flat request
// bodies, and simpler/faster than a full recursive walk on every request.
func redactSensitiveFields(raw []byte) string {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		// Not a JSON object (empty body, array body, etc.) — nothing to
		// redact, just cap the length.
		s := string(raw)
		if len(s) > maxAuditBodyLen {
			return s[:maxAuditBodyLen] + "…"
		}
		return s
	}
	for k := range m {
		if sensitiveJSONKeys[strings.ToLower(k)] {
			m[k] = "***"
		}
	}
	out, err := json.Marshal(m)
	if err != nil {
		return ""
	}
	s := string(out)
	if len(s) > maxAuditBodyLen {
		return s[:maxAuditBodyLen] + "…"
	}
	return s
}

// extractBranchID makes a best-effort guess at which branch a request
// concerns — from the JSON body's own branch_id field, falling back to a
// branch_id query param. Many actions genuinely have no branch context, so
// returning nil here is expected and fine, not an error case.
func extractBranchID(bodyJSON []byte, queryVals map[string][]string) *uint {
	var m struct {
		BranchID *uint `json:"branch_id"`
	}
	if len(bodyJSON) > 0 {
		if err := json.Unmarshal(bodyJSON, &m); err == nil && m.BranchID != nil && *m.BranchID != 0 {
			return m.BranchID
		}
	}
	if vals, ok := queryVals["branch_id"]; ok && len(vals) > 0 {
		if id, err := strconv.ParseUint(vals[0], 10, 64); err == nil && id != 0 {
			bid := uint(id)
			return &bid
		}
	}
	return nil
}

// AuditLog records every authenticated, state-changing request (POST/PUT/
// PATCH/DELETE — GETs are read-only and deliberately not logged here,
// they'd dominate the table with no real audit value) to the database,
// tagged with who did it and, best-effort, which branch it concerns.
// Super Admin actions are deliberately excluded entirely — this log is
// meant to track regular staff/branch activity, not the account that
// already bypasses every permission check.
// Reuses the same authDB reference set by InitAuth (this middleware only
// makes sense alongside Auth(), which is what populates CtxUserID).
// Writes happen in a goroutine — a slow/failed audit write must never
// delay or break the actual request that triggered it.
func AuditLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		if method != "POST" && method != "PUT" && method != "PATCH" && method != "DELETE" {
			c.Next()
			return
		}

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		c.Next()

		if authDB == nil {
			return
		}

		var userID uint
		var isSuperAdmin bool
		if userIDVal, exists := c.Get(CtxUserID); exists {
			if uid, ok := userIDVal.(uint); ok {
				userID = uid
			}
		}
		if isSAVal, exists := c.Get(CtxSuperAdmin); exists {
			if isSA, ok := isSAVal.(bool); ok {
				isSuperAdmin = isSA
			}
		}

		if userID == 0 {
			return // unauthenticated request (e.g. login itself) — nothing to attribute this to
		}
		if isSuperAdmin {
			return // Super Admin actions are deliberately excluded from the audit log
		}

		branchID := extractBranchID(bodyBytes, c.Request.URL.Query())
		redactedBody := redactSensitiveFields(bodyBytes)
		entry := &models.AuditLog{
			UserID:      userID,
			BranchID:    branchID,
			Method:      method,
			Path:        c.Request.URL.Path,
			StatusCode:  c.Writer.Status(),
			RequestBody: redactedBody,
			IPAddress:   c.ClientIP(),
			UserAgent:   c.Request.UserAgent(),
		}
		go func() {
			if err := authDB.Create(entry).Error; err != nil {
				log.Printf("[audit] failed to write audit log entry: %v", err)
			}
		}()
	}
}
