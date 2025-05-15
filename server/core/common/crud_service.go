package common

import (
	"gorm.io/gorm"
)

// CrudService CRUD服务
type CrudService struct {
	Db *gorm.DB
}

// NewCrudService 新建CRUD服务
func NewCrudService(_db *gorm.DB) *CrudService {
	return &CrudService{Db: _db}
}

// SetDB 设置DB
func (s *CrudService) SetDB(_db *gorm.DB) {
	s.Db = _db
}

// PagedList 分页列表
func (s *CrudService) PagedList(content interface{}, model interface{}, page, pageSize int) (int64, error) {

	return s.Paged(content, model, nil, "id", page, pageSize, "")
}

// PagedListOrder 分页列表
func (s *CrudService) PagedListOrder(content interface{}, model interface{}, order string, page, pageSize int) (int64, error) {

	return s.Paged(content, model, nil, order, page, pageSize, "")
}

// PagedListPreload 分页列表
func (s *CrudService) PagedListPreload(content interface{}, model interface{}, preloads []string, page, pageSize int) (int64, error) {

	return s.Paged(content, model, preloads, "id", page, pageSize, "")
}

// PagedListWhere 分页列表
func (s *CrudService) PagedListWhere(content interface{}, model interface{}, page, pageSize int, where string, params ...interface{}) (int64, error) {

	return s.Paged(content, model, nil, "id", page, pageSize, where, params...)
}

// ListAll 列出全部
func (s *CrudService) ListAll(content interface{}) error {

	return s.List(content, nil, "id", "")
}

// ListAllPreload 列出全部
func (s *CrudService) ListAllPreload(content interface{}, preloads []string) error {

	return s.List(content, preloads, "id", "")
}

// ListAllWhere 条件查询
func (s *CrudService) ListAllWhere(content interface{}, where string, params ...interface{}) error {

	return s.List(content, nil, "id", where, params...)
}

// Paged 分页列表
func (s *CrudService) PagedV2(content interface{}, model interface{}, preloads []string, joins []string, order string, page, pageSize int, where string, params ...interface{}) (int64, error) {

	return s.PagedV3(content, model, preloads, joins, nil, order, page, pageSize, where, params...)
}

// Paged 分页列表
func (s *CrudService) PagedV3(content interface{}, model interface{}, preloads []string, joins []string, innerJoins []string, order string, page, pageSize int, where string, params ...interface{}) (int64, error) {

	t := s.Db
	t0 := s.Db.Model(model)
	if pageSize > 0 {
		t = t.Offset(page * pageSize).Limit(pageSize)
	}
	for _, obj := range preloads {
		t = t.Preload(obj)
		t0 = t0.Preload(obj)
	}
	for _, obj := range joins {
		t = t.Joins(obj)
		t0 = t0.Joins(obj)
	}
	for _, obj := range innerJoins {
		t = t.InnerJoins(obj)
		t0 = t0.InnerJoins(obj)
	}
	if len(where) > 0 {
		t = t.Where(where, params...)
		t0 = t0.Where(where, params...)
	}
	if len(order) > 0 {
		t = t.Order(order)
	}
	err := t.Find(content).Error
	var count int64
	t0.Count(&count)
	return count, err
}

func (s *CrudService) Paged(content interface{}, model interface{}, preloads []string, order string, page, pageSize int, where string, params ...interface{}) (int64, error) {

	return s.PagedV2(content, model, preloads, nil, order, page, pageSize, where, params...)
}

// List 列出全部
func (s *CrudService) ListV3(content interface{}, limit int, preloads []string, joins []string, order string, where string, params ...interface{}) error {

	t := s.Db
	for _, obj := range preloads {
		t = t.Preload(obj)
	}
	for _, obj := range joins {
		t = t.Joins(obj)
	}
	if limit > 0 {
		t.Limit(limit)
	}
	if len(where) > 0 {
		t = t.Where(where, params...)
	}
	if len(order) > 0 {
		t = t.Order(order)
	}
	return t.Find(content).Error
}

