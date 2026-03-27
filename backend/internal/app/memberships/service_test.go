package memberships_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/woodleighschool/grinch/internal/app/memberships"
	"github.com/woodleighschool/grinch/internal/domain"
)

type testStore struct {
	group               domain.Group
	getGroupErr         error
	getMembershipResult domain.Membership
	getMembershipErr    error

	createdMembership domain.Membership
	createCalls       int
	deleteCalls       int

	syncedMachineIDs []uuid.UUID
	syncedUserIDs    []uuid.UUID
}

func (s *testStore) ListMemberships(context.Context, domain.MembershipListOptions) ([]domain.Membership, int32, error) {
	return nil, 0, errors.New("unexpected ListMemberships call")
}

func (s *testStore) GetMembership(context.Context, uuid.UUID) (domain.Membership, error) {
	if s.getMembershipErr != nil {
		return domain.Membership{}, s.getMembershipErr
	}

	return s.getMembershipResult, nil
}

func (s *testStore) CreateMembership(
	context.Context,
	uuid.UUID,
	domain.MemberKind,
	uuid.UUID,
	domain.MembershipOrigin,
) (domain.Membership, error) {
	s.createCalls++
	return s.createdMembership, nil
}

func (s *testStore) DeleteMembership(context.Context, uuid.UUID, domain.MemberKind) error {
	s.deleteCalls++
	return nil
}

func (s *testStore) GetGroup(context.Context, uuid.UUID) (domain.Group, error) {
	if s.getGroupErr != nil {
		return domain.Group{}, s.getGroupErr
	}

	return s.group, nil
}

func (s *testStore) UpdateMachineDesiredTargets(_ context.Context, machineID uuid.UUID) error {
	s.syncedMachineIDs = append(s.syncedMachineIDs, machineID)
	return nil
}

func (s *testStore) UpdateMachineDesiredTargetsByPrimaryUserID(_ context.Context, userID uuid.UUID) error {
	s.syncedUserIDs = append(s.syncedUserIDs, userID)
	return nil
}

func newTestService(store *testStore) *memberships.Service {
	return memberships.New(store)
}

func TestCreateMembership_CreatesMembershipAndSyncsPrimaryUserTargets(t *testing.T) {
	groupID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000002")

	store := &testStore{
		group: domain.Group{
			ID:     groupID,
			Source: domain.PrincipalSourceLocal,
		},
		createdMembership: domain.Membership{
			ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"),
		},
	}

	service := newTestService(store)
	membership, err := service.CreateMembership(context.Background(), memberships.CreateInput{
		GroupID:    groupID,
		MemberKind: domain.MemberKindUser,
		MemberID:   userID,
	})
	if err != nil {
		t.Fatalf("CreateMembership() error = %v", err)
	}

	if membership.ID != store.createdMembership.ID {
		t.Fatalf("membership.ID = %v, want %v", membership.ID, store.createdMembership.ID)
	}
	if store.createCalls != 1 {
		t.Fatalf("createCalls = %d, want 1", store.createCalls)
	}
	if len(store.syncedUserIDs) != 1 || store.syncedUserIDs[0] != userID {
		t.Fatalf("syncedUserIDs = %v, want [%v]", store.syncedUserIDs, userID)
	}
}

func TestCreateMembership_RejectsReadOnlyGroup(t *testing.T) {
	groupID := uuid.MustParse("00000000-0000-0000-0000-000000000004")

	store := &testStore{
		group: domain.Group{
			ID:     groupID,
			Source: domain.PrincipalSourceEntra,
		},
	}

	service := newTestService(store)
	_, err := service.CreateMembership(context.Background(), memberships.CreateInput{
		GroupID:    groupID,
		MemberKind: domain.MemberKindMachine,
		MemberID:   uuid.MustParse("00000000-0000-0000-0000-000000000005"),
	})
	if !errors.Is(err, domain.ErrGroupReadOnly) {
		t.Fatalf("CreateMembership() error = %v, want %v", err, domain.ErrGroupReadOnly)
	}
}

func TestDeleteMembership_DeletesMembershipAndSyncsMachineTargets(t *testing.T) {
	membershipID := uuid.MustParse("00000000-0000-0000-0000-000000000006")
	machineID := uuid.MustParse("00000000-0000-0000-0000-000000000007")

	store := &testStore{
		getMembershipResult: domain.Membership{
			ID: membershipID,
			Group: domain.MembershipGroup{
				ID:     uuid.MustParse("00000000-0000-0000-0000-000000000008"),
				Source: domain.PrincipalSourceLocal,
			},
			Member: domain.MembershipMember{
				Kind: domain.MemberKindMachine,
				ID:   machineID,
			},
		},
	}

	service := newTestService(store)
	if err := service.DeleteMembership(context.Background(), membershipID); err != nil {
		t.Fatalf("DeleteMembership() error = %v", err)
	}

	if store.deleteCalls != 1 {
		t.Fatalf("deleteCalls = %d, want 1", store.deleteCalls)
	}
	if len(store.syncedMachineIDs) != 1 || store.syncedMachineIDs[0] != machineID {
		t.Fatalf("syncedMachineIDs = %v, want [%v]", store.syncedMachineIDs, machineID)
	}
}
