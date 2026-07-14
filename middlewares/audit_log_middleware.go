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
// concerns — checked in order: the JSON body's own branch_id field, the
// first entry of a branch_ids array (e.g. creating a sub-user with
// multiple branches assigned at once — there's no single "the" branch
// there, so the first one is used as a best-effort attribution), then a
// branch_id query param. Many actions genuinely have no branch context at
// all, so returning nil here is expected and fine, not an error case.
func extractBranchID(bodyJSON []byte, queryVals map[string][]string) *uint {
	var m struct {
		BranchID  *uint  `json:"branch_id"`
		BranchIDs []uint `json:"branch_ids"`
	}
	if len(bodyJSON) > 0 {
		if err := json.Unmarshal(bodyJSON, &m); err == nil {
			if m.BranchID != nil && *m.BranchID != 0 {
				return m.BranchID
			}
			if len(m.BranchIDs) > 0 && m.BranchIDs[0] != 0 {
				id := m.BranchIDs[0]
				return &id
			}
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

// branchScopedTables maps a URL path segment (the plural resource name, as
// it appears right after /api/v1/) to the DB table that has its own
// branch_id column — used only for DELETE requests, where there's no
// request body to read branch_id from at all (the ID is in the URL, not a
// JSON body). Extend this as other branch-scoped resources gain DELETE
// endpoints.
var branchScopedTables = map[string]string{
	"company-banks": "company_banks",
	"product-types": "product_types",
	"clients":       "clients",
	"deposits":      "deposits",
	"withdrawals":   "withdrawals",
}

// lookupBranchIDForDelete extracts the trailing numeric ID from a DELETE
// request's own path (e.g. "/api/v1/product-types/42" → table
// "product_types", id 42) and looks up that specific row's branch_id.
// MUST be called BEFORE c.Next() — by the time this middleware's
// post-request code would normally run, a successful delete has already
// removed the row, so looking it up after the fact would always miss.
func lookupBranchIDForDelete(path string) *uint {
	if authDB == nil {
		return nil
	}
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 2 {
		return nil
	}
	idStr := segments[len(segments)-1]
	slug := segments[len(segments)-2]
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil || id == 0 {
		return nil
	}
	table, ok := branchScopedTables[slug]
	if !ok {
		return nil
	}
	var row struct{ BranchID *uint }
	// table is only ever one of the fixed values in branchScopedTables
	// above, never taken from request input directly.
	if err := authDB.Table(table).Select("branch_id").Where("id = ?", id).Scan(&row).Error; err != nil {
		return nil
	}
	return row.BranchID
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

		// Must happen BEFORE c.Next() — see lookupBranchIDForDelete.
		var preDeleteBranchID *uint
		if method == "DELETE" {
			preDeleteBranchID = lookupBranchIDForDelete(c.Request.URL.Path)
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
		if branchID == nil {
			branchID = preDeleteBranchID
		}
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
