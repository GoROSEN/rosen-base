package account

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/martian/log"
	"gorm.io/gorm"
)

// AccountService 资金账户服务
type AccountService struct {
	db *gorm.DB
}

// NewAccountService 新建账户服务
func NewAccountService(_db *gorm.DB) *AccountService {

	return &AccountService{db: _db}
}

// Begin 创建一个包含事物的账户服务
func (s *AccountService) Database() *gorm.DB {
	return s.db
}

func (s *AccountService) Begin() *AccountService {
	db := s.db.Begin()
	return NewAccountService(db)
}

// Commit 提交事务并废止此服务
func (s *AccountService) Commit() {
	s.db.Commit()
	s.db = nil
}

// Rollback 回滚事务并废止此服务
func (s *AccountService) Rollback() {
	s.db.Rollback()
	s.db = nil
}

// NewAccount 根据用户ID创建一个账户
func (s *AccountService) NewAccount(userID uint, userName string, atype string) *Account {

	a := &Account{Type: atype, UserID: userID, UserName: userName}
	s.db.Create(a)
	return a
}

// Account 根据用户ID和账户类型返回一个账户，若该账户不存在，则创建一个新的
func (s *AccountService) Account(userID uint, atype string) *Account {
	a := s.GetAccountByUserAndType(userID, atype)
	if a == nil {
		a = s.NewAccount(userID, "", atype)
	}
	return a
}

// SetAccountUserName 设置用户名称
func (s *AccountService) SetAccountUserName(accountID uint, userName string) error {
	if accountID == 0 {
		return errors.New("invalid account id")
	}

	a := &Account{}
	a.ID = accountID

	return s.db.Model(a).Update("user_name", userName).Error
}

// GetAccountByID 根据账户ID获取账户
func (s *AccountService) GetAccountByID(accountID uint) *Account {

	a := Account{}
	err := s.db.First(&a, accountID).Error

	if err != nil {
		return nil
	}
	return &a
}

// GetAccountByUserAndType 根据用户ID和账户类型获取账户
func (s *AccountService) GetAccountByUserAndType(userID uint, atype string) *Account {

	a := Account{}
	err := s.db.First(&a, "user_id = ? AND type = ?", userID, atype).Error

	if err != nil {
		return nil
	}
	return &a
}

// IncreaseAvailable 增加账户可用余额
func (s *AccountService) IncreaseAvailable(to *Account, value int64, description string) (*Transaction, error) {

	s.db.Model(to).Update("available", gorm.Expr("available + ?", value))
	to.Available += value
	return s.newTransaction(0, to.ID, value, "IncreaseAvailable", description), nil
}

// DecreaseAvailable 减少账户可用余额
func (s *AccountService) DecreaseAvailable(to *Account, value int64, description string) (*Transaction, error) {

	tx := s.db.Model(to).Where("available >= ?", value).Update("available", gorm.Expr("available - ?", value))
	if tx.Error != nil || tx.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient available or err: %v", tx.Error)
	}
	to.Available += value
	return s.newTransaction(0, to.ID, value, "DecreaseAvailable", description), nil
}

// Transfer 可用资金转账
func (s *AccountService) Transfer(from *Account, to *Account, value int64, description string) (*Transaction, error) {

	if from.Available < value {
		return nil, errors.New("insufficient available")
	}
	if from.Type != to.Type {
		return nil, errors.New("inconsist type")
	}
	tx := s.db.Model(from).Where("available >= ?", value).Update("available", gorm.Expr("available - ?", value))
	if tx.Error != nil || tx.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient available or err: %v", tx.Error)
	}
	s.db.Model(to).Update("available", gorm.Expr("available + ?", value))
	from.Available -= value
	to.Available += value
	return s.newTransaction(from.ID, to.ID, value, "Transfer", description), nil
}

// Transfer 可用资金转账至目标账户冻结金额
func (s *AccountService) TransferToLock(from *Account, to *Account, value int64, description string) (*Transaction, error) {

	if from.Available < value {
		return nil, errors.New("insufficient available")
	}
	if from.Type != to.Type {
		return nil, errors.New("inconsist type")
	}
	tx := s.db.Model(from).Where("available >= ?", value).Update("available", gorm.Expr("available - ?", value))
	if tx.Error != nil || tx.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient available or err: %v", tx.Error)
	}
	s.db.Model(to).Update("locked", gorm.Expr("locked + ?", value))
	from.Available -= value
	to.Locked += value
	return s.newTransaction(from.ID, to.ID, value, "TransferToLock", description), nil
}

// Lock 锁定可用资金
func (s *AccountService) Lock(from *Account, value int64, description string) (*Transaction, error) {

	if from.Available < value {
		return nil, errors.New("insufficient available")
	}
	tx := s.db.Model(from).Where("available >= ?", value).Updates(map[string]interface{}{"available": gorm.Expr("available - ?", value), "locked": gorm.Expr("locked + ?", value)})
	if tx.Error != nil || tx.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient available or err: %v", tx.Error)
	}
	from.Available -= value
	from.Locked += value
	return s.newTransaction(from.ID, from.ID, value, "Lock", description), nil
}

