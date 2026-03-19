package repository

import (
	"context"

	"gorm.io/gorm"
)

type TxRepositories interface {
	Agent() AgentRepository
	Card() CardRepository
	Points() PointsRepository
	RechargeRequest() RechargeRequestRepository
	PointsRecord() PointsRecordRepository
	SystemSetting() SystemSettingRepository
}

type TxManager interface {
	WithinTx(ctx context.Context, fn func(repos TxRepositories) error) error
}

type txManager struct {
	db *gorm.DB
}

type txRepositories struct {
	agent           AgentRepository
	card            CardRepository
	points          PointsRepository
	rechargeRequest RechargeRequestRepository
	pointsRecord    PointsRecordRepository
	systemSetting   SystemSettingRepository
}

func NewTxManager(db *gorm.DB) TxManager {
	return &txManager{db: db}
}

func (m *txManager) WithinTx(ctx context.Context, fn func(repos TxRepositories) error) error {
	return m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		repos := &txRepositories{
			agent:           NewAgentRepository(tx),
			card:            NewCardRepository(tx),
			points:          NewPointsRepository(tx),
			rechargeRequest: NewRechargeRequestRepository(tx),
			pointsRecord:    NewPointsRecordRepository(tx),
			systemSetting:   NewSystemSettingRepository(tx),
		}
		return fn(repos)
	})
}

func (r *txRepositories) Agent() AgentRepository {
	return r.agent
}

func (r *txRepositories) Card() CardRepository {
	return r.card
}

func (r *txRepositories) Points() PointsRepository {
	return r.points
}

func (r *txRepositories) RechargeRequest() RechargeRequestRepository {
	return r.rechargeRequest
}

func (r *txRepositories) PointsRecord() PointsRecordRepository {
	return r.pointsRecord
}

func (r *txRepositories) SystemSetting() SystemSettingRepository {
	return r.systemSetting
}
