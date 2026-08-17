// pkg-customer/recycle-info/index.js - 黄金回收服务介绍页
const { api } = require('../../utils/request.js');
const { absUrl } = require('../../utils/customer-services.js');

// 后台配置的服务项图标 code → emoji 映射
const ICON_MAP = {
  recycle: '♻',
  repair: '🔧',
  consult: '💎',
  custom: '✨',
};

Page({
  data: {
    recycleInfo: null,
    loading: true,
    error: null
  },

  onLoad() {
    this.loadRecycleInfo();
  },

  normalizeInfo(info) {
    if (!info) return null;
    const services = (info.services || []).map(s => ({
      icon: ICON_MAP[s.icon] || s.icon || '💎',
      title: s.title,
      desc: s.desc
    }));
    return {
      title: info.title || '黄金回收服务',
      intro: info.intro || '',
      imageUrl: info.imageUrl ? absUrl(info.imageUrl) : '',
      process: info.process || [],
      services,
      notices: info.notices || []
    };
  },

  loadRecycleInfo() {
    this.setData({ loading: true, error: null });
    api.getRecycleInfo()
      .then(data => {
        this.setData({ recycleInfo: this.normalizeInfo(data) || this.defaultInfo(), loading: false });
      })
      .catch(err => {
        // 降级：后端无数据时用本地默认文案
        this.setData({ recycleInfo: this.defaultInfo(), loading: false, error: null });
      });
  },

  defaultInfo() {
    return {
      title: '黄金回收服务',
      intro: '金匠倌提供专业的黄金回收服务，旧金换新款、黄金维修、到店回收咨询。专业检测，流程透明，包损耗。',
      process: [
        { step: 1, title: '到店预约', desc: '选择附近门店，预约到店时间' },
        { step: 2, title: '实物检测', desc: '专业检测黄金纯度与克重' },
        { step: 3, title: '服务说明', desc: '专业顾问说明检测结果与可办理服务' },
        { step: 4, title: '到店办理', desc: '确认服务内容后由门店协助办理' }
      ],
      services: [
        { icon: '♻', title: '旧金换新', desc: '旧金折价换新款，工费透明' },
        { icon: '🔧', title: '黄金维修', desc: '专业维修，恢复光泽' },
        { icon: '💎', title: '到店回收', desc: '现场检测，具体办理内容以门店说明为准' },
        { icon: '💬', title: '款式咨询', desc: '工费咨询，款式定制' }
      ],
      notices: [
        '最终服务内容以门店现场检测与沟通结果为准',
        '请携带购买凭证或相关证明',
        '回收前请确认物品权属清晰',
        '本版本仅支持预约到店咨询'
      ]
    };
  },

  onAppointment() {
    wx.navigateTo({ url: '/pkg-customer/appointment-create/index' });
  },

  onBack() {
    const pages = getCurrentPages();
    if (pages.length > 1) {
      wx.navigateBack();
    } else {
      wx.switchTab({ url: '/pages/home/index' });
    }
  }
});
