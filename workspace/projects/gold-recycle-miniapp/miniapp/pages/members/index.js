/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: index.js
 * 功能描述: 页面模块
 * 作者: 廖心慈
 * 创建日期: 2026-06-03
 */

const {
  listMembersOnline,
  saveMemberRecord,
  getMemberStats
} = require("../../utils/catalogStore");
const { mergeDraft } = require("../../utils/orderStore");
const { canAccessFeature, getProfile } = require("../../utils/userStore");
const { formatRequestError } = require("../../utils/apiClient");

const LEVEL_FILTER_VALUES = ["all", "VIP", "潜力客户", "普通会员"];
const LEVEL_FILTER_LABELS = ["全部等级", "VIP", "潜力客户", "普通会员"];
const MEMBER_LEVEL_OPTIONS = ["VIP", "潜力客户", "普通会员"];
const STATUS_FILTER_VALUES = ["all", "active", "follow_up", "sleeping"];
const STATUS_FILTER_LABELS = ["全部状态", "活跃", "待回访", "沉睡"];
const STATUS_VALUES = ["active", "follow_up", "sleeping"];
const STATUS_LABELS = ["活跃", "待回访", "沉睡"];
const STATUS_TEXT = {
  active: "活跃",
  follow_up: "待回访",
  sleeping: "沉睡"
};

function formatCurrency(amount) {
  return `¥${Number(amount || 0).toFixed(2)}`;
}

function normalizeMemberView(member) {
  const nextMember = Object.assign({}, member || {});
  nextMember.tags = Array.isArray(nextMember.tags) ? nextMember.tags : [];
  nextMember.statusText = STATUS_TEXT[nextMember.status] || "未设置";
  nextMember.managerNameText = nextMember.managerName || "门店";
  nextMember.totalRecycleAmountText = formatCurrency(nextMember.totalRecycleAmount);
  nextMember.records = [
    {
      title: "回收记录",
      desc: `${nextMember.preferredPurity || "未设置"} · ${Number(nextMember.totalOrders || 0)} 笔`,
      amount: nextMember.totalRecycleAmountText,
      time: nextMember.lastVisitAt || "暂无"
    },
    {
      title: "跟进记录",
      desc: nextMember.notes || "暂无跟进备注",
      amount: nextMember.statusText,
      time: nextMember.sourceChannel || "门店登记"
    }
  ];
  return nextMember;
}

function filterMembers(members, filters) {
  return (members || []).filter(function(item) {
    const tags = Array.isArray(item.tags) ? item.tags : [];
    const keyword = (filters.keyword || "").toLowerCase();
    const matchesKeyword = !keyword || [item.name, item.phone, item.level, item.preferredPurity, tags.join(" / ")]
      .join(" ")
      .toLowerCase()
      .indexOf(keyword) > -1;
    const matchesLevel = filters.level === "all" || item.level === filters.level;
    const matchesStatus = filters.status === "all" || item.status === filters.status;
    return matchesKeyword && matchesLevel && matchesStatus;
  });
}

function buildMemberEditor(member) {
  const source = member || {};
  return {
    id: source.id || "",
    name: source.name || "",
    phone: source.phone || "",
    level: source.level || "普通会员",
    status: source.status || "active",
    preferredPurity: source.preferredPurity || "足金999",
    sourceChannel: source.sourceChannel || "门店登记",
    managerName: source.managerName || "",
    totalOrders: String(source.totalOrders || 0),
    totalRecycleAmount: String(source.totalRecycleAmount || 0),
    tagsText: Array.isArray(source.tags) ? source.tags.join("，") : "",
    notes: source.notes || ""
  };
}