func (s *CrudService) ListV2(content interface{}, preloads []string, joins []string, order string, where string, params ...interface{}) error {

	return s.ListV2_1(content, preloads, joins, nil, order, where, params...)
}

func (s *CrudService) ListV2_1(content interface{}, preloads []string, joins []string, innerJoins []string, order string, where string, params ...interface{}) error {

	t := s.Db
	for _, obj := range preloads {
		t = t.Preload(obj)
	}
	for _, obj := range joins {
		t = t.Joins(obj)
	}
	for _, obj := range innerJoins {
		t = t.InnerJoins(obj)
	}
	if len(where) > 0 {
		t = t.Where(where, params...)
	}
	if len(order) > 0 {
		t = t.Order(order)
	}
	return t.Find(content).Error
}

func (s *CrudService) List(content interface{}, preloads []string, order string, where string, params ...interface{}) error {
	return s.ListV2(content, preloads, nil, order, where, params...)
}

// GetModelByID 根据ID查询模型
func (s *CrudService) GetModelByID(model interface{}, id uint) error {
	return s.Db.First(model, id).Error
}

// GetPreloadModelByID 根据ID查询模型
func (s *CrudService) GetPreloadModelByID(model interface{}, id uint, preloads []string) error {
	t := s.Db
	for _, obj := range preloads {
		t = t.Preload(obj)
	}
	return t.First(model, id).Error
}

// Count 根据条件查询模型总数
func (s *CrudService) Count(model interface{}, where string, params ...interface{}) int64 {

	t := s.Db
	if len(where) > 0 {
		t = t.Where(where, params...)
	}
	var count int64
	if err := t.Model(model).Where(where, params...).Count(&count).Error; err != nil {
		return 0
	}
	return count
}

// FindModelWhere 根据条件查询模型
func (s *CrudService) FindModelWhere(model interface{}, where string, params ...interface{}) error {

	t := s.Db
	if len(where) > 0 {
		t = t.Where(where, params...)
	}
	return t.First(model).Error
}

// FindPreloadModelWhere 根据条件查询模型
func (s *CrudService) FindPreloadModelWhere(model interface{}, preloads []string, where string, params ...interface{}) error {

	t := s.Db
	if preloads != nil {
		for _, obj := range preloads {
			t = t.Preload(obj)
		}
	}
	if len(where) > 0 {
		t = t.Where(where, params...)
	}
	return t.First(model).Error
}

// FindPreloadJoinModelWhere 根据条件查询模型
func (s *CrudService) FindPreloadJoinModelWhere(model interface{}, preloads []string, joins []string, where string, params ...interface{}) error {

	t := s.Db
	if preloads != nil {
		for _, obj := range preloads {
			t = t.Preload(obj)
		}
	}
	if joins != nil {
		for _, obj := range joins {
			t = t.Joins(obj)
		}
	}
	if len(where) > 0 {
		t = t.Where(where, params...)
	}
	return t.First(model).Error
}

// CreateModel 插入新记录
func (s *CrudService) CreateModel(model interface{}) error {

	return s.Db.Create(model).Error
}

// CreateModelOpt 插入新记录
func (s *CrudService) CreateModelOpt(model interface{}, selects []string, omits []string) error {

	return s.Db.Select(selects).Omit(omits...).Create(model).Error
}

// SaveModel 保存记录
func (s *CrudService) SaveModel(model interface{}) error {

	return s.Db.Save(model).Error
}

func (s *CrudService) SaveModelOpt(model interface{}, selects []string, omits []string) error {

	db := s.Db
	if selects != nil {
		db = db.Select(selects)
	}
	if omits != nil {
		db = db.Omit(omits...)
	}
	return db.Save(model).Error
}

func (s *CrudService) UpdateModel(model interface{}, selects []string, omits []string) error {

	db := s.Db
	if selects != nil {
		db = db.Select(selects)
	}
	if omits != nil {
		db = db.Omit(omits...)
	}
	return db.Updates(model).Error
}

// DeleteModel 删除记录
func (s *CrudService) DeleteModel(model interface{}) error {
	return s.Db.Delete(model).Error
}