// Unlock 解锁锁定资金
func (s *AccountService) Unlock(from *Account, value int64, description string) (*Transaction, error) {

	if from.Locked < value {
		return nil, errors.New("insufficient locked")
	}
	s.db.Model(from).Updates(map[string]interface{}{"available": gorm.Expr("available + ?", value), "locked": gorm.Expr("locked - ?", value)})
	from.Available += value
	from.Locked -= value
	return s.newTransaction(from.ID, from.ID, value, "Unlock", description), nil
}

// IncreaseLocked 增加锁定金额
func (s *AccountService) IncreaseLocked(to *Account, value int64, description string) (*Transaction, error) {

	s.db.Model(to).Update("locked", gorm.Expr("locked + ?", value))
	to.Locked += value
	return s.newTransaction(0, to.ID, value, "IncreaseLocked", description), nil
}

// DecreaseLocked 减少锁定金额
func (s *AccountService) DecreaseLocked(from *Account, value int64, description string) (*Transaction, error) {

	s.db.Model(from).Update("locked", gorm.Expr("locked - ?", value))
	from.Locked -= value
	return s.newTransaction(from.ID, 0, value, "DecreaseLocked", description), nil
}

// Freeze 冻结可用余额
func (s *AccountService) Freeze(from *Account, value int64, description string) (*Transaction, error) {

	if from.Available < value {
		return nil, errors.New("insufficient available")
	}
	tx := s.db.Model(from).Where("available >= ?", value).Updates(map[string]interface{}{"available": gorm.Expr("available - ?", value), "frozen": gorm.Expr("frozen + ?", value)})
	if tx.Error != nil || tx.RowsAffected == 0 {
		return nil, fmt.Errorf("insufficient available or err: %v", tx.Error)
	}
	from.Available -= value
	from.Frozen += value
	return s.newTransaction(from.ID, from.ID, value, "Freeze", description), nil
}

// Unfreeze 解冻可用余额
func (s *AccountService) Unfreeze(from *Account, to *Account, value int64, description string) (*Transaction, error) {

	if from.Frozen < value {
		return nil, errors.New("insufficient frozen")
	}
	if from.Type != to.Type {
		return nil, errors.New("inconsist type")
	}
	s.db.Model(from).Update("frozen", gorm.Expr("frozen - ?", value))
	s.db.Model(to).Update("available", gorm.Expr("available + ?", value))
	from.Frozen -= value
	to.Available += value
	return s.newTransaction(from.ID, to.ID, value, "Unfreeze", description), nil
}

// DecreaseFrozenLocked 减少冻结金额
func (s *AccountService) DecreaseFrozen(from *Account, value int64, description string) (*Transaction, error) {

	s.db.Model(from).Update("frozen", gorm.Expr("frozen - ?", value))
	from.Locked -= value
	return s.newTransaction(from.ID, 0, value, "DecreaseFrozen", description), nil
}

// ListAll 分页列出所有内容，返回内容条目数组与文件总数
func (s *AccountService) ListAll(page, pageSize int) ([]Account, int64) {

	var contents []Account
	s.db.Offset(page * pageSize).Limit(pageSize).Order("id").Find(&contents)
	var count int64
	s.db.Model(&Account{}).Count(&count)
	return contents, count
}

func (s *AccountService) ListTransaction(accountID uint, page, pageSize int) ([]*TransactionVO, int64, error) {

	var contents []*TransactionVO
	rawSql := `select t.*,a0.user_name as from_account_name,a1.user_name as to_account_name 
			from transactions as t 
				left join accounts as a0 on from_account_id = a0.id 
				left join accounts as a1 on to_account_id = a1.id `
	if accountID > 0 {
		rawSql += "where from_account_id = ? or to_account_id = ?"
	}
	rawSql += "order by t.created_at desc"
	pagedSql := " limit ? offset ?"
	var rows *sql.Rows
	var err error
	if accountID > 0 {
		rows, err = s.db.Raw(rawSql+pagedSql, accountID, accountID, pageSize, page*pageSize).Rows()
	} else {
		rows, err = s.db.Raw(rawSql+pagedSql, pageSize, page*pageSize).Rows()
	}
	if err != nil {
		log.Debugf("listTransaction error 1: %v", err)
		return nil, 0, err
	}
	defer rows.Close()
	contents = make([]*TransactionVO, 0)
	for rows.Next() {
		t := &TransactionVO{}
		s.db.ScanRows(rows, t)
		contents = append(contents, t)
	}
	var count int64
	if accountID > 0 {
		rawSql = "select count(1) from transactions where from_account_id = ? or to_account_id = ?"
	} else {
		rawSql = "select count(1) from transactions"
	}
	if err := s.db.Raw(rawSql, accountID, accountID).Count(&count).Error; err != nil {
		log.Debugf("listTransaction error 2: %v", err)
		return nil, 0, err
	}

	return contents, count, nil
}

