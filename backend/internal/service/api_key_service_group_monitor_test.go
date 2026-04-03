package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type monitorGroupRepoStub struct {
	activeGroups   []Group
	listActiveErr  error
	listActiveCall int

	overviewByGroup    map[int64]*PublicGroupMonitorAggregate
	overviewErr        error
	overviewCall       int
	lastOverviewIDs    []int64
	lastOverviewNowUTC time.Time
}

func (s *monitorGroupRepoStub) Create(context.Context, *Group) error {
	panic("unexpected Create call")
}

func (s *monitorGroupRepoStub) GetByID(context.Context, int64) (*Group, error) {
	panic("unexpected GetByID call")
}

func (s *monitorGroupRepoStub) GetByIDLite(context.Context, int64) (*Group, error) {
	panic("unexpected GetByIDLite call")
}

func (s *monitorGroupRepoStub) Update(context.Context, *Group) error {
	panic("unexpected Update call")
}

func (s *monitorGroupRepoStub) Delete(context.Context, int64) error {
	panic("unexpected Delete call")
}

func (s *monitorGroupRepoStub) DeleteCascade(context.Context, int64) ([]int64, error) {
	panic("unexpected DeleteCascade call")
}

func (s *monitorGroupRepoStub) List(context.Context, pagination.PaginationParams) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected List call")
}

func (s *monitorGroupRepoStub) ListWithFilters(context.Context, pagination.PaginationParams, string, string, string, *bool) ([]Group, *pagination.PaginationResult, error) {
	panic("unexpected ListWithFilters call")
}

func (s *monitorGroupRepoStub) ListActive(context.Context) ([]Group, error) {
	s.listActiveCall++
	if s.listActiveErr != nil {
		return nil, s.listActiveErr
	}
	return append([]Group(nil), s.activeGroups...), nil
}

func (s *monitorGroupRepoStub) ListActiveByPlatform(context.Context, string) ([]Group, error) {
	panic("unexpected ListActiveByPlatform call")
}

func (s *monitorGroupRepoStub) ExistsByName(context.Context, string) (bool, error) {
	panic("unexpected ExistsByName call")
}

func (s *monitorGroupRepoStub) GetAccountCount(context.Context, int64) (int64, int64, error) {
	panic("unexpected GetAccountCount call")
}

func (s *monitorGroupRepoStub) DeleteAccountGroupsByGroupID(context.Context, int64) (int64, error) {
	panic("unexpected DeleteAccountGroupsByGroupID call")
}

func (s *monitorGroupRepoStub) GetAccountIDsByGroupIDs(context.Context, []int64) ([]int64, error) {
	panic("unexpected GetAccountIDsByGroupIDs call")
}

func (s *monitorGroupRepoStub) BindAccountsToGroup(context.Context, int64, []int64) error {
	panic("unexpected BindAccountsToGroup call")
}

func (s *monitorGroupRepoStub) UpdateSortOrders(context.Context, []GroupSortOrderUpdate) error {
	panic("unexpected UpdateSortOrders call")
}

func (s *monitorGroupRepoStub) GetPublicGroupMonitorOverview(
	_ context.Context,
	groupIDs []int64,
	now time.Time,
) (map[int64]*PublicGroupMonitorAggregate, error) {
	s.overviewCall++
	s.lastOverviewIDs = append([]int64(nil), groupIDs...)
	s.lastOverviewNowUTC = now.UTC()
	if s.overviewErr != nil {
		return nil, s.overviewErr
	}
	return s.overviewByGroup, nil
}

