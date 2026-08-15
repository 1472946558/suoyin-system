/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: appointmentStore.js
 * 功能描述: 预约管理接口模块
 * 作者: 廖心慈
 * 创建日期: 2026-08-15
 */

var appConfig = require("./config").appConfig;
var request = require("./apiClient").request;
var formatRequestError = require("./apiClient").formatRequestError;

// 预约状态常量
var STATUS = {
  PENDING: "PENDING",
  CONFIRMED: "CONFIRMED",
  ARRIVED: "ARRIVED",
  COMPLETED: "COMPLETED",
  CANCELLED: "CANCELLED",
  NO_SHOW: "NO_SHOW",
  TERMINATED: "TERMINATED"
};

// 状态中文映射
var STATUS_TEXT = {
  PENDING: "待确认",
  CONFIRMED: "已确认",
  ARRIVED: "已到店",
  COMPLETED: "已完成",
  CANCELLED: "已取消",
  NO_SHOW: "未到店",
  TERMINATED: "已终止"
};

// 服务类型中文映射
var SERVICE_TYPE_TEXT = {
  OLD_FOR_NEW: "旧换新",
  REPAIR: "维修",
  CONSULT: "咨询",
  RECYCLE: "回收"
};

// 状态筛选 Tab
var STATUS_TABS = [
  { key: "", label: "全部" },
  { key: "PENDING", label: "待确认" },
  { key: "CONFIRMED", label: "已确认" },
  { key: "ARRIVED", label: "已到店" },
  { key: "COMPLETED", label: "已完成" },
  { key: "CANCELLED", label: "已取消" },
  { key: "NO_SHOW", label: "未到店" },
  { key: "TERMINATED", label: "已终止" }
];

// 状态对应可执行操作
var STATUS_ACTIONS = {
  PENDING: [
    { status: "CONFIRMED", label: "确认预约", type: "primary" },
    { status: "TERMINATED", label: "拒绝", type: "danger" }
  ],
  CONFIRMED: [
    { status: "ARRIVED", label: "标记到店", type: "primary" },
    { status: "TERMINATED", label: "终止", type: "danger" }
  ],
  ARRIVED: [
    { status: "COMPLETED", label: "完成服务", type: "primary" },
    { status: "NO_SHOW", label: "标记未到", type: "warn" }
  ],
  COMPLETED: [],
  CANCELLED: [],
  NO_SHOW: [],
  TERMINATED: []
};

// 状态颜色 class
var STATUS_CLASS = {
  PENDING: "status-pending",
  CONFIRMED: "status-confirmed",
  ARRIVED: "status-arrived",
  COMPLETED: "status-completed",
  CANCELLED: "status-cancelled",
  NO_SHOW: "status-noshow",
  TERMINATED: "status-terminated"
};

/**
 * 获取员工预约列表
 * @param {string} status - 状态筛选（空字符串=全部）
 * @returns {Promise<{items: Array, total: number}>}
 */
function listStaffAppointments(status) {
  var endpoint = appConfig.endpoints.staffAppointments;
  if (status) {
    endpoint = endpoint + "?status=" + encodeURIComponent(status);
  }
  return request({
    endpoint: endpoint,
    method: "GET"
  }).then(function(payload) {
    var data = payload && payload.data ? payload.data : payload;
    return {
      items: data.items || [],
      total: data.total || 0
    };
  });
}

/**
 * 获取员工预约详情
 * @param {string} appointmentId - 预约 ID
 * @returns {Promise<object>}
 */
function getStaffAppointment(appointmentId) {
  return request({
    endpoint: appConfig.endpoints.staffAppointmentDetail,
    method: "GET",
    params: { id: appointmentId }
  }).then(function(payload) {
    return payload && payload.data ? payload.data : payload;
  });
}

/**
 * 更新预约状态
 * @param {string} appointmentId - 预约 ID
 * @param {string} newStatus - 新状态（CONFIRMED/ARRIVED/COMPLETED/NO_SHOW/TERMINATED）
 * @returns {Promise<object>} 更新后的预约 DTO
 */
function updateAppointmentStatus(appointmentId, newStatus) {
  return request({
    endpoint: appConfig.endpoints.staffAppointmentStatus,
    method: "PUT",
    params: { id: appointmentId },
    data: { status: newStatus }
  }).then(function(payload) {
    return payload && payload.data ? payload.data : payload;
  });
}

/**
 * 格式化预约卡片展示信息
 * @param {object} appt - StaffAppointmentDTO
 * @returns {object} 格式化后的展示对象
 */
function formatAppointmentCard(appt) {
  if (!appt) return null;
  return {
    id: appt.id,
    appointmentNo: appt.appointmentNo || "",
    customerName: appt.customerName || "未提供",
    customerPhone: appt.customerPhone || "",
    storeName: appt.storeName || "",
    serviceTypeText: appt.serviceTypeText || SERVICE_TYPE_TEXT[appt.serviceType] || appt.serviceType || "",
    appointmentDate: appt.appointmentDate || "",
    appointmentTime: appt.appointmentTime || "",
    status: appt.status,
    statusText: appt.statusText || STATUS_TEXT[appt.status] || appt.status,
    statusClass: STATUS_CLASS[appt.status] || "",
    remark: appt.remark || "",
    createdAt: appt.createdAt || "",
    confirmedBy: appt.confirmedBy || ""
  };
}

module.exports = {
  STATUS: STATUS,
  STATUS_TEXT: STATUS_TEXT,
  SERVICE_TYPE_TEXT: SERVICE_TYPE_TEXT,
  STATUS_TABS: STATUS_TABS,
  STATUS_ACTIONS: STATUS_ACTIONS,
  STATUS_CLASS: STATUS_CLASS,
  listStaffAppointments: listStaffAppointments,
  getStaffAppointment: getStaffAppointment,
  updateAppointmentStatus: updateAppointmentStatus,
  formatAppointmentCard: formatAppointmentCard,
  formatRequestError: formatRequestError
};
