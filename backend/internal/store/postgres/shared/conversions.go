package pgutil

import (
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/woodleighschool/grinch/internal/domain"
)

func ToSource(value string) (domain.PrincipalSource, error) {
	return domain.ParsePrincipalSource(value)
}

func ToMemberKind(value string) (domain.MemberKind, error) {
	return domain.ParseMemberKind(value)
}

func ToRuleType(value string) (domain.RuleType, error) {
	return domain.ParseRuleType(value)
}

func ToRulePolicy(value string) (domain.RulePolicy, error) {
	return domain.ParseRulePolicy(value)
}

func ToRuleTargetAssignment(value string) (domain.RuleTargetAssignment, error) {
	return domain.ParseRuleTargetAssignment(value)
}

func ToRuleTargetSubjectKind(value string) (domain.RuleTargetSubjectKind, error) {
	return domain.ParseRuleTargetSubjectKind(value)
}

func UUIDPointer(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}

	result := uuid.UUID(value.Bytes)
	return &result
}