func TestAPIKeyService_GetPublicGroupMonitor_IncludesUsageAndPurchasableSubscription(t *testing.T) {
	price := 9.9
	zeroPrice := 0.0
	repo := &monitorGroupRepoStub{
		activeGroups: []Group{
			{ID: 10, Name: "Beta Usage", Platform: PlatformAnthropic, SortOrder: 2, SubscriptionType: SubscriptionTypeStandard},
			{ID: 20, Name: "Exclusive Usage", Platform: PlatformAnthropic, SortOrder: 1, IsExclusive: true, SubscriptionType: SubscriptionTypeStandard},
			{ID: 30, Name: "Hidden Sub", Platform: PlatformOpenAI, SortOrder: 1, SubscriptionType: SubscriptionTypeSubscription, UserPurchaseVisible: false, UserPurchasePriceUSD: &price},
			{ID: 40, Name: "Delta Sub", Platform: PlatformOpenAI, SortOrder: 1, SubscriptionType: SubscriptionTypeSubscription, UserPurchaseVisible: true, UserPurchasePriceUSD: &price},
			{ID: 50, Name: "Zero Sub", Platform: PlatformOpenAI, SortOrder: 1, SubscriptionType: SubscriptionTypeSubscription, UserPurchaseVisible: true, UserPurchasePriceUSD: &zeroPrice},
			{ID: 60, Name: "Alpha Usage", Platform: PlatformAnthropic, SortOrder: 1, SubscriptionType: SubscriptionTypeStandard},
			{ID: 70, Name: "No Price Sub", Platform: PlatformOpenAI, SortOrder: 1, SubscriptionType: SubscriptionTypeSubscription, UserPurchaseVisible: true},
		},
		overviewByGroup: map[int64]*PublicGroupMonitorAggregate{
			40: {CurrentStatus: "abnormal"},
			60: {CurrentStatus: "normal"},
		},
	}
	svc := &APIKeyService{groupRepo: repo}

	resp, err := svc.GetPublicGroupMonitor(context.Background(), 101)
	require.NoError(t, err)

	require.Equal(t, 1, repo.listActiveCall)
	require.Equal(t, 1, repo.overviewCall)
	require.ElementsMatch(t, []int64{10, 40, 60}, repo.lastOverviewIDs)
	require.False(t, repo.lastOverviewNowUTC.IsZero())

	require.Equal(t, 3, resp.PublicGroupNum)
	require.Len(t, resp.Items, 3)

	require.Equal(t, int64(60), resp.Items[0].GroupID)
	require.Equal(t, "Alpha Usage", resp.Items[0].GroupName)
	require.Equal(t, PublicGroupMonitorTypePublic, resp.Items[0].GroupType)
	require.Equal(t, "normal", resp.Items[0].CurrentStatus)

	require.Equal(t, int64(40), resp.Items[1].GroupID)
	require.Equal(t, "Delta Sub", resp.Items[1].GroupName)
	require.Equal(t, PublicGroupMonitorTypeSubscription, resp.Items[1].GroupType)
	require.Equal(t, "abnormal", resp.Items[1].CurrentStatus)

	require.Equal(t, int64(10), resp.Items[2].GroupID)
	require.Equal(t, "Beta Usage", resp.Items[2].GroupName)
	require.Equal(t, PublicGroupMonitorTypePublic, resp.Items[2].GroupType)
	require.Equal(t, "unknown", resp.Items[2].CurrentStatus)
}

func TestAPIKeyService_GetPublicGroupMonitor_NoMonitorableGroup(t *testing.T) {
	zeroPrice := 0.0
	repo := &monitorGroupRepoStub{
		activeGroups: []Group{
			{ID: 1, Name: "Exclusive Usage", IsExclusive: true, SubscriptionType: SubscriptionTypeStandard},
			{ID: 2, Name: "Not Visible Sub", SubscriptionType: SubscriptionTypeSubscription, UserPurchaseVisible: false, UserPurchasePriceUSD: &zeroPrice},
		},
	}
	svc := &APIKeyService{groupRepo: repo}

	resp, err := svc.GetPublicGroupMonitor(context.Background(), 202)
	require.NoError(t, err)

	require.Equal(t, 1, repo.listActiveCall)
	require.Equal(t, 0, repo.overviewCall)
	require.NotNil(t, resp)
	require.Equal(t, 0, resp.PublicGroupNum)
	require.Empty(t, resp.Items)
}
