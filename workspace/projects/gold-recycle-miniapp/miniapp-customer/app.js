// app.js - 金匠倌顾客端小程序入口
const { ensureCustomerSession } = require('./utils/customer-auth.js');

App({
  globalData: {
    apiBase: 'https://jinjiangguan.com',
    brandName: '金匠倌',
    servicePhone: '',
    homeConfig: null,  // 缓存首页配置（含 appointmentNotes, appointmentRules 等）
    userLocation: null
  },

  onLaunch() {
    // 启动时尝试静默登录（游客模式）
    ensureCustomerSession().catch(() => {
      // 静默登录失败不阻塞 UI
    });
  },

  onShow() {
    // 小程序从后台进入前台
  },

  onError(err) {
    console.error('[App] onError', err);
  }
});
