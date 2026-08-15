// utils/customer-services.js - 服务类型与图标映射

// 预约服务类型 → 文字
const SERVICE_TYPES = [
  { code: 'OLD_FOR_NEW', text: '旧金换新', icon: '♻' },
  { code: 'REPAIR',      text: '黄金维修', icon: '🔧' },
  { code: 'CONSULT',     text: '款式工费咨询', icon: '💬' },
  { code: 'RECYCLE',     text: '到店回收咨询', icon: '💎' }
];

function serviceTypeText(code) {
  const m = SERVICE_TYPES.find(s => s.code === code);
  return m ? m.text : code;
}

// 预约状态 → 文字 + 颜色
const APPOINTMENT_STATUS = {
  PENDING:    { text: '待确认', color: '#FF9800' },
  CONFIRMED:  { text: '已确认', color: '#4CAF50' },
  ARRIVED:    { text: '已到店', color: '#1976D2' },
  COMPLETED:  { text: '已完成', color: '#8A8F99' },
  CANCELLED:  { text: '已取消', color: '#B0B5BD' },
  NO_SHOW:    { text: '未到店', color: '#F44336' },
  TERMINATED: { text: '已终止', color: '#F44336' }
};

function appointmentStatusText(status) {
  const m = APPOINTMENT_STATUS[status];
  return m ? m.text : status;
}

function appointmentStatusColor(status) {
  const m = APPOINTMENT_STATUS[status];
  return m ? m.color : '#8A8F99';
}

// 商品分类
const PRODUCT_CATEGORIES = ['戒指', '项链', '手镯', '吊坠', '耳饰', '其他'];

/** 金额千分位格式化（保留最多 2 位小数） */
function fmtPrice(n) {
  if (n == null || isNaN(n)) return '0';
  const s = (Math.round(n * 100) / 100).toString();
  const parts = s.split('.');
  parts[0] = parts[0].replace(/\B(?=(\d{3})+(?!\d))/g, ',');
  return parts.join('.');
}

/** 相对路径图片补全为绝对地址（后端 imageUrl 为 /assets/... 相对路径） */
function absUrl(url) {
  if (!url) return '';
  if (/^https?:\/\//.test(url)) return url;
  const app = typeof getApp === 'function' ? getApp() : null;
  const base = (app && app.globalData && app.globalData.apiBase) || 'https://jinjiangguan.com';
  return base + url;
}

module.exports = {
  SERVICE_TYPES,
  serviceTypeText,
  APPOINTMENT_STATUS,
  appointmentStatusText,
  appointmentStatusColor,
  PRODUCT_CATEGORIES,
  fmtPrice,
  absUrl
};
