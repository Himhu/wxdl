package service

import (
	"backend/internal/model"
	"backend/internal/repository"
	"encoding/json"
	"errors"
	"fmt"
)

var ErrInvalidJSON = errors.New("invalid JSON format")

type MiniProgramConfigService struct {
	repo *repository.MiniProgramConfigRepository
}

func NewMiniProgramConfigService(repo *repository.MiniProgramConfigRepository) *MiniProgramConfigService {
	return &MiniProgramConfigService{repo: repo}
}

// ListConfigs 查询配置列表
func (s *MiniProgramConfigService) ListConfigs(namespace, scopeType, scopeCode string) ([]model.MiniProgramConfigItem, error) {
	return s.repo.List(namespace, scopeType, scopeCode)
}

// GetConfig 获取单个配置
func (s *MiniProgramConfigService) GetConfig(id uint) (*model.MiniProgramConfigItem, error) {
	return s.repo.GetByID(id)
}

// UpdateConfig 更新配置（直接更新发布值）
func (s *MiniProgramConfigService) UpdateConfig(id uint, value string, updatedBy uint) error {
	// 验证JSON格式
	var temp interface{}
	if err := json.Unmarshal([]byte(value), &temp); err != nil {
		return ErrInvalidJSON
	}

	item, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	item.PublishedValue = &value
	item.UpdatedBy = &updatedBy
	return s.repo.Update(item)
}

// GetBootstrapConfig 获取小程序启动配置
func (s *MiniProgramConfigService) GetBootstrapConfig(scopeType, scopeCode string) (map[string]interface{}, error) {
	items, err := s.repo.GetPublishedConfigs(scopeType, scopeCode)
	if err != nil {
		return nil, err
	}

	config := make(map[string]interface{})
	for _, item := range items {
		if item.PublishedValue == nil {
			continue
		}

		if config[item.Namespace] == nil {
			config[item.Namespace] = make(map[string]interface{})
		}

		var value interface{}
		if err := json.Unmarshal([]byte(*item.PublishedValue), &value); err != nil {
			return nil, fmt.Errorf("parse config %s.%s failed: %w", item.Namespace, item.ConfigKey, err)
		}

		config[item.Namespace].(map[string]interface{})[item.ConfigKey] = value
	}

	return config, nil
}
