// pkg-customer/appointment-detail/index.js - 预约详情
const { api } = require('../../utils/request.js');
const { appointmentStatusText, appointmentStatusColor, appointmentStatusBgColor, serviceTypeText } = require('../../utils/customer-services.js');
const { getCustomerToken, getStoredProfile } = require('../../utils/customer-auth.js');

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
    statusBgColor: '',
    serviceText: '',
    contactPhone: '',
    canCancel: false,
    canEditNotes: false,
    cancelling: false
  },

  onLoad(query) {
    var id = (query && query.id) || '';
    // 从本地存储读取用户手机号作为联系电话
    var profile = getStoredProfile();
    this.setData({
      id: id,
      contactPhone: (profile && profile.phone) ? profile.phone : ''
    });
    if (!id || !getCustomerToken()) {
      this.setData({ loading: false, error: id ? '请先登录后查看预约' : '缺少预约信息' });
      return;
    }
    this.loadDetail();
  },

  onShow() {
    if (this.data.id && !this.data.loading) {
      this.loadDetail();
    }
  },

  loadDetail() {
    if (!this.data.id || !getCustomerToken()) return;
    this.setData({ loading: true, error: null });
    api.getAppointment(this.data.id)
      .then(detail => {
        this.setData({
          detail,
          loading: false,
          statusText: detail.statusText || appointmentStatusText(detail.status),
          statusColor: appointmentStatusColor(detail.status),
          statusBgColor: appointmentStatusBgColor(detail.status),
          serviceText: detail.serviceTypeText || serviceTypeText(detail.serviceType),
          canCancel: typeof detail.canCancel === 'boolean' ? detail.canCancel : this.checkCanCancel(detail),
          canEditNotes: typeof detail.canEditNotes === 'boolean' ? detail.canEditNotes : NOTES_EDITABLE.indexOf(detail.status) !== -1
        });
      })
      .catch(err => {
        this.setData({
          loading: false,
          error: err.message || '加载失败'
        });
      });
  },

  // 检查是否可取消（PENDING/CONFIRMED 且距预约 > cancelLeadMinutes）
  checkCanCancel(detail) {
    if (CANCELLABLE.indexOf(detail.status) === -1) return false;
    var apptTime = new Date(detail.appointmentDate + 'T' + detail.appointmentTime + ':00');
    if (isNaN(apptTime.getTime())) return true;
    var now = new Date();
    // 从后端配置读取取消提前量（分钟），默认 120
    var app = getApp();
    var rules = (app.globalData.homeConfig && app.globalData.homeConfig.appointmentRules) || {};
    var cancelLeadMinutes = rules.cancelLeadMinutes || 120;
    var diff = apptTime.getTime() - now.getTime();
    return diff > cancelLeadMinutes * 60 * 1000;
  },

  // 取消预约
  onCancel() {
    const detail = this.data.detail;
    if (!detail || !this.data.canCancel || this.data.cancelling) return;
    if (!getCustomerToken()) {
      wx.showToast({ title: '请先登录后操作', icon: 'none' });
      return;
    }

    wx.showModal({
      title: '取消预约',
      content: '确定要取消此预约吗？取消后不可恢复。',
      confirmColor: '#F44336',
      success: (res) => {
        if (!res.confirm) return;
        this.setData({ cancelling: true });
        api.cancelAppointment(this.data.id, '用户取消')
          .then(() => {
            wx.showToast({ title: '已取消', icon: 'success' });
            this.loadDetail();
          })
          .catch(err => {
            wx.showToast({ title: err.message || '取消失败', icon: 'none' });
          })
          .finally(() => this.setData({ cancelling: false }));
      }
    });
  },

  // 修改备注
  onEditNotes() {
    if (!this.data.id || !this.data.canEditNotes) return;
    if (!getCustomerToken()) {
      wx.showToast({ title: '请先登录后操作', icon: 'none' });
      return;
    }
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

  // 拨打自己的联系电话
  onCallSelf() {
    var phone = this.data.contactPhone;
    if (!phone || phone.indexOf('*') !== -1) {
      wx.showToast({ title: '手机号已脱敏，请通过授权后使用', icon: 'none' });
      return;
    }
    wx.makePhoneCall({
      phoneNumber: String(phone).replace(/\s/g, '')
    }).catch(() => {});
  },

  // 导航到门店 — DTO 不含坐标，跳转到门店详情页查看地图
  onNavigate() {
    var detail = this.data.detail;
    if (!detail || !detail.storeId) {
      wx.showToast({ title: '暂无门店信息', icon: 'none' });
      return;
    }
    wx.navigateTo({
      url: '/pkg-customer/store-detail/index?id=' + detail.storeId
    });
  }
});
