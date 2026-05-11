const { listMembersOnline, getMemberStats } = require("../../utils/catalogStore");
const { mergeDraft } = require("../../utils/orderStore");
const { canAccessFeature, getProfile } = require("../../utils/userStore");

function formatCurrency(amount) {
  return `¥${Number(amount || 0).toFixed(2)}`;
}

function filterMembers(members, filters) {
  return (members || []).filter(function(item) {
    const keyword = (filters.keyword || "").toLowerCase();
    const matchesKeyword = !keyword || [item.name, item.phone, item.level, item.preferredPurity, item.tags.join(" / ")]
      .join(" ")
      .toLowerCase()
      .indexOf(keyword) > -1;
    const matchesLevel = filters.level === "all" || item.level === filters.level;
    const matchesStatus = filters.status === "all" || item.status === filters.status;
    return matchesKeyword && matchesLevel && matchesStatus;
  });
}

Page({
  data: {
    profile: getProfile(),
    members: [],
    filteredMembers: [],
    selectedId: "",
    selectedMember: null,
    stats: {
      total: 0,
      vipCount: 0,
      activeCount: 0,
      followUpCount: 0
    },
    filters: {
      keyword: "",
      level: "all",
      status: "all"
    },
    levelOptions: ["all", "VIP", "潜力客户", "普通会员"],
    statusOptions: ["all", "active", "follow_up", "sleeping"],
    loading: false
  },

  onShow() {
    if (!canAccessFeature("members")) {
      wx.showToast({ title: "当前账号没有会员查看权限", icon: "none" });
      wx.switchTab({ url: "/pages/home/index" });
      return;
    }
    this.loadMembers();
  },

  loadMembers() {
    this.setData({ loading: true, profile: getProfile() });
    listMembersOnline().then((members) => {
      const stats = getMemberStats(members);
      this.setData({
        members: members,
        stats: stats
      });
      this.applyFilters();
    }).finally(() => {
      this.setData({ loading: false });
    });
  },

  applyFilters() {
    const filteredMembers = filterMembers(this.data.members, this.data.filters);
    const selectedId = filteredMembers.some(function(item) {
      return item.id === this.data.selectedId;
    }, this)
      ? this.data.selectedId
      : (filteredMembers[0] && filteredMembers[0].id) || "";
    const selectedMember = filteredMembers.find(function(item) {
      return item.id === selectedId;
    }) || null;

    this.setData({
      filteredMembers: filteredMembers,
      selectedId: selectedId,
      selectedMember: selectedMember
    });
  },

  onKeywordInput(event) {
    const filters = Object.assign({}, this.data.filters, {
      keyword: event.detail.value
    });
    this.setData({ filters: filters });
    this.applyFilters();
  },

  onLevelChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      level: this.data.levelOptions[index]
    });
    this.setData({ filters: filters });
    this.applyFilters();
  },

  onStatusChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      status: this.data.statusOptions[index]
    });
    this.setData({ filters: filters });
    this.applyFilters();
  },

  selectMember(event) {
    const id = event.currentTarget.dataset.id;
    const selectedMember = this.data.filteredMembers.find(function(item) {
      return item.id === id;
    }) || null;
    this.setData({
      selectedId: id,
      selectedMember: selectedMember
    });
  },

  useForCashier() {
    const member = this.data.selectedMember;
    if (!member) return;
    mergeDraft({
      customerName: member.name,
      customerPhone: member.phone,
      sourceChannel: member.totalOrders > 0 ? "熟客复购" : "到店散客",
      remark: member.notes
    });
    wx.showToast({ title: "会员信息已带入收银草稿", icon: "none" });
    wx.switchTab({ url: "/pages/order/index" });
  },

  copyPhone() {
    const member = this.data.selectedMember;
    if (!member) return;
    wx.setClipboardData({
      data: member.phone
    });
  },

  formatCurrency(amount) {
    return formatCurrency(amount);
  }
});
