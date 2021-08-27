package dao

import (
	"database/sql"

	"github.com/ziidu/guid/model"

	_ "github.com/go-sql-driver/mysql"
)

// GuidAllocDao operate database
type GuidAllocDao interface {
	// GetAllSegmentMetadatas query all metadata from db
	GetAllSegmentMetadatas() ([]model.SegmentMetadata, error)

	// UpdateMaxIdAndGet Update the max_id field, and select updated result
	// update and select opertaor must be in one Transiction
	UpdateMaxIdAndGet(bizTag string) (model.SegmentMetadata, error)

	// GetAndUpdateMaxId get SegmentMetadata and update MaxId
	GetAndUpdateMaxId(bizTag string) (model.SegmentMetadata, error)

	// GetById select metadata by bizTag
	GetById(bizTag string) (model.SegmentMetadata, error)
}

var _ (GuidAllocDao) = (*DefaultDao)(nil)

// DefaultDao is a implement of GuidAllocDao dependy on mysql database
type DefaultDao struct {
	db        *sql.DB
	tableName string
}

func NewDefaultDao(connectURL, tableName string) *DefaultDao {
	var (
		dao = &DefaultDao{tableName: tableName}
		err error
	)
	dao.db, err = sql.Open("mysql", connectURL)
	if err != nil {
		panic(err)
	}
	return dao
}

func (dao *DefaultDao) GetAllSegmentMetadatas() ([]model.SegmentMetadata, error) {
	rows, err := dao.db.Query("SELECT biz_tag, max_id, step, description, update_time FROM " + dao.tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var metadatas []model.SegmentMetadata = make([]model.SegmentMetadata, 0)
	for rows.Next() {
		var metadata model.SegmentMetadata
		if err := rows.Scan(&metadata.BizTag, &metadata.MaxId, &metadata.Step,
			&metadata.Description, &metadata.UpdateTime); err != nil {
			return nil, err
		}
		metadatas = append(metadatas, metadata)
	}
	return metadatas, nil
}

func (dao *DefaultDao) GetAndUpdateMaxId(bizTag string) (metadata model.SegmentMetadata, err error) {
	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()
	row := tx.QueryRow(`SELECT biz_tag, max_id, step, description, update_time FROM `+dao.tableName+` where biz_tag = ?`, bizTag)
	if err = row.Err(); err != nil {
		return
	}
	if err = row.Scan(&metadata.BizTag, &metadata.MaxId, &metadata.Step,
		&metadata.Description, &metadata.UpdateTime); err != nil {
		return
	}
	_, err = tx.Exec(`UPDATE `+dao.tableName+` SET max_id = max_id + step WHERE biz_tag = ?`, bizTag)
	if err != nil {
		return
	}
	return
}

func (dao *DefaultDao) UpdateMaxIdAndGet(bizTag string) (metadata model.SegmentMetadata, err error) {
	tx, err := dao.db.Begin()
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		tx.Commit()
	}()
	_, err = tx.Exec(`UPDATE `+dao.tableName+` SET max_id = max_id + step WHERE biz_tag = ?`, bizTag)
	if err != nil {
		return
	}
	row := tx.QueryRow(`SELECT biz_tag, max_id, step, description, update_time FROM `+dao.tableName+` where biz_tag = ?`, bizTag)
	if err = row.Err(); err != nil {
		return
	}
	if err = row.Scan(&metadata.BizTag, &metadata.MaxId, &metadata.Step,
		&metadata.Description, &metadata.UpdateTime); err != nil {
		return
	}
	return
}

func (dao *DefaultDao) GetById(bizTag string) (metadata model.SegmentMetadata, err error) {
	row := dao.db.QueryRow(`SELECT biz_tag, max_id, step, description, update_time FROM `+dao.tableName+` where biz_tag = ?`, bizTag)
	if err = row.Err(); err != nil {
		return
	}
	if err = row.Scan(&metadata.BizTag, &metadata.MaxId, &metadata.Step,
		&metadata.Description, &metadata.UpdateTime); err != nil {
		return
	}
	return
}
