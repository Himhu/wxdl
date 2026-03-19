// Package model 定义系统常量
package model

const (
	AgentStatusActive   = 1
	AgentStatusDisabled = 2
)

const (
	CardStatusUnused    = 1
	CardStatusUsed      = 2
	CardStatusDestroyed = 3
)

const (
	RechargeRequestStatusPending  = 0
	RechargeRequestStatusApproved = 1
	RechargeRequestStatusRejected = 2
)

const (
	PointsRecordTypeRecharge = 1
	PointsRecordTypeConsume  = 2
	PointsRecordTypeRefund   = 3
)

const (
	AdminStatusActive   = 1
	AdminStatusDisabled = 2
)
