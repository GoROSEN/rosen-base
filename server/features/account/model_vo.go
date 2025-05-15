package account

import "time"

// AccountVO 视图模型
type AccountVO struct {
	ID        uint   `json:"id"`
	Type      string `json:"type"`      // 币种、类型
	Available int64  `json:"available"` // 可用金额
	Locked    int64  `json:"locked"`    // 锁定金额
	Frozen    int64  `json:"frozen"`    // 冻结金额
	UserID    uint   `json:"userId"`    // 所属用户ID
	UserName  string `json:"userName"`  // 用户名称
}

func account2AccountVO(account *Account) *AccountVO {
	return &AccountVO{
		ID:        account.ID,
		Type:      account.Type,
		Available: account.Available,
		Locked:    account.Locked,
		Frozen:    account.Frozen,
		UserID:    account.UserID,
		UserName:  account.UserName}
}

func accountVo2Account(vo *AccountVO) *Account {
	a := &Account{
		Type:      vo.Type,
		Available: vo.Available,
		Locked:    vo.Locked,
		Frozen:    vo.Frozen,
		UserID:    vo.UserID,
		UserName:  vo.UserName}
	a.ID = vo.ID
	return a
}

func accountList2AccountVOList(array []Account) []*AccountVO {
	vo := make([]*AccountVO, 0, len(array))
	for _, a := range array {
		item := account2AccountVO(&a)
		vo = append(vo, item)
	}
	return vo
}

// TransactionVO 转账记录
type TransactionVO struct {
	ID              uint      `json:"id"`
	CreatedAt       time.Time `json:"createdTime"`
	FromAccountID   uint      `json:"fromAccountId"`
	FromAccountName string    `json:"fromAccountName"`
	ToAccountID     uint      `json:"toAccountId"`
	ToAccountName   string    `json:"toAccountName"`
	Operation       string    `json:"operation"`
	Value           int64     `json:"amount"`
	Description     string    `json:"description"`
}
