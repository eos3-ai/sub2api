package service

import "context"

// BonusCalculator 计算赠送金额的接口
type BonusCalculator interface {
	// Calculate 计算赠送金额
	Calculate(ctx context.Context, userID int64, amountUSD float64) (*CalculationResult, error)
}

// CalculationResult 赠送计算结果
type CalculationResult struct {
	ShouldGrant bool    // 是否应该赠送
	Amount      float64 // 赠送金额
	Percent     float64 // 赠送比例
	Tier        int     // 阶梯等级
	Reason      string  // 原因说明
}

// BonusIssuer 发放赠送的接口
type BonusIssuer interface {
	// Issue 发放赠送到用户余额
	Issue(ctx context.Context, req *IssueRequest) error
}

// IssueRequest 发放请求
type IssueRequest struct {
	UserID          int64
	Amount          float64
	Type            string  // "promotion", "referral", etc.
	Remark          string
	RelatedID       *string // 关联订单号
	ActivityOrderNo *string // 活动订单号
}

// BonusRecorder 记录赠送使用情况的接口
type BonusRecorder interface {
	// Record 记录赠送使用情况
	Record(ctx context.Context, req *RecordRequest) error
}

// RecordRequest 记录请求
type RecordRequest struct {
	UserID    int64
	Tier      int
	AmountUSD float64
	BonusUSD  float64
}

// BonusRequest 赠送处理请求
type BonusRequest struct {
	OrderID   int64
	OrderNo   string
	UserID    int64
	Username  string
	AmountUSD float64
	AmountCNY float64
	Provider  string // "zpay", "stripe"
}

// BonusResult 赠送处理结果
type BonusResult struct {
	Applied      bool
	BonusAmount  float64
	Tier         int
	BonusPercent float64
	RecordID     *int64 // 赠送记录ID
	Error        error  // 发放失败时的错误
}
