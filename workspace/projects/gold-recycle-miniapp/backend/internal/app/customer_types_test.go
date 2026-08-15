/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: customer_types_test.go
 * 功能描述: CustomerRecycleInfo 新旧 JSON 格式兼容性测试
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

package app

import (
	"encoding/json"
	"testing"
)

// 旧版扁平字符串格式（存量 app_configs 持久化数据）必须能成功反序列化，
// 否则服务启动时 loadConfig 报错导致崩溃（生产事故 2026-08-15）。
func TestCustomerRecycleInfoLegacyStringFormat(t *testing.T) {
	legacy := `{"title":"黄金回收服务","content":"旧版介绍文案","process":"旧版流程描述","notes":"注意事项一\n注意事项二\n\n注意事项三"}`
	var info CustomerRecycleInfo
	if err := json.Unmarshal([]byte(legacy), &info); err != nil {
		t.Fatalf("legacy format unmarshal failed: %v", err)
	}
	if info.Title != "黄金回收服务" {
		t.Errorf("Title = %q, want %q", info.Title, "黄金回收服务")
	}
	if info.Intro != "旧版介绍文案" {
		t.Errorf("Intro = %q, want %q", info.Intro, "旧版介绍文案")
	}
	if len(info.Notices) != 3 {
		t.Errorf("Notices length = %d, want 3 (got %v)", len(info.Notices), info.Notices)
	}
	if len(info.Process) != 0 || len(info.Services) != 0 {
		t.Errorf("legacy process/services should be empty, got process=%v services=%v", info.Process, info.Services)
	}
}

// 新版结构化格式正常反序列化
func TestCustomerRecycleInfoModernFormat(t *testing.T) {
	modern := `{"title":"黄金回收","intro":"介绍","process":[{"step":1,"title":"预约","desc":"到店"}],"services":[{"icon":"♻","title":"旧金换新"}],"notices":["注意"]}`
	var info CustomerRecycleInfo
	if err := json.Unmarshal([]byte(modern), &info); err != nil {
		t.Fatalf("modern format unmarshal failed: %v", err)
	}
	if len(info.Process) != 1 || info.Process[0].Title != "预约" {
		t.Errorf("Process = %v, want 1 step titled 预约", info.Process)
	}
	if len(info.Services) != 1 || info.Services[0].Title != "旧金换新" {
		t.Errorf("Services = %v, want 1 item titled 旧金换新", info.Services)
	}
}

// 完全无法识别的格式仍应返回错误
func TestCustomerRecycleInfoInvalidFormat(t *testing.T) {
	invalid := `{"title":123,"process":[1,2]}`
	var info CustomerRecycleInfo
	if err := json.Unmarshal([]byte(invalid), &info); err == nil {
		t.Errorf("expected error for invalid format, got %+v", info)
	}
}

// 旧格式但不含 process 字段：不能因新格式解析"成功"而静默丢弃 content/notes
func TestCustomerRecycleInfoLegacyWithoutProcess(t *testing.T) {
	legacy := `{"title":"T","content":"C","notes":"a\nb"}`
	var info CustomerRecycleInfo
	if err := json.Unmarshal([]byte(legacy), &info); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if info.Intro != "C" {
		t.Errorf("Intro = %q, want %q（旧格式静默走新格式分支会丢失数据）", info.Intro, "C")
	}
	if len(info.Notices) != 2 {
		t.Errorf("Notices = %v, want 2 items", info.Notices)
	}
}

// 旧数据经 Marshal → Unmarshal 往返后应稳定为新格式
func TestCustomerRecycleInfoRoundTrip(t *testing.T) {
	legacy := `{"title":"T","content":"C","notes":"a\nb"}`
	var info CustomerRecycleInfo
	if err := json.Unmarshal([]byte(legacy), &info); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	data, err := json.Marshal(info)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var again CustomerRecycleInfo
	if err := json.Unmarshal(data, &again); err != nil {
		t.Fatalf("round-trip unmarshal failed: %v", err)
	}
	if again.Title != "T" || again.Intro != "C" || len(again.Notices) != 2 {
		t.Errorf("round-trip mismatch: %+v", again)
	}
}