function isValidPhone(phone) {
  return /^1[3-9]\d{9}$/.test(String(phone || ""));
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
    levelFilterValues: LEVEL_FILTER_VALUES,
    levelFilterLabels: LEVEL_FILTER_LABELS,
    memberLevelOptions: MEMBER_LEVEL_OPTIONS,
    statusFilterValues: STATUS_FILTER_VALUES,
    statusFilterLabels: STATUS_FILTER_LABELS,
    statusValues: STATUS_VALUES,
    statusLabels: STATUS_LABELS,
    levelFilterIndex: 0,
    statusFilterIndex: 0,
    loading: false,
    memberEditorVisible: false,
    memberEditorTitle: "新增会员",
    memberEditor: buildMemberEditor(),
    memberLevelIndex: 2,
    memberStatusIndex: 0
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
    listMembersOnline().then((result) => {
      let members = result && Array.isArray(result.items) ? result.items : [];
      members = members.map(normalizeMemberView);
      this.setData({
        members,
        stats: getMemberStats(members)
      });
      this.applyFilters();
    }).catch((error) => {
      this.setData({
        members: [],
        filteredMembers: [],
        selectedId: "",
        selectedMember: null,
        stats: getMemberStats([])
      });
      wx.showToast({ title: formatRequestError(error), icon: "none" });
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
      filteredMembers,
      selectedId,
      selectedMember
    });
  },

  onKeywordInput(event) {
    const filters = Object.assign({}, this.data.filters, {
      keyword: event.detail.value
    });
    this.setData({ filters });
    this.applyFilters();
  },

  onLevelChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      level: this.data.levelFilterValues[index]
    });
    this.setData({
      filters,
      levelFilterIndex: index
    });
    this.applyFilters();
  },

  onStatusChange(event) {
    const index = Number(event.detail.value);
    const filters = Object.assign({}, this.data.filters, {
      status: this.data.statusFilterValues[index]
    });
    this.setData({
      filters,
      statusFilterIndex: index
    });
    this.applyFilters();
  },

  selectMember(event) {
    const id = event.currentTarget.dataset.id;
    const selectedMember = this.data.filteredMembers.find(function(item) {
      return item.id === id;
    }) || null;
    this.setData({
      selectedId: id,
      selectedMember
    });
  },

  openCreateMember() {
    this.setData({
      memberEditorVisible: true,
      memberEditorTitle: "新增会员",
      memberEditor: buildMemberEditor(),
      memberLevelIndex: 2,
      memberStatusIndex: 0
    });
  },

  openEditMember() {
    const member = this.data.selectedMember;
    if (!member) return;
    this.setData({
      memberEditorVisible: true,
      memberEditorTitle: "编辑会员",
      memberEditor: buildMemberEditor(member),
      memberLevelIndex: Math.max(0, MEMBER_LEVEL_OPTIONS.indexOf(member.level)),
      memberStatusIndex: Math.max(0, STATUS_VALUES.indexOf(member.status))
    });
  },

  closeMemberEditor() {
    this.setData({ memberEditorVisible: false });
  },

  noop() {},

  onMemberEditorInput(event) {
    const key = event.currentTarget.dataset.key;
    const update = {};
    update[`memberEditor.${key}`] = event.detail.value;
    this.setData(update);
  },

  onMemberLevelChange(event) {
    const index = Number(event.detail.value);
    this.setData({
      memberLevelIndex: index,
      "memberEditor.level": MEMBER_LEVEL_OPTIONS[index] || "普通会员"
    });
  },

  onMemberStatusChange(event) {
    const index = Number(event.detail.value);
    this.setData({
      memberStatusIndex: index,
      "memberEditor.status": STATUS_VALUES[index] || "active"
    });
  },

  saveMemberEditor() {
    const editor = this.data.memberEditor;
    if (!editor.name) {
      wx.showToast({ title: "请填写会员姓名", icon: "none" });
      return;
    }
    if (!isValidPhone(editor.phone)) {
      wx.showToast({ title: "请输入正确手机号", icon: "none" });
      return;
    }
    if (this.data.loading) {
      return;
    }
    this.setData({ loading: true });
    saveMemberRecord(editor).then((savedMember) => {
      const normalizedMember = normalizeMemberView(savedMember);
      const members = [normalizedMember].concat(this.data.members.filter(function(item) {
        return item.id !== normalizedMember.id;
      }));
      this.setData({
        members,
        stats: getMemberStats(members),
        selectedId: normalizedMember.id,
        memberEditorVisible: false
      });
      this.applyFilters();
      wx.showToast({ title: "会员已保存", icon: "success" });
    }).catch((error) => {
      wx.showToast({ title: formatRequestError(error), icon: "none" });
    }).finally(() => {
      this.setData({ loading: false });
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

  reloadMembers() {
    if (this.data.loading) return;
    this.loadMembers();
  }
});
