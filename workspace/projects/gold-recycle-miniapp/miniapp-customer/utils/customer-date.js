// utils/customer-date.js - 日期工具

const WEEKDAYS = ['周日', '周一', '周二', '周三', '周四', '周五', '周六'];

/** 格式化日期 YYYY-MM-DD */
function fmtDate(d) {
  d = d || new Date();
  if (typeof d === 'string') return d;
  return [
    d.getFullYear(),
    pad(d.getMonth() + 1),
    pad(d.getDate())
  ].join('-');
}

/** 格式化为 MM/DD */
function fmtShort(d) {
  d = d || new Date();
  return [pad(d.getMonth() + 1), pad(d.getDate())].join('/');
}

/** 加天数 */
function addDays(date, n) {
  const d = new Date(date.getTime());
  d.setDate(d.getDate() + n);
  return d;
}

/** 中文星期 */
function weekdayCN(d) {
  return WEEKDAYS[d.getDay()];
}

/** 相对今日描述 */
function relativeToday(date) {
  const today = new Date();
  today.setHours(0, 0, 0, 0);
  const target = new Date(date);
  target.setHours(0, 0, 0, 0);
  const diff = Math.round((target - today) / 86400000);
  if (diff === 0) return '今天';
  if (diff === 1) return '明天';
  if (diff === 2) return '后天';
  return weekdayCN(date);
}

/** 解析 HH:MM 为分钟数 */
function parseTimeMin(t) {
  const m = /^(\d{1,2}):(\d{2})$/.exec(t || '');
  if (!m) return 0;
  return parseInt(m[1], 10) * 60 + parseInt(m[2], 10);
}

/** 分钟数转为 HH:MM */
function minToTime(min) {
  const h = Math.floor(min / 60);
  const m = min % 60;
  return pad(h) + ':' + pad(m);
}

/** 计算 endTime = startTime + 30 分钟 */
function timePlusMin(time, minutes) {
  return minToTime(parseTimeMin(time) + minutes);
}

/** 时间是否已过（对比现在） */
function isPastTime(date, time) {
  const now = new Date();
  const t = new Date(fmtDate(date) + 'T' + time + ':00');
  return t.getTime() <= now.getTime();
}

/** 7 天日期范围（用于预约可选日期） */
function next7Days() {
  const days = [];
  for (let i = 0; i < 7; i++) {
    const d = addDays(new Date(), i);
    days.push({
      date: fmtDate(d),
      label: i === 0 ? '今天' : (i === 1 ? '明天' : weekdayCN(d)),
      shortLabel: fmtShort(d),
      fullDate: d
    });
  }
  return days;
}

function pad(n) { return n < 10 ? '0' + n : '' + n; }

module.exports = {
  fmtDate,
  fmtShort,
  addDays,
  weekdayCN,
  relativeToday,
  parseTimeMin,
  minToTime,
  timePlusMin,
  isPastTime,
  next7Days
};
