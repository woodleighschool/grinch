package santa

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	syncv1 "buf.build/gen/go/northpolesec/protos/protocolbuffers/go/sync"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5"

	"github.com/woodleighschool/grinch/internal/rules"
	"github.com/woodleighschool/grinch/internal/store"
)

type ruleDownloadHandler struct {
	store    *store.Store
	logger   *slog.Logger
	compiler *rules.Compiler
}

func (h *ruleDownloadHandler) Handle(w http.ResponseWriter, r *http.Request) {
	machineIdentifier := strings.TrimSpace(chi.URLParam(r, "machineID"))
	if machineIdentifier == "" {
		respondError(w, http.StatusBadRequest, "machine id required")
		return
	}
	defer func() {
		if closeErr := r.Body.Close(); closeErr != nil {
			h.logger.Warn("failed to close request body", "err", closeErr)
		}
	}()
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		h.logger.Error("read rule download payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	var req syncv1.RuleDownloadRequest
	if err := unmarshalProtoJSON(bodyBytes, &req); err != nil {
		h.logger.Error("decode rule download payload", "err", err)
		respondError(w, http.StatusBadRequest, "invalid payload")
		return
	}
	ctx := r.Context()
	machine, err := h.store.GetMachineByIdentifier(ctx, machineIdentifier)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(w, http.StatusNotFound, "machine not registered")
			return
		}
		h.logger.Error("fetch machine", "err", err)
		respondError(w, http.StatusInternalServerError, "lookup failed")
		return
	}
	ruleRows, err := h.store.ListRules(ctx)
	if err != nil {
		h.logger.Error("list rules", "err", err)
		respondError(w, http.StatusInternalServerError, "rules unavailable")
		return
	}
	assignments, err := h.store.ListAllAssignments(ctx)
	if err != nil {
		h.logger.Error("list rule assignments", "err", err)
		respondError(w, http.StatusInternalServerError, "assignments unavailable")
		return
	}
	payload := h.compiler.BuildPayload(machine, ruleRows, assignments)
	if req.GetCursor() != "" && req.GetCursor() == payload.Cursor {
		respondProtoJSON(w, http.StatusOK, &syncv1.RuleDownloadResponse{Cursor: payload.Cursor})
		return
	}
	wireRules := make([]*syncv1.Rule, 0, len(payload.Rules))
	for _, rule := range payload.Rules {
		wireRules = append(wireRules, convertRule(rule))
	}
	if _, err := h.store.TouchMachine(ctx, machine.ID, machine.ClientVersion.String, machine.SyncCursor.String, payload.Cursor); err != nil {
		h.logger.Warn("touch machine", "err", err)
	}
	respondProtoJSON(w, http.StatusOK, &syncv1.RuleDownloadResponse{Cursor: payload.Cursor, Rules: wireRules})
}

func convertRule(rule rules.SyncRule) *syncv1.Rule {
	return &syncv1.Rule{
		Identifier: rule.Target,
		Policy:     mapPolicy(rule.Action),
		RuleType:   mapRuleType(rule.Type),
		CustomMsg:  rule.CustomMsg,
	}
}

func mapPolicy(action rules.RuleAction) syncv1.Policy {
	switch action {
	case rules.RuleActionBlock:
		return syncv1.Policy_BLOCKLIST
	default:
		return syncv1.Policy_ALLOWLIST
	}
}

func mapRuleType(value rules.RuleType) syncv1.RuleType {
	switch strings.ToLower(string(value)) {
	case "certificate":
		return syncv1.RuleType_CERTIFICATE
	case "signing_id", "signingid":
		return syncv1.RuleType_SIGNINGID
	case "teamid":
		return syncv1.RuleType_TEAMID
	case "cdhash":
		return syncv1.RuleType_CDHASH
	case "binary":
		return syncv1.RuleType_BINARY
	default:
		return syncv1.RuleType_RULETYPE_UNKNOWN
	}
}
