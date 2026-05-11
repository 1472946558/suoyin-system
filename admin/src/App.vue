<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from "vue";
import {
  createProductRecord,
  createStoreRecord,
  createUserAccount,
  fetchAdminConsole,
  loginAdmin,
  saveProductRecord,
  savePrintTemplate,
  savePaymentCenterConfig,
  saveRoleTemplate,
  saveStoreRecord,
  saveSystemProfile,
  saveUserAccount,
  type AbilityCode,
  type ConsoleBootstrap,
  type DataScope,
  type DataSource,
  type PageId,
  type PaymentConfig,
  type PrintTemplate,
  type ProductRecord,
  type RoleTemplate,
  type StoreRecord,
  type SystemProfile,
  type UserAccount,
} from "./api";

type NoticeTone = "success" | "warning" | "info";

const SESSION_STORAGE_KEY = "gold_recycle_admin_session";

const PAGE_META: Record<
  PageId,
  {
    title: string;
    description: string;
    group: string;
    ability: AbilityCode;
    phase: "week1" | "phase2";
  }
> = {
  dashboard: {
    title: "后台首页",
    description: "经营概览、待办入口和支付/权限骨架状态。",
    group: "首页",
    ability: "dashboard.view",
    phase: "week1",
  },
  stores: {
    title: "门店管理",
    description: "门店状态、负责人、营业信息与设备位。",
    group: "组织与权限",
    ability: "store.view",
    phase: "week1",
  },
  users: {
    title: "账号管理",
    description: "后台账号、角色绑定、门店归属和启停用入口。",
    group: "组织与权限",
    ability: "user.view",
    phase: "week1",
  },
  roles: {
    title: "角色权限",
    description: "固定角色模板、能力项分组和数据范围配置。",
    group: "组织与权限",
    ability: "role.view",
    phase: "week1",
  },
  products: {
    title: "商品管理",
    description: "商品分类、SKU、价格与适用门店已进入后台首版。",
    group: "商品与业务",
    ability: "product.view",
    phase: "week1",
  },
  orders: {
    title: "订单管理",
    description: "收银订单列表、金额、支付状态与操作人信息。",
    group: "商品与业务",
    ability: "order.view",
    phase: "week1",
  },
  recycle: {
    title: "回收单管理",
    description: "回收单、照片张数、确认状态和客户信息。",
    group: "商品与业务",
    ability: "recycle.view",
    phase: "week1",
  },
  "payment-config": {
    title: "支付配置",
    description: "微信支付、现金记账和回调重试规则。",
    group: "支付中心",
    ability: "payment.config.view",
    phase: "week1",
  },
  "payment-records": {
    title: "支付流水",
    description: "流水筛选、异常核对和回调状态查看。",
    group: "支付中心",
    ability: "payment.record.view",
    phase: "week1",
  },
  "system-config": {
    title: "系统配置",
    description: "拍照规则、云资源状态、打印准备和品牌基础信息。",
    group: "系统与模板",
    ability: "system.config.view",
    phase: "week1",
  },
  "template-init": {
    title: "模板初始化",
    description: "新客户开通向导占位，表达模板化交付方向。",
    group: "系统与模板",
    ability: "template.init",
    phase: "week1",
  },
  audit: {
    title: "操作日志",
    description: "角色调整、支付配置修改和系统风险提醒审计。",
    group: "系统与模板",
    ability: "audit.view",
    phase: "week1",
  },
};

const GROUP_ORDER = ["首页", "组织与权限", "商品与业务", "支付中心", "系统与模板"];
const DATA_SCOPE_LABELS: Record<DataScope, string> = {
  all_stores: "全部门店",
  assigned_store: "本门店",
  self: "本人",
};
const DATA_SCOPE_HINTS: Record<DataScope, string> = {
  all_stores: "适用于老板主控台，可查看全门店汇总与配置项。",
  assigned_store: "适用于店长，只查看所属门店和可授权账号。",
  self: "适用于员工模板，第一版默认不开放后台。",
};
const PHASE_TWO_ITEMS: Record<PageId, string[]> = {
  dashboard: [],
  stores: [],
  users: [],
  roles: [],
  products: ["商品分类与 SKU 结构冻结", "价格字段和状态管理接口", "按门店维度的商品可见性规则"],
  orders: ["订单列表接口与详情抽屉", "支付状态回写与退款信息", "按门店的数据范围隔离"],
  recycle: ["回收单状态流定义", "图片附件和留痕校验", "支付流水与回收单关联"],
  "payment-config": [],
  "payment-records": [],
  "system-config": ["拍照规则字段清单", "编号规则和品牌配置", "系统配置修改审计"],
  "template-init": [],
  audit: ["高风险操作日志模型", "权限变更前后差异展示", "支付配置修改审计查询"],
};

function deepCopy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function formatCurrency(amount: number) {
  return `¥${amount.toLocaleString("zh-CN")}`;
}

function paymentMethodLabel(method: string) {
  return { wechat: "微信支付", cash: "现金记账", bank_transfer: "银行转账" }[method] || method;
}

function paymentStatusLabel(status: string) {
  return { paid: "已支付", refunding: "退款中", failed: "失败" }[status] || status;
}

function callbackStatusLabel(status: string) {
  return { delivered: "已回调", pending: "待补偿", exception: "异常" }[status] || status;
}

function productStatusLabel(status: string) {
  return { active: "上架中", draft: "待完善", disabled: "已停用" }[status] || status;
}

function orderStatusLabel(status: string) {
  return { paid: "已支付", pending: "待补偿", refunded: "已退款" }[status] || status;
}

function recycleStatusLabel(status: string) {
  return { draft: "待确认", confirmed: "已确认", cancelled: "已作废" }[status] || status;
}

function auditToneLabel(level: string) {
  return { high: "高风险", medium: "中风险", low: "低风险" }[level] || level;
}

function auditResultLabel(result: string) {
  return { success: "成功", warning: "提醒", info: "记录" }[result] || result;
}

function storeStatusLabel(status: string) {
  return { active: "营业中", pending: "待开业", disabled: "停用" }[status] || status;
}

function userStatusLabel(status: string) {
  return { enabled: "启用中", disabled: "已停用", invited: "待激活" }[status] || status;
}

function formatDateLabel(value: string) {
  return value.includes("T") ? value.replace("T", " ").slice(0, 16) : value;
}

function withinRange(dateText: string, range: string) {
  if (range === "all") return true;
  const normalized = dateText.includes("T") ? dateText : dateText.replace(" ", "T");
  const timestamp = new Date(normalized).getTime();
  const now = Date.now();
  const delta = now - timestamp;
  const day = 24 * 60 * 60 * 1000;
  if (range === "today") return delta <= day;
  if (range === "7d") return delta <= 7 * day;
  return delta <= 30 * day;
}

const bootstrap = ref<ConsoleBootstrap | null>(null);
const sourceMode = ref<DataSource>("mock");
const loading = ref(false);
const loginPending = ref(false);
const saveRolePending = ref(false);
const savePaymentPending = ref(false);
const saveStorePending = ref(false);
const saveUserPending = ref(false);
const saveProductPending = ref(false);
const saveSystemPending = ref(false);
const savePrintPending = ref(false);
const sessionToken = ref("");
const activePage = ref<PageId>("dashboard");
const loginError = ref("");
const notice = ref<{ tone: NoticeTone; text: string } | null>(null);
const selectedStoreId = ref("");
const selectedUserId = ref("");
const selectedRoleId = ref<number | null>(null);
const selectedRecordId = ref("");
const selectedProductId = ref("");
const selectedOrderId = ref("");
const selectedRecycleId = ref("");
const selectedAuditId = ref("");
const roleDraft = ref<RoleTemplate | null>(null);
const paymentDraft = ref<PaymentConfig | null>(null);
const storeDraft = ref<StoreRecord | null>(null);
const userDraft = ref<UserAccount | null>(null);
const productDraft = ref<ProductRecord | null>(null);
const systemDraft = ref<SystemProfile | null>(null);
const printDraft = ref<PrintTemplate | null>(null);

const loginForm = reactive({
  username: "boss",
  password: "Boss123!",
});

const storeFilters = reactive({
  keyword: "",
  status: "all",
});

const userFilters = reactive({
  keyword: "",
  roleKey: "all",
  status: "all",
});

const recordFilters = reactive({
  storeId: "all",
  method: "all",
  status: "all",
  callback: "all",
  range: "30d",
});

const productFilters = reactive({
  keyword: "",
  status: "all",
});

const orderFilters = reactive({
  keyword: "",
  status: "all",
});

const recycleFilters = reactive({
  status: "all",
});

const auditFilters = reactive({
  keyword: "",
  module: "all",
  risk: "all",
  result: "all",
});

const templateInitForm = reactive({
  tenantName: "金匠倌深圳门店包",
  ownerName: "周老板",
  city: "深圳",
  initialStoreCount: 1,
  enablePayment: true,
  enablePrint: true,
  requirePhotoAudit: true,
});

const currentUser = computed(() => bootstrap.value?.currentUser ?? null);
const abilityGroups = computed(() => bootstrap.value?.abilityGroups ?? []);
const currentPage = computed(() => PAGE_META[activePage.value]);
const dashboardData = computed(() => bootstrap.value?.dashboard ?? null);
const stores = computed(() => bootstrap.value?.stores ?? []);
const users = computed(() => bootstrap.value?.users ?? []);
const roles = computed(() => bootstrap.value?.roles ?? []);
const products = computed(() => bootstrap.value?.products ?? []);
const cashierOrders = computed(() => bootstrap.value?.cashierOrders ?? []);
const recycleOrders = computed(() => bootstrap.value?.recycleOrders ?? []);
const paymentRecords = computed(() => bootstrap.value?.paymentRecords ?? []);
const systemProfile = computed(() => bootstrap.value?.systemProfile ?? null);
const printTemplate = computed(() => bootstrap.value?.printTemplate ?? null);
const auditLogs = computed(() => bootstrap.value?.auditLogs ?? []);
const scopeOptions = computed(() =>
  (Object.entries(DATA_SCOPE_LABELS) as Array<[DataScope, string]>).map(([value, label]) => ({
    value,
    label,
    hint: DATA_SCOPE_HINTS[value],
  })),
);

const visibleNavigation = computed(() => {
  const items = (Object.entries(PAGE_META) as Array<[PageId, (typeof PAGE_META)[PageId]]>)
    .filter(([, meta]) => currentUser.value?.abilities.includes(meta.ability))
    .map(([id, meta]) => ({ id, ...meta }));

  return GROUP_ORDER.map((group) => ({
    group,
    items: items.filter((item) => item.group === group),
  })).filter((section) => section.items.length > 0);
});

