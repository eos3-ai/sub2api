//go:build unit

package service

import (
	"context"
	"errors"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type internalBalanceOrderRecorderStub struct {
	calls []InternalBalanceOrderInput
}

func (s *internalBalanceOrderRecorderStub) RecordInternalBalanceOrder(ctx context.Context, input InternalBalanceOrderInput) (*dbent.PaymentOrder, error) {
	s.calls = append(s.calls, input)
	return nil, nil
}

func TestAdminService_CreateUser_Success(t *testing.T) {
	repo := &userRepoStub{nextID: 10}
	svc := &adminServiceImpl{userRepo: repo}
	balance := 12.5

	input := &CreateUserInput{
		Email:         "user@test.com",
		Password:      "strong-pass",
		Username:      "tester",
		Notes:         "note",
		Balance:       &balance,
		Concurrency:   7,
		AllowedGroups: []int64{3, 5},
	}

	user, err := svc.CreateUser(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, int64(10), user.ID)
	require.Equal(t, input.Email, user.Email)
	require.Equal(t, input.Username, user.Username)
	require.Equal(t, input.Notes, user.Notes)
	require.Equal(t, balance, user.Balance)
	require.Equal(t, input.Concurrency, user.Concurrency)
	require.Equal(t, input.AllowedGroups, user.AllowedGroups)
	require.Equal(t, RoleUser, user.Role)
	require.Equal(t, StatusActive, user.Status)
	require.True(t, user.CheckPassword(input.Password))
	require.Len(t, repo.created, 1)
	require.Equal(t, user, repo.created[0])
}

func TestAdminService_CreateUser_RecordsInitialBalanceOrder(t *testing.T) {
	repo := &userRepoStub{nextID: 10}
	redeemRepo := &balanceRedeemRepoStub{redeemRepoStub: &redeemRepoStub{}}
	recorder := &internalBalanceOrderRecorderStub{}
	svc := &adminServiceImpl{
		userRepo:              repo,
		redeemCodeRepo:        redeemRepo,
		internalOrderRecorder: recorder,
	}
	balance := 12.5

	input := &CreateUserInput{
		Email:    "balance@test.com",
		Password: "strong-pass",
		Balance:  &balance,
		Notes:    "initial grant",
	}

	user, err := svc.CreateUser(context.Background(), input)
	require.NoError(t, err)
	require.NotNil(t, user)
	require.Len(t, recorder.calls, 1)
	call := recorder.calls[0]
	require.Equal(t, user.ID, call.UserID)
	require.Equal(t, balance, call.Amount)
	require.Equal(t, PaymentTypeAdminAdjustment, call.SourceType)
	require.Equal(t, PaymentTypeAdminAdjustment, call.PaymentType)
	require.Equal(t, "admin_user_create:10", call.SourceRef)
	require.Equal(t, input.Notes, call.Notes)
	require.Equal(t, "admin", call.Operator)
	require.NotNil(t, call.CreatedAt)
	require.Equal(t, "create_user", call.Metadata["operation"])
	require.Equal(t, 0, call.Metadata["old_balance"])
	require.Equal(t, balance, call.Metadata["new_balance"])
	require.NotEmpty(t, call.Metadata["history_code"])

	require.Len(t, redeemRepo.created, 1)
	history := redeemRepo.created[0]
	require.Equal(t, AdjustmentTypeAdminBalance, history.Type)
	require.Equal(t, balance, history.Value)
	require.Equal(t, StatusUsed, history.Status)
	require.NotNil(t, history.UsedBy)
	require.Equal(t, user.ID, *history.UsedBy)
	require.Equal(t, input.Notes, history.Notes)
	require.NotNil(t, history.UsedAt)
}

func TestAdminService_CreateUser_UsesDefaultBalanceWhenBalanceOmitted(t *testing.T) {
	repo := &userRepoStub{nextID: 11}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserBalance: 0,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultBalance: "0.02",
	}}, cfg)
	svc := &adminServiceImpl{userRepo: repo, settingService: settingService}

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "default-balance@test.com",
		Password: "strong-pass",
	})

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, 0.02, user.Balance)
	require.Len(t, repo.created, 1)
	require.Equal(t, 0.02, repo.created[0].Balance)
}

func TestAdminService_CreateUser_ExplicitZeroBalanceOverridesDefault(t *testing.T) {
	repo := &userRepoStub{nextID: 12}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserBalance: 0,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultBalance: "0.02",
	}}, cfg)
	svc := &adminServiceImpl{userRepo: repo, settingService: settingService}
	balance := 0.0

	user, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "zero-balance@test.com",
		Password: "strong-pass",
		Balance:  &balance,
	})

	require.NoError(t, err)
	require.NotNil(t, user)
	require.Equal(t, 0.0, user.Balance)
	require.Len(t, repo.created, 1)
	require.Equal(t, 0.0, repo.created[0].Balance)
}

func TestAdminService_CreateUser_EmailExists(t *testing.T) {
	repo := &userRepoStub{createErr: ErrEmailExists}
	svc := &adminServiceImpl{userRepo: repo}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "dup@test.com",
		Password: "password",
	})
	require.ErrorIs(t, err, ErrEmailExists)
	require.Empty(t, repo.created)
}

func TestAdminService_CreateUser_CreateError(t *testing.T) {
	createErr := errors.New("db down")
	repo := &userRepoStub{createErr: createErr}
	svc := &adminServiceImpl{userRepo: repo}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "user@test.com",
		Password: "password",
	})
	require.ErrorIs(t, err, createErr)
	require.Empty(t, repo.created)
}

func TestAdminService_CreateUser_AssignsDefaultSubscriptions(t *testing.T) {
	repo := &userRepoStub{nextID: 21}
	assigner := &defaultSubscriptionAssignerStub{}
	cfg := &config.Config{
		Default: config.DefaultConfig{
			UserBalance:     0,
			UserConcurrency: 1,
		},
	}
	settingService := NewSettingService(&settingRepoStub{values: map[string]string{
		SettingKeyDefaultSubscriptions: `[{"group_id":5,"validity_days":30}]`,
	}}, cfg)
	svc := &adminServiceImpl{
		userRepo:           repo,
		settingService:     settingService,
		defaultSubAssigner: assigner,
	}

	_, err := svc.CreateUser(context.Background(), &CreateUserInput{
		Email:    "new-user@test.com",
		Password: "password",
	})
	require.NoError(t, err)
	require.Len(t, assigner.calls, 1)
	require.Equal(t, int64(21), assigner.calls[0].UserID)
	require.Equal(t, int64(5), assigner.calls[0].GroupID)
	require.Equal(t, 30, assigner.calls[0].ValidityDays)
}
