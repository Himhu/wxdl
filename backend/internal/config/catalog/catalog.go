package catalog

import (
	"slices"
	"strings"
)

// ConfigOwner 配置所有者
type ConfigOwner string

const (
	OwnerMiniProgramConfig ConfigOwner = "mini_program_config"
	OwnerSystemSetting     ConfigOwner = "system_setting"
)

// ConfigExposure 配置曝光度
type ConfigExposure string

const (
	ExposurePublicBootstrap ConfigExposure = "public_bootstrap"
	ExposureAdminOnly       ConfigExposure = "admin_only"
	ExposureBackendRuntime  ConfigExposure = "backend_runtime"
)

// ValueType 值类型
type ValueType string

const (
	ValueTypeString  ValueType = "string"
	ValueTypeBoolean ValueType = "boolean"
	ValueTypeNumber  ValueType = "number"
	ValueTypeJSON    ValueType = "json"
	ValueTypeSecret  ValueType = "secret"
)

// Entry 配置目录条目
type Entry struct {
	Key         string
	DisplayName string
	Owner       ConfigOwner
	Exposure    ConfigExposure
	ValueType   ValueType
	IsSecret    bool
	NeedsReload bool
	Description string
}

var entries = map[string]Entry{
	"general.app_name": {
		Key:         "general.app_name",
		DisplayName: "应用名称",
		Owner:       OwnerMiniProgramConfig,
		Exposure:    ExposurePublicBootstrap,
		ValueType:   ValueTypeString,
		Description: "Mini-program application name",
	},
	"general.support_wechat": {
		Key:         "general.support_wechat",
		DisplayName: "客服微信",
		Owner:       OwnerMiniProgramConfig,
		Exposure:    ExposurePublicBootstrap,
		ValueType:   ValueTypeString,
		Description: "Customer support WeChat shown to clients",
	},
	"feature.recharge_enabled": {
		Key:         "feature.recharge_enabled",
		DisplayName: "充值功能开关",
		Owner:       OwnerMiniProgramConfig,
		Exposure:    ExposurePublicBootstrap,
		ValueType:   ValueTypeBoolean,
		Description: "Whether recharge is enabled in the mini-program",
	},
	"recharge.min_amount": {
		Key:         "recharge.min_amount",
		DisplayName: "最小充值金额",
		Owner:       OwnerMiniProgramConfig,
		Exposure:    ExposurePublicBootstrap,
		ValueType:   ValueTypeNumber,
		Description: "Minimum recharge amount",
	},
	"recharge.max_amount": {
		Key:         "recharge.max_amount",
		DisplayName: "最大充值金额",
		Owner:       OwnerMiniProgramConfig,
		Exposure:    ExposurePublicBootstrap,
		ValueType:   ValueTypeNumber,
		Description: "Maximum recharge amount",
	},
	"wechat.app_id": {
		Key:         "wechat.app_id",
		DisplayName: "微信 AppID",
		Owner:       OwnerSystemSetting,
		Exposure:    ExposureBackendRuntime,
		ValueType:   ValueTypeString,
		NeedsReload: true,
		Description: "WeChat Mini-Program AppID used by backend runtime",
	},
	"wechat.app_secret": {
		Key:         "wechat.app_secret",
		DisplayName: "微信 Secret",
		Owner:       OwnerSystemSetting,
		Exposure:    ExposureBackendRuntime,
		ValueType:   ValueTypeSecret,
		IsSecret:    true,
		NeedsReload: true,
		Description: "WeChat Mini-Program AppSecret used by backend runtime",
	},
}

// Get 获取配置条目
func Get(key string) (Entry, bool) {
	entry, ok := entries[key]
	return entry, ok
}

// MustGet 获取配置条目（不存在则 panic）
func MustGet(key string) Entry {
	entry, ok := entries[key]
	if !ok {
		panic("unknown config key: " + key)
	}
	return entry
}

// List 列出所有配置条目（按 key 排序）
func List() []Entry {
	out := make([]Entry, 0, len(entries))
	for _, entry := range entries {
		out = append(out, entry)
	}
	slices.SortFunc(out, func(a, b Entry) int {
		return strings.Compare(a.Key, b.Key)
	})
	return out
}