const filteredStores = computed(() => {
  return stores.value.filter((store) => {
    const matchesKeyword =
      !storeFilters.keyword ||
      [store.name, store.code, store.managerName, store.city].some((field) =>
        field.toLowerCase().includes(storeFilters.keyword.toLowerCase()),
      );
    const matchesStatus = storeFilters.status === "all" || store.status === storeFilters.status;
    return matchesKeyword && matchesStatus;
  });
});

const filteredUsers = computed(() => {
  return users.value.filter((account) => {
    const matchesKeyword =
      !userFilters.keyword ||
      [account.name, account.account, account.roleName, account.storeNames.join(" / ")].some((field) =>
        field.toLowerCase().includes(userFilters.keyword.toLowerCase()),
      );
    const matchesRole = userFilters.roleKey === "all" || account.roleKey === userFilters.roleKey;
    const matchesStatus = userFilters.status === "all" || account.status === userFilters.status;
    return matchesKeyword && matchesRole && matchesStatus;
  });
});

const filteredRecords = computed(() => {
  return paymentRecords.value.filter((record) => {
    const matchesStore = recordFilters.storeId === "all" || record.storeId === recordFilters.storeId;
    const matchesMethod = recordFilters.method === "all" || record.method === recordFilters.method;
    const matchesStatus = recordFilters.status === "all" || record.status === recordFilters.status;
    const matchesCallback = recordFilters.callback === "all" || record.callbackStatus === recordFilters.callback;
    return matchesStore && matchesMethod && matchesStatus && matchesCallback && withinRange(record.paidAt, recordFilters.range);
  });
});

const filteredProducts = computed(() => {
  return products.value.filter((product) => {
    const matchesKeyword =
      !productFilters.keyword ||
      [product.name, product.sku, product.category, product.tags.join(" / ")].some((field) =>
        field.toLowerCase().includes(productFilters.keyword.toLowerCase()),
      );
    const matchesStatus = productFilters.status === "all" || product.status === productFilters.status;
    return matchesKeyword && matchesStatus;
  });
});

const filteredOrders = computed(() => {
  return cashierOrders.value.filter((order) => {
    const matchesKeyword =
      !orderFilters.keyword ||
      [order.orderNo, order.storeName, order.createdBy, order.remark].some((field) =>
        field.toLowerCase().includes(orderFilters.keyword.toLowerCase()),
      );
    const matchesStatus = orderFilters.status === "all" || order.status === orderFilters.status;
    return matchesKeyword && matchesStatus;
  });
});

const filteredRecycleOrders = computed(() => {
  return recycleOrders.value.filter((order) => {
    return recycleFilters.status === "all" || order.status === recycleFilters.status;
  });
});

const storeSummary = computed(() => {
  const totalAmount = filteredStores.value.reduce((sum, store) => sum + store.todayAmount, 0);
  const pendingCount = filteredStores.value.reduce((sum, store) => sum + store.pendingTasks, 0);
  const activeCount = filteredStores.value.filter((store) => store.status === "active").length;
  return { totalAmount, pendingCount, activeCount, count: filteredStores.value.length };
});

const userSummary = computed(() => {
  const enabled = filteredUsers.value.filter((account) => account.status === "enabled").length;
  const invited = filteredUsers.value.filter((account) => account.status === "invited").length;
  const ownerCount = filteredUsers.value.filter((account) => account.roleKey === "owner").length;
  return { enabled, invited, ownerCount, count: filteredUsers.value.length };
});

const productCatalogSummary = computed(() => {
  const active = filteredProducts.value.filter((product) => product.status === "active").length;
  const draft = filteredProducts.value.filter((product) => product.status === "draft").length;
  const totalValue = filteredProducts.value.reduce((sum, product) => sum + product.price, 0);
  return { active, draft, totalValue, count: filteredProducts.value.length };
});

const auditModuleOptions = computed(() => {
  return ["all", ...new Set(auditLogs.value.map((record) => record.module))];
});

const filteredAuditLogs = computed(() => {
  return auditLogs.value.filter((record) => {
    const matchesKeyword =
      !auditFilters.keyword ||
      [record.module, record.action, record.operatorName, record.summary].some((field) =>
        field.toLowerCase().includes(auditFilters.keyword.toLowerCase()),
      );
    const matchesModule = auditFilters.module === "all" || record.module === auditFilters.module;
    const matchesRisk = auditFilters.risk === "all" || record.riskLevel === auditFilters.risk;
    const matchesResult = auditFilters.result === "all" || record.result === auditFilters.result;
    return matchesKeyword && matchesModule && matchesRisk && matchesResult;
  });
});

const selectedStore = computed(() => {
  return filteredStores.value.find((store) => store.id === selectedStoreId.value) ?? filteredStores.value[0] ?? null;
});

const selectedUser = computed(() => {
  return filteredUsers.value.find((account) => account.id === selectedUserId.value) ?? filteredUsers.value[0] ?? null;
});

const selectedRole = computed(() => {
  return roles.value.find((role) => role.id === selectedRoleId.value) ?? roles.value[0] ?? null;
});

const selectedRecord = computed(() => {
  return filteredRecords.value.find((record) => record.id === selectedRecordId.value) ?? filteredRecords.value[0] ?? null;
});

const selectedProduct = computed(() => {
  return filteredProducts.value.find((product) => product.id === selectedProductId.value) ?? filteredProducts.value[0] ?? null;
});

const selectedOrder = computed(() => {
  return filteredOrders.value.find((order) => order.id === selectedOrderId.value) ?? filteredOrders.value[0] ?? null;
});

const selectedRecycleOrder = computed(() => {
  return (
    filteredRecycleOrders.value.find((order) => order.id === selectedRecycleId.value) ?? filteredRecycleOrders.value[0] ?? null
  );
});

const selectedAuditLog = computed(() => {
  return filteredAuditLogs.value.find((record) => record.id === selectedAuditId.value) ?? filteredAuditLogs.value[0] ?? null;
});

const rolePreviewPages = computed(() => {
  if (!roleDraft.value) return [];
  return (Object.entries(PAGE_META) as Array<[PageId, (typeof PAGE_META)[PageId]]>).filter(([, meta]) =>
    roleDraft.value?.abilities.includes(meta.ability),
  );
});

const paymentSummary = computed(() => {
  const total = filteredRecords.value.reduce((sum, record) => sum + record.amount, 0);
  const abnormal = filteredRecords.value.filter((record) => record.anomaly).length;
  const pending = filteredRecords.value.filter((record) => record.callbackStatus !== "delivered").length;
  return { total, abnormal, pending };
});

const orderSummary = computed(() => {
  const total = filteredOrders.value.reduce((sum, order) => sum + order.totalAmount, 0);
  const paid = filteredOrders.value.filter((order) => order.status === "paid").length;
  const pending = filteredOrders.value.filter((order) => order.status === "pending").length;
  return { total, paid, pending, count: filteredOrders.value.length };
});

const recycleSummary = computed(() => {
  const estimated = filteredRecycleOrders.value.reduce((sum, order) => sum + order.estimatedAmount, 0);
  const confirmed = filteredRecycleOrders.value.reduce((sum, order) => sum + order.confirmedAmount, 0);
  const confirmedCount = filteredRecycleOrders.value.filter((order) => order.status === "confirmed").length;
  return { estimated, confirmed, confirmedCount, count: filteredRecycleOrders.value.length };
});

const auditSummary = computed(() => {
  const total = filteredAuditLogs.value.length;
  const high = filteredAuditLogs.value.filter((record) => record.riskLevel === "high").length;
  const warnings = filteredAuditLogs.value.filter((record) => record.result === "warning").length;
  const modules = new Set(filteredAuditLogs.value.map((record) => record.module)).size;
  return { total, high, warnings, modules };
});

const templateProgress = computed(() => {
  const steps = bootstrap.value?.templateInit.steps ?? [];
  const done = steps.filter((step) => step.status === "done").length;
  const current = steps.find((step) => step.status === "current") ?? null;
  return {
    total: steps.length,
    done,
    current,
    percent: steps.length ? Math.round((done / steps.length) * 100) : 0,
  };
});

const templateOutputPreview = computed(() => {
  return [
    `${templateInitForm.tenantName} 初始化包`,
    `${templateInitForm.city} 首店模板 x${templateInitForm.initialStoreCount}`,
    templateInitForm.enablePayment ? "支付中心预置开启" : "支付中心后置接入",
    templateInitForm.enablePrint ? "小票与标签模板一并下发" : "打印模块暂缓开通",
    templateInitForm.requirePhotoAudit ? "回收拍照规则默认启用" : "拍照规则由门店自行调整",
  ];
});

const templateReady = computed(() => {
  return Boolean(templateInitForm.tenantName.trim() && templateInitForm.ownerName.trim() && templateInitForm.city.trim());
});

function setNotice(tone: NoticeTone, text: string) {
  notice.value = { tone, text };
}

function can(ability: AbilityCode) {
  return currentUser.value?.abilities.includes(ability) ?? false;
}

function ensureActivePage() {
  if (can(PAGE_META[activePage.value].ability)) return;
  const firstPage = visibleNavigation.value[0]?.items[0]?.id;
  activePage.value = firstPage ?? "dashboard";
}

function applyScopedDefaults() {
  const user = currentUser.value;
  if (!user) return;
  recordFilters.storeId = user.dataScope !== "all_stores" && user.storeIds[0] ? user.storeIds[0] : "all";
}

function persistSession(token: string) {
  localStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify({ token }));
}

function clearSession() {
  localStorage.removeItem(SESSION_STORAGE_KEY);
  sessionToken.value = "";
  bootstrap.value = null;
  activePage.value = "dashboard";
}

async function loadConsole(token: string) {
  loading.value = true;
  loginError.value = "";
  try {
    const result = await fetchAdminConsole(token);
    bootstrap.value = result.data;
    sessionToken.value = token;
    sourceMode.value = result.source;
    paymentDraft.value = deepCopy(result.data.paymentConfig);
    persistSession(token);
    applyScopedDefaults();
    ensureActivePage();
  } catch (error) {
    clearSession();
    loginError.value = error instanceof Error ? error.message : "加载后台数据失败。";
  } finally {
    loading.value = false;
  }
}

async function submitLogin() {
  loginPending.value = true;
  loginError.value = "";
  try {
    const result = await loginAdmin(loginForm.username.trim(), loginForm.password.trim());
    activePage.value = result.data.landingPage;
    await loadConsole(result.data.token);
  } catch (error) {
    loginError.value = error instanceof Error ? error.message : "登录失败，请稍后重试。";
  } finally {
    loginPending.value = false;
  }
}

function selectDemoAccount(username: string) {
  loginForm.username = username;
  loginForm.password = username === "boss" ? "Boss123!" : "Manager123!";
}

function logout() {
  clearSession();
  paymentDraft.value = null;
  roleDraft.value = null;
  notice.value = null;
}

function openPage(page: PageId) {
  if (can(PAGE_META[page].ability)) {
    activePage.value = page;
  }
}