func (s *AccountService) ListTransaction2(accountName, accountType, operation string, page, pageSize int) ([]*TransactionVO, int64, error) {

	var contents []*TransactionVO
	params := []interface{}{}
	rawSql := `select t.*,a0.user_name as from_account_name,a1.user_name as to_account_name 
			from transactions as t 
				left join accounts as a0 on from_account_id = a0.id 
				left join accounts as a1 on to_account_id = a1.id `
	// where
	whereSql := ""
	if len(accountName) > 0 {
		whereSql += "where (a0.user_name = ? or a1.user_name = ?) "
		params = append(params, accountName, accountName)
	}
	if len(accountType) > 0 {
		if len(whereSql) > 0 {
			whereSql += "and (a1.type = ? or a0.type = ?)"
		} else {
			whereSql += "where (a1.type = ? or a0.type = ?)"
		}
		params = append(params, accountType, accountType)
	}
	if len(operation) > 0 {
		if len(whereSql) > 0 {
			whereSql += "and t.operation = ? "
		} else {
			whereSql += "where t.operation = ? "
		}
		params = append(params, operation)
	}
	rawSql += whereSql
	rawSql += "order by t.created_at desc"
	log.Debugf(rawSql)
	// paging
	pagedSql := " limit ? offset ?"
	pagedParams := append(params, pageSize, page*pageSize)
	// paged query
	rows, err := s.db.Raw(rawSql+pagedSql, pagedParams...).Rows()
	if err != nil {
		log.Debugf("listTransaction2 error 1: %v", err)
		return nil, 0, err
	}
	defer rows.Close()
	contents = make([]*TransactionVO, 0)
	for rows.Next() {
		t := &TransactionVO{}
		s.db.ScanRows(rows, t)
		contents = append(contents, t)
	}
	// calculate count
	rawSql = `select count(*)
						from transactions as t 
							left join accounts as a0 on from_account_id = a0.id 
							left join accounts as a1 on to_account_id = a1.id `
	var count int64
	if err := s.db.Raw(rawSql+whereSql, params...).Count(&count).Error; err != nil {
		log.Debugf("listTransaction2 error 2: %v", err)
		return nil, 0, err
	}

	return contents, count, nil
}

func (s *AccountService) ListAllTransaction2(accountName, accountType, operation string) ([]*TransactionVO, error) {

	var contents []*TransactionVO
	params := []interface{}{}
	rawSql := `select t.*,a0.user_name as from_account_name,a1.user_name as to_account_name 
			from transactions as t 
				left join accounts as a0 on from_account_id = a0.id 
				left join accounts as a1 on to_account_id = a1.id `
	// where
	whereSql := ""
	if len(accountName) > 0 {
		whereSql += "where (a0.user_name = ? or a1.user_name = ?) "
		params = append(params, accountName, accountName)
	}
	if len(accountType) > 0 {
		if len(whereSql) > 0 {
			whereSql += "and (a1.type = ? or a0.type = ?)"
		} else {
			whereSql += "where (a1.type = ? or a0.type = ?)"
		}
		params = append(params, accountType, accountType)
	}
	if len(operation) > 0 {
		if len(whereSql) > 0 {
			whereSql += "and t.operation = ? "
		} else {
			whereSql += "where t.operation = ? "
		}
		params = append(params, operation)
	}
	rawSql += whereSql
	rawSql += "order by t.id asc"
	log.Debugf(rawSql)
	// paged query
	rows, err := s.db.Raw(rawSql, params...).Rows()
	if err != nil {
		log.Debugf("ListAllTransaction2 error 1: %v", err)
		return nil, err
	}
	defer rows.Close()
	contents = make([]*TransactionVO, 0)
	for rows.Next() {
		t := &TransactionVO{}
		s.db.ScanRows(rows, t)
		contents = append(contents, t)
	}

	return contents, nil
}

func (s *AccountService) SaveReceiption(channel, receptionId string, userId uint, rawContent interface{}) error {
	jsonStr, err := json.Marshal(rawContent)
	if err != nil {
		return err
	}
	reception := Receiption{
		ID:         fmt.Sprintf("%v:%v", channel, receptionId),
		CreatedAt:  time.Now(),
		OwnerID:    userId,
		RawContent: string(jsonStr),
	}
	return s.db.Create(&reception).Error
}

func (s *AccountService) RevokeReceiption(channel, receptionId string, rawContent interface{}) error {
	jsonStr, err := json.Marshal(rawContent)
	if err != nil {
		return err
	}
	var reception Receiption
	if err := s.db.First(&reception, "id = ?", fmt.Sprintf("%v:%v", channel, receptionId)).Error; err != nil {
		return err
	}
	if len(reception.Revoke) > 0 {
		return errors.New("already revoked")
	}
	// reception.Revoke = string(jsonStr)
	// reception.RevokeAt = time.Now()
	if err := s.db.Model(&reception).Updates(map[string]interface{}{"revoke": string(jsonStr), "revoke_at": time.Now()}).Error; err != nil {
		return err
	}

	return nil
}

// -- 私有方法

func (s *AccountService) newTransaction(fromAccountID, toAccountID uint, value int64, operation, description string) *Transaction {

	t := &Transaction{FromAccountID: fromAccountID, ToAccountID: toAccountID, Operation: operation, Value: value, Description: description}
	s.db.Create(t)
	return t
}
