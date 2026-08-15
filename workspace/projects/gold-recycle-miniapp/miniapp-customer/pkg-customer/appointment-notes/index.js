// pkg-customer/appointment-notes/index.js - 修改预约备注
const { api } = require('../../utils/request.js');

Page({
  data: {
    appointmentId: '',
    remark: '',
    originalRemark: '',
    saving: false
  },

  onLoad(query) {
    this.setData({
      appointmentId: query.id || '',
      remark: query.remark ? decodeURIComponent(query.remark) : '',
      originalRemark: query.remark ? decodeURIComponent(query.remark) : ''
    });
  },

  onInput(e) {
    this.setData({ remark: e.detail.value });
  },

  onSave() {
    const remark = (this.data.remark || '').trim();
    if (remark.length > 200) {
      wx.showToast({ title: '备注不超过200字', icon: 'none' });
      return;
    }
    this.setData({ saving: true });
    api.updateAppointmentNotes(this.data.appointmentId, remark)
      .then(() => {
        wx.showToast({ title: '保存成功', icon: 'success' });
        setTimeout(() => wx.navigateBack(), 800);
      })
      .catch(err => {
        wx.showToast({ title: err.message || '保存失败', icon: 'none' });
      })
      .then(() => {
        this.setData({ saving: false });
      });
  },

  onClear() {
    this.setData({ remark: '' });
  }
});
