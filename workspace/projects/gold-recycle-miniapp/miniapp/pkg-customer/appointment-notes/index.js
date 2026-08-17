// pkg-customer/appointment-notes/index.js - 预约备注修改页
const { api } = require('../../utils/request.js');
const { getCustomerToken } = require('../../utils/customer-auth.js');

Page({
  data: {
    id: '',
    remark: '',
    saving: false
  },

  onLoad(query) {
    var id = (query && query.id) || '';
    var remark = '';
    try {
      remark = decodeURIComponent((query && query.remark) || '');
    } catch (e) {
      remark = '';
    }
    this.setData({
      id: id,
      remark: remark
    });
    if (!id || !getCustomerToken()) {
      wx.showToast({ title: '请先登录后编辑预约', icon: 'none' });
      setTimeout(() => wx.navigateBack(), 300);
    }
  },

  onInput(e) {
    this.setData({ remark: e.detail.value });
  },

  onClear() {
    this.setData({ remark: '' });
  },

  onSave() {
    if (this.data.saving) return;
    if (!this.data.id || !getCustomerToken()) {
      wx.showToast({ title: '请先登录后编辑预约', icon: 'none' });
      return;
    }
    const remark = (this.data.remark || '').trim();
    if (remark.length > 200) {
      wx.showToast({ title: '备注不能超过 200 字', icon: 'none' });
      return;
    }
    this.setData({ saving: true });
    wx.showLoading({ title: '保存中...' });
    api.updateAppointmentNotes(this.data.id, remark)
      .then(() => {
        wx.hideLoading();
        this.setData({ saving: false, remark: remark });
        wx.showToast({ title: '已保存', icon: 'success' });
        setTimeout(() => wx.navigateBack(), 800);
      })
      .catch(err => {
        wx.hideLoading();
        this.setData({ saving: false });
        wx.showToast({ title: err.message || '保存失败', icon: 'none' });
      });
  }
});
