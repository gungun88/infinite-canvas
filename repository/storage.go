package repository

import (
	"github.com/tigerowo/infinite-canvas/model"
	"gorm.io/gorm"
)

// SaveStorageObject 保存存储对象记录。
func SaveStorageObject(object model.StorageObject) (model.StorageObject, error) {
	db, err := DB()
	if err != nil {
		return model.StorageObject{}, err
	}
	return object, db.Save(&object).Error
}

// GetStorageObject 根据 ID 获取存储对象。
func GetStorageObject(id string) (model.StorageObject, error) {
	db, err := DB()
	if err != nil {
		return model.StorageObject{}, err
	}
	var object model.StorageObject
	err = db.First(&object, "id = ?", id).Error
	return object, err
}

// DeleteStorageObjectRecord 删除存储对象记录（软删除）。
func DeleteStorageObjectRecord(id string) error {
	db, err := DB()
	if err != nil {
		return err
	}
	return db.Delete(&model.StorageObject{}, "id = ?", id).Error
}

func ListStorageObjectsByProvider(providerID string) ([]model.StorageObject, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var objects []model.StorageObject
	query := db.Where("provider_id = ?", providerID)
	if err := query.Find(&objects).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return objects, nil
}

func ListStorageObjectsByOwner(ownerID string) ([]model.StorageObject, error) {
	db, err := DB()
	if err != nil {
		return nil, err
	}
	var objects []model.StorageObject
	if err := db.Where("created_by = ?", ownerID).Find(&objects).Error; err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return objects, nil
}
