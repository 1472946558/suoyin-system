// pkg-customer/appointment-detail/index.js - 预约详情
const { api } = require('../../utils/request.js');
const { appointmentStatusText, appointmentStatusColor, serviceTypeText } = require('../../utils/customer-services.js');

// 可取消状态
var CANCELLABLE = ['PENDING', 'CONFIRMED'];
// 可改备注状态
var NOTES_EDITABLE = ['PENDING', 'CONFIRMED'];

Page({
  data: {
    id: '',
    detail: null,
    loading: true,
    error: null,
    statusText: '',
    statusColor: '',
    serviceText: '',
    canCancel: false,
    canEditNotes: false
  },

  onLoad(query) {
    this.setData({ id: query.id || '' });
    this.loadDetail();
  },

  onShow() {
    if (this.data.id && !this.data.loading) {
      this.loadDetail();
    }
  },

  loadDetail() {
    this.setData({ loading: true, error: null });
    api.getAppointment(this.data.id)
      .then(detail => {
        this.setData({
          detail,
          loading: false,
          statusText: appointmentStatusText(detail.status),
          statusColor: appointmentStatusColor(detail.status),
          serviceText: serviceTypeText(detail.serviceType),
          canCancel: this.checkCanCancel(detail),
          canEditNotes: NOTES_EDITABLE.indexOf(detail.status) !== -1
        });
      })
      .catch(err => {
        this.setData({
          loading: false,
          error: err.message || '加载失败'
        });
      });
  },

  // 检查是否可取消（PENDING/CONFIRMED 且距预约 > 2 小时）
  checkCanCancel(detail) {
    if (CANCELLABLE.indexOf(detail.status) === -1) return false;
    const apptTime = new Date(detail.appointmentDate + 'T' + detail.startTime + ':00');
    const now = new Date();
    const diff = apptTime.getTime() - now.getTime();
    return diff > 2 * 60 * 60 * 1000; // > 2 小时
  },

  // 取消预约
  onCancel() {
    const detail = this.data.detail;
    if (!detail) return;

    wx.showModal({
      title: '取消预约',
      content: '确定要取消此预约吗？取消后不可恢复。',
      confirmColor: '#F44336',
      success: (res) => {
        if (!res.confirm) return;
        api.cancelAppointment(this.data.id, '用户取消')
          .then(() => {
            wx.showToast({ title: '已取消', icon: 'success' });
            this.loadDetail();
          })
          .catch(err => {
            wx.showToast({ title: err.message || '取消失败', icon: 'none' });
          });
      }
    });
  },

  // 修改备注
  onEditNotes() {
    const remark = this.data.detail.remark || '';
    wx.navigateTo({
      url: '/pkg-customer/appointment-notes/index?id=' + this.data.id + '&remark=' + encodeURIComponent(remark)
    });
  },

  // 拨打门店电话
  onCallStore() {
    const detail = this.data.detail;
    if (!detail || !detail.storePhone) {
      wx.showToast({ title: '暂无门店电话', icon: 'none' });
      return;
    }
    wx.makePhoneCall({
      phoneNumber: String(detail.storePhone).replace(/\s/g, '')
    }).catch(() => {});
  },

  // 导航到门店
  onNavigate() {
    const detail = this.data.detail;
    if (!detail) return;
    if (!detail.storeLongitude || !detail.storeLatitude) {
      wx.showToast({ title: '该门店暂未配置位置', icon: 'none' });
      return;
    }
    wx.openLocation({
      latitude: detail.storeLatitude,
      longitude: detail.storeLongitude,
      name: detail.storeName || '',
      address: detail.storeAddress || '',
      scale: 16
    });
  }
});
