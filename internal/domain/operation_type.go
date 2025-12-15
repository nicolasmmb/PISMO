package domain

// OperationType represents transaction type and its sign.
type OperationType struct {
	ID          int
	Description string
	Sign        int // -1 or +1
}

const (
	OperationTypeNormalPurchase      = 1
	OperationTypePurchaseInstallment = 2
	OperationTypeWithdrawal          = 3
	OperationTypeCreditVoucher       = 4
)