function advanceTemplateInit() {
  if (!bootstrap.value || !templateReady.value) {
    setNotice("warning", "请先补齐客户包名称、负责人和城市，再推进初始化。");
    return;
  }
  const steps = deepCopy(bootstrap.value.templateInit.steps);
  const currentIndex = steps.findIndex((step) => step.status === "current");
  if (currentIndex >= 0) {
    steps[currentIndex].status = "done";
    if (steps[currentIndex + 1]) {
      steps[currentIndex + 1].status = "current";
    }
  } else {
    const plannedIndex = steps.findIndex((step) => step.status === "planned");
    if (plannedIndex >= 0) {
      steps[plannedIndex].status = "current";
    }
  }

  bootstrap.value = {
    ...bootstrap.value,
    templateInit: {
      ...bootstrap.value.templateInit,
      steps,
      outputs: templateOutputPreview.value,
    },
  };

  const nextStep = steps.find((step) => step.status === "current");
  if (nextStep) {
    setNotice("info", `初始化推进完成，下一步建议处理「${nextStep.title}」。`);
    return;
  }
  setNotice("success", "模板初始化步骤已全部推进完成，当前可进入门店与角色页做实装。");
}

function toggleRoleAbility(code: AbilityCode) {
  if (!roleDraft.value || !can("role.manage")) return;
  const abilitySet = new Set(roleDraft.value.abilities);
  if (abilitySet.has(code)) {
    abilitySet.delete(code);
  } else {
    abilitySet.add(code);
  }
  roleDraft.value.abilities = Array.from(abilitySet);
}

async function handleSaveRole() {
  if (!roleDraft.value || !bootstrap.value || !sessionToken.value || !can("role.manage")) return;
  if (!window.confirm("确认保存当前角色模板配置吗？真实联调后这里应接入审计日志。")) return;
  saveRolePending.value = true;
  try {
    const result = await saveRoleTemplate(sessionToken.value, roleDraft.value);
    bootstrap.value = {
      ...bootstrap.value,
      roles: bootstrap.value.roles.map((role) => (role.id === result.data.id ? deepCopy(result.data) : role)),
    };
    setNotice("success", `已保存 ${result.data.name}，当前仍为 ${result.source === "mock" ? "mock" : "API"} 模式。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "角色模板保存失败。");
  } finally {
    saveRolePending.value = false;
  }
}

async function handleSavePaymentConfig() {
  if (!paymentDraft.value || !bootstrap.value || !sessionToken.value || !can("payment.config.manage")) return;
  if (!window.confirm("确认更新支付配置吗？生产环境应在后端做二次校验与审计。")) return;
  savePaymentPending.value = true;
  try {
    const result = await savePaymentCenterConfig(sessionToken.value, paymentDraft.value);
    bootstrap.value = { ...bootstrap.value, paymentConfig: deepCopy(result.data) };
    paymentDraft.value = deepCopy(result.data);
    setNotice("success", `支付配置已保存，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "支付配置保存失败。");
  } finally {
    savePaymentPending.value = false;
  }
}

async function handleSaveStore() {
  if (!storeDraft.value || !sessionToken.value || !can("store.manage")) return;
  if (!window.confirm(`确认保存门店「${storeDraft.value.name}」的资料吗？`)) return;
  saveStorePending.value = true;
  try {
    const result = await saveStoreRecord(sessionToken.value, storeDraft.value);
    await loadConsole(sessionToken.value);
    selectedStoreId.value = result.data.id;
    setNotice("success", `门店资料已保存，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "门店资料保存失败。");
  } finally {
    saveStorePending.value = false;
  }
}

async function handleSaveUser() {
  if (!userDraft.value || !sessionToken.value || !can("user.manage")) return;
  if (!window.confirm(`确认保存账号「${userDraft.value.name}」的资料吗？`)) return;
  saveUserPending.value = true;
  try {
    const matchedRole = bootstrap.value?.roles.find((role) => role.key === userDraft.value?.roleKey);
    const nextUser: UserAccount = {
      ...userDraft.value,
      roleName:
        userDraft.value.roleKey === "owner"
          ? "老板"
          : userDraft.value.roleKey === "manager"
            ? "店长"
            : "员工",
      abilities: matchedRole ? [...matchedRole.abilities] : [...userDraft.value.abilities],
    };
    const result = await saveUserAccount(sessionToken.value, nextUser);
    await loadConsole(sessionToken.value);
    selectedUserId.value = result.data.id;
    setNotice("success", `账号资料已保存，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "账号资料保存失败。");
  } finally {
    saveUserPending.value = false;
  }
}

function toggleUserDraftStatus() {
  if (!userDraft.value || !can("user.manage")) return;
  userDraft.value.status = userDraft.value.status === "enabled" ? "disabled" : "enabled";
}

async function handleSaveProduct() {
  if (!productDraft.value || !sessionToken.value || !can("product.manage")) return;
  if (!window.confirm(`确认保存商品「${productDraft.value.name}」的资料吗？`)) return;
  saveProductPending.value = true;
  try {
    const result = await saveProductRecord(sessionToken.value, productDraft.value);
    await loadConsole(sessionToken.value);
    selectedProductId.value = result.data.id;
    setNotice("success", `商品资料已保存，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "商品资料保存失败。");
  } finally {
    saveProductPending.value = false;
  }
}

async function handleSaveSystem() {
  if (!systemDraft.value || !sessionToken.value || !can("system.config.manage")) return;
  if (!window.confirm("确认保存系统配置和打印准备信息吗？")) return;
  saveSystemPending.value = true;
  try {
    const result = await saveSystemProfile(sessionToken.value, systemDraft.value);
    await loadConsole(sessionToken.value);
    systemDraft.value = deepCopy(result.data);
    setNotice("success", `系统配置已保存，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "系统配置保存失败。");
  } finally {
    saveSystemPending.value = false;
  }
}

async function handleSavePrint() {
  if (!printDraft.value || !sessionToken.value || !can("system.config.manage")) return;
  if (!window.confirm("确认保存打印模板结构吗？")) return;
  savePrintPending.value = true;
  try {
    const result = await savePrintTemplate(sessionToken.value, printDraft.value);
    await loadConsole(sessionToken.value);
    printDraft.value = deepCopy(result.data);
    setNotice("success", `打印模板已保存，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "打印模板保存失败。");
  } finally {
    savePrintPending.value = false;
  }
}

async function handleCreateStore() {
  if (!sessionToken.value || !can("store.manage")) return;
  try {
    const result = await createStoreRecord(sessionToken.value);
    await loadConsole(sessionToken.value);
    selectedStoreId.value = result.data.id;
    setNotice("info", `已创建门店草稿，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "门店草稿创建失败。");
  }
}

