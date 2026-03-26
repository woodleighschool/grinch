package groups_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/app/groups"
	"github.com/woodleighschool/grinch/internal/domain"
)

type testStore struct {
	t *testing.T

	group          domain.Group
	getGroupErr    error
	memberships    []domain.Membership
	listMembersErr error

	createMembershipCalls int
	deleteMembershipCalls int

	syncedMachineIDs []uuid.UUID
	syncedUserIDs    []uuid.UUID
}

func (s *testStore) ListGroups(context.Context, domain.ListOptions) ([]domain.Group, int32, error) {
	s.unexpectedCall("ListGroups")
	return nil, 0, nil
}

func (s *testStore) GetGroup(context.Context, uuid.UUID) (domain.Group, error) {
	if s.getGroupErr != nil {
		return domain.Group{}, s.getGroupErr
	}

	return s.group, nil
}

func (s *testStore) CreateLocalGroup(context.Context, string, string) (domain.Group, error) {
	s.unexpectedCall("CreateLocalGroup")
	return domain.Group{}, nil
}

func (s *testStore) UpdateGroup(context.Context, uuid.UUID, string, string) (domain.Group, error) {
	s.unexpectedCall("UpdateGroup")
	return domain.Group{}, nil
}

func (s *testStore) DeleteGroup(context.Context, uuid.UUID) error {
	s.unexpectedCall("DeleteGroup")
	return nil
}

func (s *testStore) ListMemberships(
	context.Context,
	domain.MembershipListOptions,
) ([]domain.Membership, int32, error) {
	if s.listMembersErr != nil {
		return nil, 0, s.listMembersErr
	}

	return s.memberships, int32(len(s.memberships)), nil
}

func (s *testStore) CreateMembership(
	context.Context,
	uuid.UUID,
	domain.MemberKind,
	uuid.UUID,
	domain.MembershipOrigin,
) (domain.Membership, error) {
	s.createMembershipCalls++
	return domain.Membership{}, nil
}

func (s *testStore) DeleteMembership(context.Context, uuid.UUID, domain.MemberKind) error {
	s.deleteMembershipCalls++
	return nil
}

func (s *testStore) UpdateMachineDesiredTargets(_ context.Context, machineID uuid.UUID) error {
	s.syncedMachineIDs = append(s.syncedMachineIDs, machineID)
	return nil
}

func (s *testStore) UpdateMachineDesiredTargetsByPrimaryUserID(_ context.Context, userID uuid.UUID) error {
	s.syncedUserIDs = append(s.syncedUserIDs, userID)
	return nil
}

func (s *testStore) unexpectedCall(method string) {
	s.t.Helper()
	s.t.Fatalf("unexpected %s call", method)
}

func newTestService(t *testing.T, store *testStore) *groups.Service {
	t.Helper()
	store.t = t
	return groups.New(store)
}

func TestAddUser_CreatesMembershipAndSyncsPrimaryUserTargets(t *testing.T) {
	groupID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	store := &testStore{
		group: domain.Group{
			ID:     groupID,
			Source: domain.PrincipalSourceLocal,
		},
	}
	service := newTestService(t, store)

	if err := service.AddUser(context.Background(), groupID, userID); err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	if store.createMembershipCalls != 1 {
		t.Fatalf("createMembershipCalls = %d, want 1", store.createMembershipCalls)
	}
	if store.deleteMembershipCalls != 0 {
		t.Fatalf("deleteMembershipCalls = %d, want 0", store.deleteMembershipCalls)
	}
	if len(store.syncedUserIDs) != 1 || store.syncedUserIDs[0] != userID {
		t.Fatalf("syncedUserIDs = %v, want [%v]", store.syncedUserIDs, userID)
	}
}

func TestAddUser_IsIdempotentWhenMembershipExists(t *testing.T) {
	groupID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	store := &testStore{
		group: domain.Group{
			ID:     groupID,
			Source: domain.PrincipalSourceLocal,
		},
		memberships: []domain.Membership{
			{
				ID: uuid.MustParse("00000000-0000-0000-0000-000000000005"),
				Member: domain.MembershipMember{
					Kind: domain.MemberKindUser,
					ID:   userID,
				},
			},
		},
	}
	service := newTestService(t, store)

	if err := service.AddUser(context.Background(), groupID, userID); err != nil {
		t.Fatalf("AddUser() error = %v", err)
	}

	if store.createMembershipCalls != 0 {
		t.Fatalf("createMembershipCalls = %d, want 0", store.createMembershipCalls)
	}
	if len(store.syncedUserIDs) != 0 {
		t.Fatalf("syncedUserIDs = %v, want empty", store.syncedUserIDs)
	}
}

func TestRemoveMachine_IsIdempotentWhenMembershipMissing(t *testing.T) {
	groupID := uuid.MustParse("00000000-0000-0000-0000-000000000006")
	machineID := uuid.MustParse("00000000-0000-0000-0000-000000000007")

	store := &testStore{
		group: domain.Group{
			ID:     groupID,
			Source: domain.PrincipalSourceLocal,
		},
	}
	service := newTestService(t, store)

	if err := service.RemoveMachine(context.Background(), groupID, machineID); err != nil {
		t.Fatalf("RemoveMachine() error = %v", err)
	}

	if store.deleteMembershipCalls != 0 {
		t.Fatalf("deleteMembershipCalls = %d, want 0", store.deleteMembershipCalls)
	}
	if len(store.syncedMachineIDs) != 0 {
		t.Fatalf("syncedMachineIDs = %v, want empty", store.syncedMachineIDs)
	}
}

func TestAddMachine_RejectsReadOnlyGroup(t *testing.T) {
	groupID := uuid.MustParse("00000000-0000-0000-0000-000000000008")
	machineID := uuid.MustParse("00000000-0000-0000-0000-000000000009")

	store := &testStore{
		group: domain.Group{
			ID:     groupID,
			Source: domain.PrincipalSourceEntra,
		},
	}
	service := newTestService(t, store)

	err := service.AddMachine(context.Background(), groupID, machineID)
	if !errors.Is(err, domain.ErrGroupReadOnly) {
		t.Fatalf("AddMachine() error = %v, want %v", err, domain.ErrGroupReadOnly)
	}
}
