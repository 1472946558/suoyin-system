// pkg-customer/appointment-notes/index.js - 预约备注修改页
const { api } = require('../../utils/request.js');

Page({
  data: {
    id: '',
    remark: ''
  },

  onLoad(query) {
    this.setData({
      id: query.id || '',
      remark: decodeURIComponent(query.remark || '')
    });
  },

  onInput(e) {
    this.setData({ remark: e.detail.value });
  },

  onClear() {
    this.setData({ remark: '' });
  },

  onSave() {
    if (!this.data.id) {
      wx.showToast({ title: '参数错误', icon: 'none' });
      return;
    }
    wx.showLoading({ title: '保存中...' });
    api.updateAppointmentNotes(this.data.id, this.data.remark)
      .then(() => {
        wx.hideLoading();
        wx.showToast({ title: '已保存', icon: 'success' });
        setTimeout(() => wx.navigateBack(), 800);
      })
      .catch(err => {
        wx.hideLoading();
        wx.showToast({ title: err.message || '保存失败', icon: 'none' });
      });
  }
});