async function handleCreateUser() {
  if (!sessionToken.value || !can("user.manage")) return;
  try {
    const result = await createUserAccount(sessionToken.value);
    await loadConsole(sessionToken.value);
    selectedUserId.value = result.data.id;
    setNotice("info", `已创建账号草稿，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "账号草稿创建失败。");
  }
}

async function handleCreateProduct() {
  if (!sessionToken.value || !can("product.manage")) return;
  try {
    const result = await createProductRecord(sessionToken.value);
    await loadConsole(sessionToken.value);
    selectedProductId.value = result.data.id;
    setNotice("info", `已创建商品草稿，当前来源：${result.source === "mock" ? "mock 回退" : "真实 API"}。`);
  } catch (error) {
    setNotice("warning", error instanceof Error ? error.message : "商品草稿创建失败。");
  }
}

watch(
  filteredProducts,
  (items) => {
    if (!items.some((item) => item.id === selectedProductId.value)) {
      selectedProductId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  filteredOrders,
  (items) => {
    if (!items.some((item) => item.id === selectedOrderId.value)) {
      selectedOrderId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  filteredRecycleOrders,
  (items) => {
    if (!items.some((item) => item.id === selectedRecycleId.value)) {
      selectedRecycleId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  filteredStores,
  (items) => {
    if (!items.some((item) => item.id === selectedStoreId.value)) {
      selectedStoreId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  selectedStore,
  (store) => {
    storeDraft.value = store ? deepCopy(store) : null;
  },
  { immediate: true },
);

watch(
  filteredUsers,
  (items) => {
    if (!items.some((item) => item.id === selectedUserId.value)) {
      selectedUserId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  selectedUser,
  (account) => {
    userDraft.value = account ? deepCopy(account) : null;
  },
  { immediate: true },
);

watch(
  filteredRecords,
  (items) => {
    if (!items.some((item) => item.id === selectedRecordId.value)) {
      selectedRecordId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  filteredAuditLogs,
  (items) => {
    if (!items.some((item) => item.id === selectedAuditId.value)) {
      selectedAuditId.value = items[0]?.id ?? "";
    }
  },
  { immediate: true },
);

watch(
  roles,
  (items) => {
    if (!items.length) {
      selectedRoleId.value = null;
      roleDraft.value = null;
      return;
    }
    if (!items.some((item) => item.id === selectedRoleId.value)) {
      selectedRoleId.value = items[0].id;
    }
  },
  { immediate: true },
);

watch(
  selectedRole,
  (role) => {
    roleDraft.value = role ? deepCopy(role) : null;
  },
  { immediate: true },
);

watch(
  selectedProduct,
  (product) => {
    productDraft.value = product ? deepCopy(product) : null;
  },
  { immediate: true },
);

watch(
  systemProfile,
  (profile) => {
    systemDraft.value = profile ? deepCopy(profile) : null;
  },
  { immediate: true },
);

watch(
  printTemplate,
  (profile) => {
    printDraft.value = profile ? deepCopy(profile) : null;
  },
  { immediate: true },
);

watch(activePage, () => {
  notice.value = null;
});

onMounted(async () => {
  const stored = localStorage.getItem(SESSION_STORAGE_KEY);
  if (!stored) return;
  try {
    const parsed = JSON.parse(stored) as { token?: string };
    if (parsed.token) {
      await loadConsole(parsed.token);
    }
  } catch {
    clearSession();
  }
});
</script>

<template>
  <div class="admin-app">
    <section v-if="!bootstrap" class="login-shell">
      <div class="login-hero panel">
        <p class="eyebrow">Week 1 Admin Skeleton</p>
        <h1>黄金回收收银系统后台管理端</h1>
        <p class="hero-copy">
          这是首版后台骨架，聚焦登录、首页、门店管理、账号管理、角色权限、支付配置、支付流水和模板初始化入口。
        </p>

        <div class="hero-grid">
          <article class="mini-card">
            <span>组织与权限</span>
            <strong>门店 / 账号 / 角色</strong>
            <p>固定角色模板 + 能力码 + 数据范围控制。</p>
          </article>
          <article class="mini-card">
            <span>支付中心</span>
            <strong>配置 / 流水</strong>
            <p>微信支付、现金记账、回调状态和异常复核占位。</p>
          </article>
          <article class="mini-card">
            <span>模板化交付</span>
            <strong>初始化入口</strong>
            <p>为后续新客户快速开通预留统一入口。</p>
          </article>
        </div>

        <div class="demo-box">
          <div>
            <h3>演示账号</h3>
            <p>当前优先连接真实后台接口，接口不可用时才回退到 mock 数据。</p>
          </div>
          <button type="button" class="secondary-btn" @click="selectDemoAccount('boss')">填充老板账号</button>
          <button type="button" class="secondary-btn" @click="selectDemoAccount('manager.sz')">填充店长账号</button>
        </div>
      </div>

      <div class="login-panel panel">
        <div class="panel-head">
          <div>
            <p class="eyebrow">Admin Login</p>
            <h2>后台登录</h2>
          </div>
          <span class="badge gold">Mock Ready</span>
        </div>

        <form class="form-grid" @submit.prevent="submitLogin">
          <label class="field">
            <span>账号</span>
            <input v-model="loginForm.username" type="text" autocomplete="username" placeholder="boss / manager.sz" />
          </label>

          <label class="field">
            <span>密码</span>
            <input v-model="loginForm.password" type="password" autocomplete="current-password" placeholder="Boss123! / Manager123!" />
          </label>

          <button class="primary-btn" type="submit" :disabled="loginPending || loading">
            {{ loginPending || loading ? "进入中..." : "进入后台" }}
          </button>
        </form>

        <p v-if="loginError" class="state-text error">{{ loginError }}</p>

        <div class="login-tips">
          <p><strong>老板：</strong>`boss / Boss123!`，可查看全门店和角色权限页。</p>
          <p><strong>店长：</strong>`manager.sz / Manager123!`，默认只看所属门店和支付流水。</p>
        </div>
      </div>
    </section>

    <div v-else class="workspace-shell">
      <aside class="sidebar panel">
        <div class="sidebar-brand">
          <div class="brand-mark">GC</div>
          <div>
            <strong>Gold Recycle Admin</strong>
            <p>黄金回收收银系统后台</p>
          </div>
        </div>

        <div class="sidebar-user">
          <span class="badge slate">{{ currentUser?.roleName }}</span>
          <h3>{{ currentUser?.name }}</h3>
          <p>{{ currentUser?.account }} · {{ currentUser ? DATA_SCOPE_LABELS[currentUser.dataScope] : "" }}</p>
        </div>

        <div v-for="section in visibleNavigation" :key="section.group" class="nav-group">
          <p class="nav-group-title">{{ section.group }}</p>
          <button
            v-for="item in section.items"
            :key="item.id"
            type="button"
            class="nav-item"
            :class="{ active: activePage === item.id }"
            @click="openPage(item.id)"
          >
            <span>{{ item.title }}</span>
            <small>{{ item.phase === "week1" ? "W1" : "P2" }}</small>
          </button>
        </div>

        <div class="sidebar-note">
          <p>当前模式</p>
          <strong>{{ sourceMode === "api" ? "真实 API" : "Mock / Fallback" }}</strong>
          <span>首版骨架优先表达信息架构，不依赖真实后端。</span>
        </div>
      </aside>

      <main class="workspace-main">
        <header class="workspace-header panel">
          <div>
            <p class="eyebrow">{{ currentPage.group }}</p>
            <h2>{{ currentPage.title }}</h2>
            <p class="header-copy">{{ currentPage.description }}</p>
          </div>

          <div class="header-actions">
            <span class="badge gold">{{ currentPage.phase === "week1" ? "Week 1" : "Phase 2" }}</span>
            <span class="badge slate">{{ currentUser ? DATA_SCOPE_LABELS[currentUser.dataScope] : "" }}</span>
            <span class="badge" :class="sourceMode === 'api' ? 'emerald' : 'gold'">
              {{ sourceMode === "api" ? "API" : "Mock" }}
            </span>
            <button type="button" class="secondary-btn" @click="logout">退出登录</button>
          </div>
        </header>

        <div v-if="notice" class="notice" :class="notice.tone">
          <span>{{ notice.text }}</span>
          <button type="button" class="text-btn" @click="notice = null">关闭</button>
        </div>

        <section v-if="activePage === 'dashboard' && dashboardData" class="page-grid">
          <article class="hero-banner panel">
            <div>
              <p class="eyebrow">Console Overview</p>
              <h3>{{ dashboardData.title }}</h3>
              <p>{{ dashboardData.subtitle }}</p>
            </div>
            <div class="hero-summary">
              <span>最近刷新：{{ bootstrap?.updatedAt }}</span>
              <span>可见门店：{{ stores.length }}</span>
              <span>可见账号：{{ users.length }}</span>
            </div>
          </article>

          <div class="stat-grid">
            <article v-for="metric in dashboardData.metrics" :key="metric.key" class="stat-card panel" :class="metric.tone">
              <span>{{ metric.label }}</span>
              <strong>{{ metric.value }}</strong>
              <p>{{ metric.delta }}</p>
            </article>
          </div>

          <div class="content-grid two-up">
            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">Todo Queue</p>
                  <h3>待处理事项</h3>
                </div>
              </div>

              <div class="stack-list">
                <article v-for="todo in dashboardData.todos" :key="todo.id" class="list-card">
                  <div class="list-card-head">
                    <strong>{{ todo.title }}</strong>
                    <span class="badge" :class="todo.level === 'high' ? 'danger' : 'slate'">
                      {{ todo.level === "high" ? "高优" : "处理中" }}
                    </span>
                  </div>
                  <p>{{ todo.description }}</p>
                  <button type="button" class="text-btn" @click="openPage(todo.page)">前往处理</button>
                </article>
              </div>
            </article>

            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">Shortcuts</p>
                  <h3>快捷入口</h3>
                </div>
              </div>

              <div class="stack-list">
                <button
                  v-for="shortcut in dashboardData.shortcuts"
                  :key="shortcut.id"
                  type="button"
                  class="shortcut-card"
                  @click="openPage(shortcut.page)"
                >
                  <strong>{{ shortcut.label }}</strong>
                  <span>{{ shortcut.hint }}</span>
                </button>
              </div>
            </article>
          </div>

          <div class="content-grid two-up">
            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">Store Snapshot</p>
                  <h3>门店经营快照</h3>
                </div>
              </div>

              <div class="stack-list">
                <article v-for="store in stores.slice(0, 3)" :key="store.id" class="list-card">
                  <div class="list-card-head">
                    <strong>{{ store.name }}</strong>
                    <span class="badge" :class="store.status === 'active' ? 'emerald' : store.status === 'pending' ? 'gold' : 'danger'">
                      {{ storeStatusLabel(store.status) }}
                    </span>
                  </div>
                  <p>{{ store.city }} · {{ store.managerName }} · {{ store.businessHours }}</p>
                  <div class="inline-metrics">
                    <span>今日收银 {{ formatCurrency(store.todayAmount) }}</span>
                    <span>订单 {{ store.todayOrders }}</span>
                    <span>待办 {{ store.pendingTasks }}</span>
                  </div>
                </article>
              </div>
            </article>

            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">Readiness</p>
                  <h3>支付与联调准备</h3>
                </div>
              </div>

              <div class="stack-list">
                <article class="list-card">
                  <strong>微信支付</strong>
                  <p>{{ bootstrap?.paymentConfig.wechat.certificateStatus }}</p>
                </article>
                <article class="list-card">
                  <strong>支付异常</strong>
                  <p>当前可见流水中有 {{ paymentSummary.abnormal }} 笔异常/待补偿记录。</p>
                </article>
                <article v-for="noticeItem in dashboardData.notices" :key="noticeItem" class="list-card">
                  <p>{{ noticeItem }}</p>
                </article>
              </div>
            </article>
          </div>
        </section>

        <section v-else-if="activePage === 'stores'" class="page-grid">
          <article class="panel">
            <div class="toolbar">
              <label class="field grow">
                <span>搜索门店</span>
                <input v-model="storeFilters.keyword" type="text" placeholder="门店名 / 编码 / 负责人 / 城市" />
              </label>
              <label class="field compact">
                <span>状态</span>
                <select v-model="storeFilters.status">
                  <option value="all">全部</option>
                  <option value="active">营业中</option>
                  <option value="pending">待开业</option>
                  <option value="disabled">停用</option>
                </select>
              </label>
              <button type="button" class="primary-btn" :disabled="!can('store.manage')" @click="handleCreateStore">新建门店</button>
            </div>

            <div class="summary-row">
              <div class="summary-pill">门店 {{ storeSummary.count }} 家</div>
              <div class="summary-pill">营业中 {{ storeSummary.activeCount }} 家</div>
              <div class="summary-pill">待办 {{ storeSummary.pendingCount }} 项</div>
              <div class="summary-pill">今日收银 {{ formatCurrency(storeSummary.totalAmount) }}</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>门店</th>
                      <th>负责人</th>
                      <th>营业信息</th>
                      <th>今日收银</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="store in filteredStores"
                      :key="store.id"
                      :class="{ selected: selectedStore?.id === store.id }"
                      @click="selectedStoreId = store.id"
                    >
                      <td>
                        <strong>{{ store.name }}</strong>
                        <span>{{ store.code }}</span>
                      </td>
                      <td>{{ store.managerName }}</td>
                      <td>{{ store.city }} · {{ store.businessHours }}</td>
                      <td>{{ formatCurrency(store.todayAmount) }}</td>
                      <td>
                        <span class="badge" :class="store.status === 'active' ? 'emerald' : store.status === 'pending' ? 'gold' : 'danger'">
                          {{ storeStatusLabel(store.status) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <aside v-if="storeDraft" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Store Detail</p>
                    <h3>{{ storeDraft.name }}</h3>
                  </div>
                  <span class="badge" :class="storeDraft.status === 'active' ? 'emerald' : storeDraft.status === 'pending' ? 'gold' : 'danger'">
                    {{ storeStatusLabel(storeDraft.status) }}
                  </span>
                </div>

                <div class="form-card">
                  <label class="field">
                    <span>门店名称</span>
                    <input v-model="storeDraft.name" type="text" :disabled="!can('store.manage')" />
                  </label>
                  <label class="field">
                    <span>负责人</span>
                    <input v-model="storeDraft.managerName" type="text" :disabled="!can('store.manage')" />
                  </label>
                  <label class="field">
                    <span>联系电话</span>
                    <input v-model="storeDraft.contactPhone" type="text" :disabled="!can('store.manage')" />
                  </label>
                  <label class="field">
                    <span>营业时间</span>
                    <input v-model="storeDraft.businessHours" type="text" :disabled="!can('store.manage')" />
                  </label>
                  <label class="field">
                    <span>门店状态</span>
                    <select v-model="storeDraft.status" :disabled="!can('store.manage')">
                      <option value="active">营业中</option>
                      <option value="pending">待开业</option>
                      <option value="disabled">停用</option>
                    </select>
                  </label>
                  <label class="field">
                    <span>设备位</span>
                    <input v-model.number="storeDraft.cashierDevices" type="number" min="0" :disabled="!can('store.manage')" />
                  </label>
                  <label class="field">
                    <span>地址</span>
                    <textarea v-model="storeDraft.address" rows="3" :disabled="!can('store.manage')"></textarea>
                  </label>
                </div>

                <div class="chip-row">
                  <span v-for="tag in storeDraft.tags" :key="tag" class="chip">{{ tag }}</span>
                </div>

                <div class="summary-row">
                  <div class="summary-pill">设备位 {{ storeDraft.cashierDevices }}</div>
                  <div class="summary-pill">订单 {{ storeDraft.todayOrders }}</div>
                  <div class="summary-pill">待办 {{ storeDraft.pendingTasks }}</div>
                </div>

                <div class="list-card">
                  <strong>经营提示</strong>
                  <p v-if="storeDraft.status === 'pending'">当前门店仍处于待开业状态，建议补齐负责人、设备位和营业资料。</p>
                  <p v-else-if="storeDraft.status === 'disabled'">当前门店已停用，建议复核历史订单、账号归属和设备状态。</p>
                  <p v-else>当前门店处于营业中，可继续核对支付配置、人员归属和营业时间口径。</p>
                </div>

                <div class="detail-actions">
                  <button type="button" class="primary-btn" :disabled="!can('store.manage') || saveStorePending" @click="handleSaveStore">
                    {{ saveStorePending ? "保存中..." : "保存门店资料" }}
                  </button>
                  <button type="button" class="secondary-btn" :disabled="!can('user.view')" @click="activePage = 'users'">查看门店账号</button>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'users'" class="page-grid">
          <article class="panel">
            <div class="toolbar">
              <label class="field grow">
                <span>搜索账号</span>
                <input v-model="userFilters.keyword" type="text" placeholder="姓名 / 账号 / 角色 / 门店" />
              </label>
              <label class="field compact">
                <span>角色</span>
                <select v-model="userFilters.roleKey">
                  <option value="all">全部</option>
                  <option value="owner">老板</option>
                  <option value="manager">店长</option>
                  <option value="clerk">员工</option>
                </select>
              </label>
              <label class="field compact">
                <span>状态</span>
                <select v-model="userFilters.status">
                  <option value="all">全部</option>
                  <option value="enabled">启用中</option>
                  <option value="invited">待激活</option>
                  <option value="disabled">已停用</option>
                </select>
              </label>
              <button type="button" class="primary-btn" :disabled="!can('user.manage')" @click="handleCreateUser">新建账号</button>
            </div>

            <div class="summary-row">
              <div class="summary-pill">账号 {{ userSummary.count }} 个</div>
              <div class="summary-pill">启用中 {{ userSummary.enabled }} 个</div>
              <div class="summary-pill">待激活 {{ userSummary.invited }} 个</div>
              <div class="summary-pill">老板模板 {{ userSummary.ownerCount }} 个</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>账号</th>
                      <th>角色模板</th>
                      <th>门店归属</th>
                      <th>数据范围</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="account in filteredUsers"
                      :key="account.id"
                      :class="{ selected: selectedUser?.id === account.id }"
                      @click="selectedUserId = account.id"
                    >
                      <td>
                        <strong>{{ account.name }}</strong>
                        <span>{{ account.account }} · {{ account.phone }}</span>
                      </td>
                      <td>{{ account.roleName }}</td>
                      <td>{{ account.storeNames.join(" / ") }}</td>
                      <td>{{ DATA_SCOPE_LABELS[account.dataScope] }}</td>
                      <td>
                        <span class="badge" :class="account.status === 'enabled' ? 'emerald' : account.status === 'invited' ? 'gold' : 'danger'">
                          {{ userStatusLabel(account.status) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <aside v-if="userDraft" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Account Detail</p>
                    <h3>{{ userDraft.name }}</h3>
                  </div>
                  <span class="badge" :class="userDraft.status === 'enabled' ? 'emerald' : userDraft.status === 'invited' ? 'gold' : 'danger'">
                    {{ userStatusLabel(userDraft.status) }}
                  </span>
                </div>

                <div class="form-card">
                  <label class="field">
                    <span>姓名</span>
                    <input v-model="userDraft.name" type="text" :disabled="!can('user.manage')" />
                  </label>
                  <label class="field">
                    <span>账号</span>
                    <input v-model="userDraft.account" type="text" :disabled="!can('user.manage')" />
                  </label>
                  <label class="field">
                    <span>手机号</span>
                    <input v-model="userDraft.phone" type="text" :disabled="!can('user.manage')" />
                  </label>
                  <label class="field">
                    <span>角色</span>
                    <select v-model="userDraft.roleKey" :disabled="!can('user.manage')">
                      <option value="owner">老板</option>
                      <option value="manager">店长</option>
                      <option value="clerk">员工</option>
                    </select>
                  </label>
                  <label class="field">
                    <span>状态</span>
                    <select v-model="userDraft.status" :disabled="!can('user.manage')">
                      <option value="enabled">启用中</option>
                      <option value="invited">待激活</option>
                      <option value="disabled">已停用</option>
                    </select>
                  </label>
                  <label class="field">
                    <span>数据范围</span>
                    <select v-model="userDraft.dataScope" :disabled="!can('user.manage')">
                      <option value="all_stores">全部门店</option>
                      <option value="assigned_store">本门店</option>
                      <option value="self">本人</option>
                    </select>
                  </label>
                </div>

                <div class="chip-row">
                  <span v-for="ability in userDraft.abilities" :key="ability" class="chip">{{ ability }}</span>
                </div>

                <div class="summary-row">
                  <div class="summary-pill">{{ userDraft.roleName }}</div>
                  <div class="summary-pill">{{ DATA_SCOPE_LABELS[userDraft.dataScope] }}</div>
                  <div class="summary-pill">能力 {{ userDraft.abilities.length }} 项</div>
                </div>

                <div class="list-card">
                  <strong>账号提示</strong>
                  <p v-if="userDraft.status === 'invited'">当前账号还未激活，建议确认门店归属和初始角色后再通知使用。</p>
                  <p v-else-if="userDraft.status === 'disabled'">当前账号已停用，建议同步排查历史操作记录与权限残留。</p>
                  <p v-else>当前账号已启用，可继续核对角色模板、门店可见范围和末次登录时间。</p>
                </div>

                <div class="detail-actions">
                  <button type="button" class="secondary-btn" :disabled="!can('user.manage')" @click="toggleUserDraftStatus()">
                    {{ userDraft.status === "enabled" ? "切换为停用" : "切换为启用" }}
                  </button>
                  <button type="button" class="primary-btn" :disabled="!can('user.manage') || saveUserPending" @click="handleSaveUser">
                    {{ saveUserPending ? "保存中..." : "保存账号资料" }}
                  </button>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'roles'" class="page-grid">
          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">Role Templates</p>
                <h3>固定角色模板</h3>
              </div>
              <span class="badge gold">不可新增 / 删除，仅调整能力项</span>
            </div>

            <div class="role-grid">
              <button
                v-for="role in roles"
                :key="role.id"
                type="button"
                class="role-card"
                :class="{ active: roleDraft?.id === role.id }"
                @click="selectedRoleId = role.id"
              >
                <strong>{{ role.name }}</strong>
                <span>{{ role.description }}</span>
                <small>{{ DATA_SCOPE_LABELS[role.dataScope] }} · {{ role.memberCount }} 个账号</small>
              </button>
            </div>
          </article>

          <div v-if="roleDraft" class="content-grid two-up">
            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">Ability Editor</p>
                  <h3>{{ roleDraft.name }}</h3>
                </div>
              </div>

              <div class="scope-switch">
                <label v-for="option in scopeOptions" :key="option.value" class="scope-option">
                  <input v-model="roleDraft.dataScope" type="radio" :value="option.value" :disabled="!can('role.manage')" />
                  <div>
                    <strong>{{ option.label }}</strong>
                    <span>{{ option.hint }}</span>
                  </div>
                </label>
              </div>

              <div class="ability-stack">
                <section v-for="group in abilityGroups" :key="group.key" class="ability-group">
                  <header>
                    <strong>{{ group.label }}</strong>
                    <p>{{ group.description }}</p>
                  </header>
                  <label v-for="item in group.items" :key="item.code" class="ability-item">
                    <input
                      type="checkbox"
                      :checked="roleDraft.abilities.includes(item.code)"
                      :disabled="!can('role.manage')"
                      @change="toggleRoleAbility(item.code)"
                    />
                    <div>
                      <strong>{{ item.label }}</strong>
                      <span>{{ item.description }}</span>
                    </div>
                  </label>
                </section>
              </div>

              <button type="button" class="primary-btn" :disabled="!can('role.manage') || saveRolePending" @click="handleSaveRole">
                {{ saveRolePending ? "保存中..." : "保存角色模板" }}
              </button>
            </article>

            <aside class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">Preview</p>
                  <h3>影响模块预览</h3>
                </div>
              </div>

              <div class="chip-row">
                <span v-for="[pageId, meta] in rolePreviewPages" :key="pageId" class="chip">{{ meta.title }}</span>
              </div>

              <div class="list-card">
                <strong>当前数据范围</strong>
                <p>{{ DATA_SCOPE_HINTS[roleDraft.dataScope] }}</p>
              </div>

              <div class="list-card">
                <strong>风险提示</strong>
                <p>账号停用、权限变更、支付配置修改后续都应由后端写入操作日志。</p>
              </div>
            </aside>
          </div>
        </section>

        <section v-else-if="activePage === 'products'" class="page-grid">
          <article class="panel">
            <div class="toolbar">
              <label class="field grow">
                <span>搜索商品</span>
                <input v-model="productFilters.keyword" type="text" placeholder="商品名 / SKU / 分类 / 标签" />
              </label>
              <label class="field compact">
                <span>状态</span>
                <select v-model="productFilters.status">
                  <option value="all">全部</option>
                  <option value="active">上架中</option>
                  <option value="draft">待完善</option>
                  <option value="disabled">已停用</option>
                </select>
              </label>
              <button type="button" class="primary-btn" :disabled="!can('product.manage')" @click="handleCreateProduct">新建商品</button>
            </div>

            <div class="summary-row">
              <div class="summary-pill">商品 {{ productCatalogSummary.count }} 个</div>
              <div class="summary-pill">上架中 {{ productCatalogSummary.active }} 个</div>
              <div class="summary-pill">待完善 {{ productCatalogSummary.draft }} 个</div>
              <div class="summary-pill">总价盘 {{ formatCurrency(productCatalogSummary.totalValue) }}</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>商品</th>
                      <th>分类</th>
                      <th>克重</th>
                      <th>售价</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="product in filteredProducts"
                      :key="product.id"
                      :class="{ selected: selectedProduct?.id === product.id }"
                      @click="selectedProductId = product.id"
                    >
                      <td>
                        <strong>{{ product.name }}</strong>
                        <span>{{ product.sku }}</span>
                      </td>
                      <td>{{ product.category }}</td>
                      <td>{{ product.gramWeight > 0 ? `${product.gramWeight} g` : "无" }}</td>
                      <td>{{ formatCurrency(product.price) }}</td>
                      <td>
                        <span class="badge" :class="product.status === 'active' ? 'emerald' : product.status === 'draft' ? 'gold' : 'danger'">
                          {{ productStatusLabel(product.status) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <aside v-if="productDraft" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Product Detail</p>
                    <h3>{{ productDraft.name }}</h3>
                  </div>
                  <span class="badge" :class="productDraft.status === 'active' ? 'emerald' : productDraft.status === 'draft' ? 'gold' : 'danger'">
                    {{ productStatusLabel(productDraft.status) }}
                  </span>
                </div>

                <div class="form-card">
                  <label class="field">
                    <span>商品名称</span>
                    <input v-model="productDraft.name" type="text" :disabled="!can('product.manage')" />
                  </label>
                  <label class="field">
                    <span>SKU</span>
                    <input v-model="productDraft.sku" type="text" :disabled="!can('product.manage')" />
                  </label>
                  <label class="field">
                    <span>分类</span>
                    <input v-model="productDraft.category" type="text" :disabled="!can('product.manage')" />
                  </label>
                  <label class="field">
                    <span>售价</span>
                    <input v-model.number="productDraft.price" type="number" min="0" :disabled="!can('product.manage')" />
                  </label>
                  <label class="field">
                    <span>克重</span>
                    <input v-model.number="productDraft.gramWeight" type="number" min="0" step="0.01" :disabled="!can('product.manage')" />
                  </label>
                  <label class="field">
                    <span>状态</span>
                    <select v-model="productDraft.status" :disabled="!can('product.manage')">
                      <option value="active">上架中</option>
                      <option value="draft">待完善</option>
                      <option value="disabled">已停用</option>
                    </select>
                  </label>
                </div>

                <div class="chip-row">
                  <span v-for="tag in productDraft.tags" :key="tag" class="chip">{{ tag }}</span>
                </div>

                <div class="summary-row">
                  <div class="summary-pill">{{ productDraft.category }}</div>
                  <div class="summary-pill">{{ productDraft.gramWeight > 0 ? `${productDraft.gramWeight} g` : "无克重" }}</div>
                  <div class="summary-pill">{{ formatCurrency(productDraft.price) }}</div>
                </div>

                <div class="list-card">
                  <strong>商品提示</strong>
                  <p v-if="productDraft.status === 'draft'">当前商品仍待完善，建议补齐价格、分类和适用场景后再上架。</p>
                  <p v-else-if="productDraft.status === 'disabled'">当前商品已停用，建议核对历史订单引用和门店展示状态。</p>
                  <p v-else>当前商品已上架，可继续核对 SKU、克重模板和门店可见性规则。</p>
                </div>

                <div class="detail-actions">
                  <button type="button" class="primary-btn" :disabled="!can('product.manage') || saveProductPending" @click="handleSaveProduct">
                    {{ saveProductPending ? "保存中..." : "保存商品资料" }}
                  </button>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'orders'" class="page-grid">
          <article class="panel">
            <div class="toolbar">
              <label class="field grow">
                <span>搜索订单</span>
                <input v-model="orderFilters.keyword" type="text" placeholder="订单号 / 门店 / 操作人 / 备注" />
              </label>
              <label class="field compact">
                <span>状态</span>
                <select v-model="orderFilters.status">
                  <option value="all">全部</option>
                  <option value="paid">已支付</option>
                  <option value="pending">待补偿</option>
                  <option value="refunded">已退款</option>
                </select>
              </label>
            </div>

            <div class="summary-row">
              <div class="summary-pill">订单 {{ orderSummary.count }} 笔</div>
              <div class="summary-pill">已支付 {{ orderSummary.paid }} 笔</div>
              <div class="summary-pill">待补偿 {{ orderSummary.pending }} 笔</div>
              <div class="summary-pill">金额 {{ formatCurrency(orderSummary.total) }}</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>订单号</th>
                      <th>门店</th>
                      <th>支付方式</th>
                      <th>金额</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="order in filteredOrders"
                      :key="order.id"
                      :class="{ selected: selectedOrder?.id === order.id }"
                      @click="selectedOrderId = order.id"
                    >
                      <td>
                        <strong>{{ order.orderNo }}</strong>
                        <span>{{ formatDateLabel(order.createdAt) }}</span>
                      </td>
                      <td>{{ order.storeName }}</td>
                      <td>{{ paymentMethodLabel(order.paymentMethod) }}</td>
                      <td>{{ formatCurrency(order.totalAmount) }}</td>
                      <td>
                        <span class="badge" :class="order.status === 'paid' ? 'emerald' : order.status === 'pending' ? 'gold' : 'danger'">
                          {{ orderStatusLabel(order.status) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <aside v-if="selectedOrder" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Order Detail</p>
                    <h3>{{ selectedOrder.orderNo }}</h3>
                  </div>
                  <span class="badge" :class="selectedOrder.status === 'paid' ? 'emerald' : selectedOrder.status === 'pending' ? 'gold' : 'danger'">
                    {{ orderStatusLabel(selectedOrder.status) }}
                  </span>
                </div>

                <div class="detail-grid">
                  <p><strong>创建时间：</strong>{{ formatDateLabel(selectedOrder.createdAt) }}</p>
                  <p><strong>门店：</strong>{{ selectedOrder.storeName }}</p>
                  <p><strong>金额：</strong>{{ formatCurrency(selectedOrder.totalAmount) }}</p>
                  <p><strong>商品数：</strong>{{ selectedOrder.itemCount }}</p>
                  <p><strong>支付方式：</strong>{{ paymentMethodLabel(selectedOrder.paymentMethod) }}</p>
                  <p><strong>操作人：</strong>{{ selectedOrder.createdBy }}</p>
                  <p><strong>备注：</strong>{{ selectedOrder.remark }}</p>
                </div>

                <div class="summary-row">
                  <div class="summary-pill">件数 {{ selectedOrder.itemCount }}</div>
                  <div class="summary-pill">{{ paymentMethodLabel(selectedOrder.paymentMethod) }}</div>
                  <div class="summary-pill">{{ orderStatusLabel(selectedOrder.status) }}</div>
                </div>

                <div class="list-card">
                  <strong>处理建议</strong>
                  <p v-if="selectedOrder.status === 'pending'">当前订单仍待补偿或待核对，建议联动支付流水页检查回调状态和备注。</p>
                  <p v-else-if="selectedOrder.status === 'refunded'">当前订单已退款，建议同步核对支付中心异常记录和门店口径。</p>
                  <p v-else>当前订单已支付，可继续核对门店、商品数与操作人留痕是否完整。</p>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'recycle'" class="page-grid">
          <article class="panel">
            <div class="toolbar">
              <label class="field compact">
                <span>状态</span>
                <select v-model="recycleFilters.status">
                  <option value="all">全部</option>
                  <option value="draft">待确认</option>
                  <option value="confirmed">已确认</option>
                  <option value="cancelled">已作废</option>
                </select>
              </label>
            </div>

            <div class="summary-row">
              <div class="summary-pill">回收单 {{ recycleSummary.count }} 笔</div>
              <div class="summary-pill">已确认 {{ recycleSummary.confirmedCount }} 笔</div>
              <div class="summary-pill">预估 {{ formatCurrency(recycleSummary.estimated) }}</div>
              <div class="summary-pill">确认 {{ formatCurrency(recycleSummary.confirmed) }}</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>回收单</th>
                      <th>客户</th>
                      <th>门店</th>
                      <th>照片</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="order in filteredRecycleOrders"
                      :key="order.id"
                      :class="{ selected: selectedRecycleOrder?.id === order.id }"
                      @click="selectedRecycleId = order.id"
                    >
                      <td>
                        <strong>{{ order.orderNo }}</strong>
                        <span>{{ formatDateLabel(order.createdAt) }}</span>
                      </td>
                      <td>{{ order.customerName }} · {{ order.customerPhone }}</td>
                      <td>{{ order.storeName }}</td>
                      <td>{{ order.photoCount }} 张</td>
                      <td>
                        <span class="badge" :class="order.status === 'confirmed' ? 'emerald' : order.status === 'draft' ? 'gold' : 'danger'">
                          {{ recycleStatusLabel(order.status) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <aside v-if="selectedRecycleOrder" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Recycle Detail</p>
                    <h3>{{ selectedRecycleOrder.orderNo }}</h3>
                  </div>
                  <span class="badge" :class="selectedRecycleOrder.status === 'confirmed' ? 'emerald' : selectedRecycleOrder.status === 'draft' ? 'gold' : 'danger'">
                    {{ recycleStatusLabel(selectedRecycleOrder.status) }}
                  </span>
                </div>

                <div class="detail-grid">
                  <p><strong>创建时间：</strong>{{ formatDateLabel(selectedRecycleOrder.createdAt) }}</p>
                  <p><strong>确认时间：</strong>{{ selectedRecycleOrder.confirmedAt ? formatDateLabel(selectedRecycleOrder.confirmedAt) : '待确认' }}</p>
                  <p><strong>客户：</strong>{{ selectedRecycleOrder.customerName }}</p>
                  <p><strong>联系电话：</strong>{{ selectedRecycleOrder.customerPhone }}</p>
                  <p><strong>预估金额：</strong>{{ formatCurrency(selectedRecycleOrder.estimatedAmount) }}</p>
                  <p><strong>确认金额：</strong>{{ formatCurrency(selectedRecycleOrder.confirmedAmount) }}</p>
                  <p><strong>照片数：</strong>{{ selectedRecycleOrder.photoCount }} 张</p>
                  <p><strong>操作人：</strong>{{ selectedRecycleOrder.createdBy }}</p>
                  <p><strong>备注：</strong>{{ selectedRecycleOrder.remark }}</p>
                </div>

                <div class="summary-row">
                  <div class="summary-pill">照片 {{ selectedRecycleOrder.photoCount }} 张</div>
                  <div class="summary-pill">{{ recycleStatusLabel(selectedRecycleOrder.status) }}</div>
                  <div class="summary-pill">差额 {{ formatCurrency(selectedRecycleOrder.confirmedAmount - selectedRecycleOrder.estimatedAmount) }}</div>
                </div>

                <div class="list-card">
                  <strong>留痕检查</strong>
                  <p v-if="selectedRecycleOrder.photoCount < 2">当前照片数偏少，建议补齐现场照片再做确认。</p>
                  <p v-else-if="selectedRecycleOrder.status === 'draft'">当前仍为待确认状态，建议尽快完成金额确认与客户签收。</p>
                  <p v-else>当前回收单留痕基本完整，可继续联动支付流水和打印模板验收。</p>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'payment-config' && paymentDraft" class="page-grid">
          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">Payment Setup</p>
                <h3>支付配置</h3>
              </div>
              <span class="badge gold">首版静态表单 + 本地保存</span>
            </div>

            <div class="content-grid two-up">
              <div class="form-card">
                <label class="field">
                  <span>微信支付开关</span>
                  <select v-model="paymentDraft.wechat.enabled">
                    <option :value="true">开启</option>
                    <option :value="false">关闭</option>
                  </select>
                </label>
                <label class="field">
                  <span>AppID</span>
                  <input v-model="paymentDraft.wechat.appId" type="text" />
                </label>
                <label class="field">
                  <span>商户号</span>
                  <input v-model="paymentDraft.wechat.merchantId" type="text" />
                </label>
                <label class="field">
                  <span>子商户号</span>
                  <input v-model="paymentDraft.wechat.subMerchantId" type="text" />
                </label>
                <label class="field">
                  <span>回调地址</span>
                  <input v-model="paymentDraft.wechat.callbackUrl" type="text" />
                </label>
                <label class="field">
                  <span>证书状态</span>
                  <input v-model="paymentDraft.wechat.certificateStatus" type="text" />
                </label>
                <label class="toggle-field">
                  <input v-model="paymentDraft.wechat.sandboxMode" type="checkbox" />
                  <span>启用沙箱模式</span>
                </label>
              </div>

              <div class="form-card">
                <label class="toggle-field">
                  <input v-model="paymentDraft.cash.enabled" type="checkbox" />
                  <span>开启现金记账</span>
                </label>
                <label class="toggle-field">
                  <input v-model="paymentDraft.cash.receiptRequired" type="checkbox" />
                  <span>现金单必须生成小票</span>
                </label>
                <label class="toggle-field">
                  <input v-model="paymentDraft.cash.shiftReconciliationRequired" type="checkbox" />
                  <span>要求班次交接结算</span>
                </label>
                <label class="toggle-field">
                  <input v-model="paymentDraft.bankTransfer.enabled" type="checkbox" />
                  <span>开放银行转账</span>
                </label>
                <label class="field">
                  <span>对公账户名</span>
                  <input v-model="paymentDraft.bankTransfer.accountName" type="text" />
                </label>
                <label class="field">
                  <span>账户尾号</span>
                  <input v-model="paymentDraft.bankTransfer.accountSuffix" type="text" />
                </label>
                <label class="toggle-field">
                  <input v-model="paymentDraft.reconciliation.autoRetryEnabled" type="checkbox" />
                  <span>开启回调自动补偿</span>
                </label>
                <label class="field">
                  <span>补偿间隔（分钟）</span>
                  <input v-model.number="paymentDraft.reconciliation.retryMinutes" type="number" min="1" />
                </label>
                <label class="field">
                  <span>异常通知</span>
                  <input v-model="paymentDraft.reconciliation.abnormalNotify" type="text" />
                </label>
              </div>
            </div>

            <div class="detail-actions">
              <button type="button" class="primary-btn" :disabled="!can('payment.config.manage') || savePaymentPending" @click="handleSavePaymentConfig">
                {{ savePaymentPending ? "保存中..." : "保存支付配置" }}
              </button>
              <span class="state-text">最近校验：{{ paymentDraft.wechat.lastVerifiedAt }}</span>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'payment-records'" class="page-grid">
          <article class="panel">
            <div class="toolbar">
              <label class="field compact">
                <span>门店</span>
                <select v-model="recordFilters.storeId">
                  <option value="all">全部</option>
                  <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                </select>
              </label>
              <label class="field compact">
                <span>支付方式</span>
                <select v-model="recordFilters.method">
                  <option value="all">全部</option>
                  <option value="wechat">微信支付</option>
                  <option value="cash">现金记账</option>
                  <option value="bank_transfer">银行转账</option>
                </select>
              </label>
              <label class="field compact">
                <span>交易状态</span>
                <select v-model="recordFilters.status">
                  <option value="all">全部</option>
                  <option value="paid">已支付</option>
                  <option value="refunding">退款中</option>
                  <option value="failed">失败</option>
                </select>
              </label>
              <label class="field compact">
                <span>回调状态</span>
                <select v-model="recordFilters.callback">
                  <option value="all">全部</option>
                  <option value="delivered">已回调</option>
                  <option value="pending">待补偿</option>
                  <option value="exception">异常</option>
                </select>
              </label>
              <label class="field compact">
                <span>时间</span>
                <select v-model="recordFilters.range">
                  <option value="today">近 1 天</option>
                  <option value="7d">近 7 天</option>
                  <option value="30d">近 30 天</option>
                  <option value="all">全部</option>
                </select>
              </label>
            </div>

            <div class="summary-row">
              <div class="summary-pill">流水金额 {{ formatCurrency(paymentSummary.total) }}</div>
              <div class="summary-pill">异常 {{ paymentSummary.abnormal }} 笔</div>
              <div class="summary-pill">待补偿 {{ paymentSummary.pending }} 笔</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>流水号</th>
                      <th>门店 / 订单</th>
                      <th>方式</th>
                      <th>金额</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="record in filteredRecords"
                      :key="record.id"
                      :class="{ selected: selectedRecord?.id === record.id }"
                      @click="selectedRecordId = record.id"
                    >
                      <td>
                        <strong>{{ record.paymentNo }}</strong>
                        <span>{{ formatDateLabel(record.paidAt) }}</span>
                      </td>
                      <td>{{ record.storeName }} · {{ record.orderNo }}</td>
                      <td>{{ paymentMethodLabel(record.method) }}</td>
                      <td>{{ formatCurrency(record.amount) }}</td>
                      <td>
                        <div class="badge-stack">
                          <span class="badge" :class="record.status === 'paid' ? 'emerald' : record.status === 'refunding' ? 'gold' : 'danger'">
                            {{ paymentStatusLabel(record.status) }}
                          </span>
                          <span class="badge" :class="record.callbackStatus === 'delivered' ? 'slate' : 'gold'">
                            {{ callbackStatusLabel(record.callbackStatus) }}
                          </span>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <aside v-if="selectedRecord" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Payment Detail</p>
                    <h3>{{ selectedRecord.paymentNo }}</h3>
                  </div>
                  <span class="badge" :class="selectedRecord.anomaly ? 'danger' : 'emerald'">
                    {{ selectedRecord.anomaly ? "需复核" : "正常" }}
                  </span>
                </div>

                <div class="detail-grid">
                  <p><strong>订单号：</strong>{{ selectedRecord.orderNo }}</p>
                  <p><strong>业务类型：</strong>{{ selectedRecord.bizType === "recycle" ? "回收单" : "收银订单" }}</p>
                  <p><strong>支付方式：</strong>{{ paymentMethodLabel(selectedRecord.method) }}</p>
                  <p><strong>交易状态：</strong>{{ paymentStatusLabel(selectedRecord.status) }}</p>
                  <p><strong>回调状态：</strong>{{ callbackStatusLabel(selectedRecord.callbackStatus) }}</p>
                  <p><strong>操作人：</strong>{{ selectedRecord.operatorName }}</p>
                  <p><strong>顾客标识：</strong>{{ selectedRecord.customerLabel }}</p>
                  <p><strong>备注：</strong>{{ selectedRecord.remark }}</p>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'system-config' && systemDraft" class="page-grid">
          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">System Profile</p>
                <h3>系统配置总览</h3>
              </div>
              <span class="badge gold">客户要求优先 / ERP 参考补齐</span>
            </div>

            <div class="content-grid two-up">
              <div class="form-card">
                <label class="field">
                  <span>品牌名称</span>
                  <input v-model="systemDraft.brandName" type="text" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>客服电话</span>
                  <input v-model="systemDraft.servicePhone" type="text" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>小票抬头</span>
                  <input v-model="systemDraft.receiptTitle" type="text" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>最少照片数</span>
                  <input v-model.number="systemDraft.minPhotoCount" type="number" min="1" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>最多照片数</span>
                  <input v-model.number="systemDraft.maxPhotoCount" type="number" min="1" :disabled="!can('system.config.manage')" />
                </label>
                <label class="toggle-field">
                  <input v-model="systemDraft.requireExactThree" type="checkbox" :disabled="!can('system.config.manage')" />
                  <span>强制必须 3 张照片</span>
                </label>
                <label class="toggle-field">
                  <input v-model="systemDraft.requireIdCheck" type="checkbox" :disabled="!can('system.config.manage')" />
                  <span>要求身份证核验</span>
                </label>
              </div>

              <div class="form-card">
                <label class="toggle-field">
                  <input v-model="systemDraft.wechatPayEnabled" type="checkbox" :disabled="!can('system.config.manage')" />
                  <span>开启微信支付</span>
                </label>
                <label class="toggle-field">
                  <input v-model="systemDraft.cashEnabled" type="checkbox" :disabled="!can('system.config.manage')" />
                  <span>开启现金记账</span>
                </label>
                <label class="toggle-field">
                  <input v-model="systemDraft.bankTransferEnabled" type="checkbox" :disabled="!can('system.config.manage')" />
                  <span>开启银行转账</span>
                </label>
                <label class="field">
                  <span>域名</span>
                  <input v-model="systemDraft.domainName" type="text" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>域名状态</span>
                  <textarea v-model="systemDraft.domainStatus" rows="2" :disabled="!can('system.config.manage')"></textarea>
                </label>
                <label class="field">
                  <span>OSS 状态</span>
                  <textarea v-model="systemDraft.ossStatus" rows="2" :disabled="!can('system.config.manage')"></textarea>
                </label>
                <label class="field">
                  <span>AppID 状态</span>
                  <input v-model="systemDraft.appIdStatus" type="text" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>商户状态</span>
                  <input v-model="systemDraft.merchantStatus" type="text" :disabled="!can('system.config.manage')" />
                </label>
                <label class="field">
                  <span>打印设备与模板说明</span>
                  <textarea v-model="systemDraft.printerStatus" rows="3" :disabled="!can('system.config.manage')"></textarea>
                </label>
              </div>
            </div>

            <div class="detail-actions">
              <button type="button" class="primary-btn" :disabled="!can('system.config.manage') || saveSystemPending" @click="handleSaveSystem">
                {{ saveSystemPending ? "保存中..." : "保存系统配置" }}
              </button>
              <span class="state-text">这里同时覆盖拍照规则、云资源状态和打印准备信息。</span>
            </div>

            <div v-if="printDraft" class="content-grid two-up">
              <article class="panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Receipt Template</p>
                    <h3>小票模板</h3>
                  </div>
                  <span class="badge slate">P0</span>
                </div>

                <div class="form-card">
                  <label class="toggle-field">
                    <input v-model="printDraft.receipt.enabled" type="checkbox" :disabled="!can('system.config.manage')" />
                    <span>开启小票模板</span>
                  </label>
                  <label class="field">
                    <span>纸宽</span>
                    <input v-model="printDraft.receipt.paperWidth" type="text" :disabled="!can('system.config.manage')" />
                  </label>
                  <label class="field">
                    <span>小票抬头</span>
                    <input v-model="printDraft.receipt.headerTitle" type="text" :disabled="!can('system.config.manage')" />
                  </label>
                  <label class="field">
                    <span>页脚说明</span>
                    <textarea v-model="printDraft.receipt.footerNote" rows="2" :disabled="!can('system.config.manage')"></textarea>
                  </label>
                  <label class="toggle-field">
                    <input v-model="printDraft.receipt.showPhotoSummary" type="checkbox" :disabled="!can('system.config.manage')" />
                    <span>打印照片摘要</span>
                  </label>
                  <label class="toggle-field">
                    <input v-model="printDraft.receipt.showPhotoThumbnails" type="checkbox" :disabled="!can('system.config.manage')" />
                    <span>预留照片缩略图</span>
                  </label>
                </div>
              </article>

              <article class="panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Label & Recycle</p>
                    <h3>标签与回收留痕</h3>
                  </div>
                  <span class="badge gold">P0 / P1</span>
                </div>

                <div class="form-card">
                  <label class="toggle-field">
                    <input v-model="printDraft.label.enabled" type="checkbox" :disabled="!can('system.config.manage')" />
                    <span>开启标签模板</span>
                  </label>
                  <label class="field">
                    <span>标签尺寸</span>
                    <input v-model="printDraft.label.size" type="text" :disabled="!can('system.config.manage')" />
                  </label>
                  <label class="field">
                    <span>打印份数</span>
                    <input v-model.number="printDraft.label.copies" type="number" min="1" :disabled="!can('system.config.manage')" />
                  </label>
                  <label class="field">
                    <span>条码类型</span>
                    <input v-model="printDraft.label.barcodeType" type="text" :disabled="!can('system.config.manage')" />
                  </label>
                  <label class="toggle-field">
                    <input v-model="printDraft.recycle.enabled" type="checkbox" :disabled="!can('system.config.manage')" />
                    <span>开启回收单打印留痕</span>
                  </label>
                  <label class="field">
                    <span>照片摘要文案</span>
                    <textarea v-model="printDraft.recycle.summaryText" rows="2" :disabled="!can('system.config.manage')"></textarea>
                  </label>
                </div>

                <div class="detail-actions">
                  <button type="button" class="primary-btn" :disabled="!can('system.config.manage') || savePrintPending" @click="handleSavePrint">
                    {{ savePrintPending ? "保存中..." : "保存打印模板" }}
                  </button>
                </div>
              </article>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'audit'" class="page-grid">
          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">Audit Trail</p>
                <h3>操作日志</h3>
              </div>
              <span class="badge slate">筛选 / 高亮 / 详情</span>
            </div>

            <div class="toolbar">
              <label class="field grow">
                <span>搜索日志</span>
                <input v-model="auditFilters.keyword" type="text" placeholder="模块 / 动作 / 操作人 / 摘要" />
              </label>
              <label class="field compact">
                <span>模块</span>
                <select v-model="auditFilters.module">
                  <option value="all">全部</option>
                  <option v-for="module in auditModuleOptions.filter((item) => item !== 'all')" :key="module" :value="module">
                    {{ module }}
                  </option>
                </select>
              </label>
              <label class="field compact">
                <span>风险</span>
                <select v-model="auditFilters.risk">
                  <option value="all">全部</option>
                  <option value="high">高风险</option>
                  <option value="medium">中风险</option>
                  <option value="low">低风险</option>
                </select>
              </label>
              <label class="field compact">
                <span>结果</span>
                <select v-model="auditFilters.result">
                  <option value="all">全部</option>
                  <option value="success">成功</option>
                  <option value="warning">提醒</option>
                  <option value="info">记录</option>
                </select>
              </label>
            </div>

            <div class="summary-row">
              <div class="summary-pill">日志 {{ auditSummary.total }} 条</div>
              <div class="summary-pill">高风险 {{ auditSummary.high }} 条</div>
              <div class="summary-pill">提醒 {{ auditSummary.warnings }} 条</div>
              <div class="summary-pill">覆盖模块 {{ auditSummary.modules }} 个</div>
            </div>

            <div class="content-grid sidebar-layout">
              <div class="table-card">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>时间</th>
                      <th>模块</th>
                      <th>动作</th>
                      <th>操作人</th>
                      <th>风险</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr
                      v-for="log in filteredAuditLogs"
                      :key="log.id"
                      :class="{
                        selected: selectedAuditLog?.id === log.id,
                        'table-row-danger': log.riskLevel === 'high',
                        'table-row-warning': log.result === 'warning',
                      }"
                      @click="selectedAuditId = log.id"
                    >
                      <td>{{ formatDateLabel(log.createdAt) }}</td>
                      <td>{{ log.module }}</td>
                      <td>{{ log.action }}</td>
                      <td>{{ log.operatorName }}</td>
                      <td>
                        <span class="badge" :class="log.riskLevel === 'high' ? 'danger' : log.riskLevel === 'medium' ? 'gold' : 'slate'">
                          {{ auditToneLabel(log.riskLevel) }}
                        </span>
                      </td>
                    </tr>
                  </tbody>
                </table>
                <div v-if="!filteredAuditLogs.length" class="empty-block">
                  当前筛选下暂无日志，建议切换模块或放宽关键字。
                </div>
              </div>

              <aside v-if="selectedAuditLog" class="detail-card panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Audit Detail</p>
                    <h3>{{ selectedAuditLog.action }}</h3>
                  </div>
                  <span class="badge" :class="selectedAuditLog.result === 'success' ? 'emerald' : selectedAuditLog.result === 'warning' ? 'gold' : 'slate'">
                    {{ auditResultLabel(selectedAuditLog.result) }}
                  </span>
                </div>

                <div class="detail-grid">
                  <p><strong>时间：</strong>{{ formatDateLabel(selectedAuditLog.createdAt) }}</p>
                  <p><strong>模块：</strong>{{ selectedAuditLog.module }}</p>
                  <p><strong>操作人：</strong>{{ selectedAuditLog.operatorName }}</p>
                  <p><strong>风险级别：</strong>{{ auditToneLabel(selectedAuditLog.riskLevel) }}</p>
                  <p><strong>结果：</strong>{{ auditResultLabel(selectedAuditLog.result) }}</p>
                  <p><strong>摘要：</strong>{{ selectedAuditLog.summary }}</p>
                </div>

                <div class="list-card">
                  <strong>处置建议</strong>
                  <p v-if="selectedAuditLog.riskLevel === 'high'">建议优先复核配置影响面，并同步支付、门店或权限负责人。</p>
                  <p v-else-if="selectedAuditLog.result === 'warning'">建议在真实联调前补齐回调、审计写入点或操作说明。</p>
                  <p v-else>当前为留痕记录，可保留在操作审计中用于后续验收追踪。</p>
                </div>
              </aside>
            </div>
          </article>
        </section>

        <section v-else-if="activePage === 'template-init' && bootstrap" class="page-grid">
          <article class="panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">Template Init</p>
                <h3>{{ bootstrap.templateInit.title }}</h3>
              </div>
              <span class="badge gold">模板化开通工作台</span>
            </div>

            <p class="header-copy">{{ bootstrap.templateInit.description }}</p>

            <div class="summary-row">
              <div class="summary-pill">已完成 {{ templateProgress.done }} / {{ templateProgress.total }}</div>
              <div class="summary-pill">完成度 {{ templateProgress.percent }}%</div>
              <div class="summary-pill">
                当前步骤 {{ templateProgress.current ? templateProgress.current.title : "全部完成" }}
              </div>
            </div>

            <div class="content-grid two-up">
              <article class="panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Launch Form</p>
                    <h3>客户初始化参数</h3>
                  </div>
                </div>

                <div class="form-card">
                  <label class="field">
                    <span>客户包名称</span>
                    <input v-model="templateInitForm.tenantName" type="text" />
                  </label>
                  <label class="field">
                    <span>负责人</span>
                    <input v-model="templateInitForm.ownerName" type="text" />
                  </label>
                  <label class="field">
                    <span>首店城市</span>
                    <input v-model="templateInitForm.city" type="text" />
                  </label>
                  <label class="field">
                    <span>初始门店数</span>
                    <input v-model.number="templateInitForm.initialStoreCount" type="number" min="1" />
                  </label>
                  <label class="toggle-field">
                    <input v-model="templateInitForm.enablePayment" type="checkbox" />
                    <span>同步开通支付中心配置</span>
                  </label>
                  <label class="toggle-field">
                    <input v-model="templateInitForm.enablePrint" type="checkbox" />
                    <span>同步下发打印模板</span>
                  </label>
                  <label class="toggle-field">
                    <input v-model="templateInitForm.requirePhotoAudit" type="checkbox" />
                    <span>默认启用回收拍照规则</span>
                  </label>
                </div>

                <div class="detail-actions">
                  <button type="button" class="primary-btn" :disabled="!templateReady" @click="advanceTemplateInit">
                    推进一步初始化
                  </button>
                  <span class="state-text">当前为前端工作台模拟推进，真实联调后可接开通接口。</span>
                </div>
              </article>

              <article class="panel">
                <div class="panel-head">
                  <div>
                    <p class="eyebrow">Playbook</p>
                    <h3>步骤与交付清单</h3>
                  </div>
                </div>

                <div class="template-grid">
                  <article v-for="step in bootstrap.templateInit.steps" :key="step.id" class="list-card">
                    <div class="list-card-head">
                      <strong>{{ step.title }}</strong>
                      <span class="badge" :class="step.status === 'done' ? 'emerald' : step.status === 'current' ? 'gold' : 'slate'">
                        {{ step.status === "done" ? "已完成" : step.status === "current" ? "当前步骤" : "待规划" }}
                      </span>
                    </div>
                    <p>{{ step.description }}</p>
                  </article>
                </div>

                <div class="chip-row">
                  <span v-for="output in templateOutputPreview" :key="output" class="chip">{{ output }}</span>
                </div>
              </article>
            </div>
          </article>
        </section>

        <section v-else class="page-grid">
          <article class="panel placeholder-panel">
            <p class="eyebrow">Phase 2 Placeholder</p>
            <h3>{{ currentPage.title }}</h3>
            <p>{{ currentPage.description }}</p>

            <div class="template-grid">
              <article v-for="item in PHASE_TWO_ITEMS[activePage]" :key="item" class="list-card">
                <p>{{ item }}</p>
              </article>
            </div>
          </article>
        </section>
      </main>
    </div>
  </div>
</template>
