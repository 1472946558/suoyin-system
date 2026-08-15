<!--
 Copyright (c) 2026 北京纵横时空科技有限责任公司

 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。

 文件名: App.vue
 功能描述: 应用主模块
 作者: 廖心慈
 创建日期: 2026-06-07
-->

<script setup lang="ts">
import { computed, onMounted, reactive, ref, watchEffect } from "vue";
import {
  appointmentAction,
  cancelAppointment,
  cancelRecycleOrder,
  createCustomerHomeBanner,
  createInventoryLedgerItem,
  createMaterialLedgerItem,
  createMemberProfile,
  createProductRecord,
  createStoreRecord,
  createUserAccount,
  deleteCustomerHomeBanner,
  deleteCustomerStyle,
  deleteMaterialLedgerItem,
  deleteProductRecords,
  disableMemberProfile,
  disableProductRecord,
  disableStoreRecord,
  disableUserAccount,
  downloadProductImportTemplate,
  fetchAdminConsole,
  fetchAppointmentDetail,
  fetchAppointmentRules,
  fetchAppointments,
  fetchCashierOrderDetail,
  fetchCashierOrders,
  fetchCustomerHomeConfig,
  fetchCustomerRecycleInfo,
  fetchCustomerStyleCategories,
  fetchCustomerStyles,
  fetchInventoryLedger,
  fetchMaterialLedger,
  fetchMemberProfiles,
  fetchOrderRows,
  fetchProductRecords,
  fetchRecycleOrderDetail,
  fetchRecycleOrders,
  fetchRoleTemplates,
  fetchStoreCustomerConfig,
  fetchStoreRecords,
  fetchUserAccounts,
  importProducts,
  loginAdmin,
  outboundMaterialLedgerItem,
  saveAppointmentRules,
  saveCustomerHomeBanner,
  saveCustomerHomeConfig,
  saveCustomerRecycleInfo,
  saveCustomerStyle,
  saveMemberProfile,
  saveProductRecord,
  saveRoleTemplate,
  saveStoreCustomerConfig,
  saveStoreRecord,
  saveSystemProfile,
  saveUserAccount,
  updateAppointmentStaffNote,
  uploadCustomerImage,
  voidCashierOrder,
  type AbilityCode,
  type AbilityGroup,
  type AdminOrderRow,
  type AppointmentListQuery,
  type AppointmentRecord,
  type AppointmentRules,
  type AttachmentAsset,
  type CashierOrderView,
  type ConsoleBootstrap,
  type CustomerHomeBanner,
  type CustomerHomeBannerInput,
  type CustomerHomeConfig,
  type CustomerRecycleInfoRecord,
  type RecycleProcessStepRecord,
  type RecycleServiceItemRecord,
  type CustomerStyleListQuery,
  type CustomerStyleRecord,
  type DataSource,
  type ListQuery,
  type InventoryLedgerItem as ApiInventoryLedgerItem,
  type MemberProfile,
  type MaterialLedgerItem as ApiMaterialLedgerItem,
  type ProductImportResult,
  type ProductRecord,
  type RecycleOrderView,
  type RoleKey,
  type RoleTemplate,
  type StoreRecord,
  type SystemProfile,
  type UploadScene,
  type UserAccount,
} from "./api";

type NoticeTone = "success" | "warning" | "info";
type PageId =
  | "dashboard"
  | "stores"
  | "users"
  | "roles"
  | "cashier"
  | "recycle"
  | "orders"
  | "members"
  | "products"
  | "inventory"
  | "materials"
  | "settings"
  | "appointments"
  | "customer-home"
  | "customer-styles";
type MetricTone = "gold" | "emerald" | "slate" | "danger";

interface OrderRow {
  id: string;
  type: "cashier" | "recycle";
  orderNo: string;
  customer: string;
  storeName: string;
  amount: number;
  statusLabel: string;
  createdAt: string;
  meta: string;
}

type InventoryStatus = "normal" | "low" | "out" | "review";
type MaterialType = "recycle" | "leftover" | "pledge";

interface InventoryStockItem {
  id: string;
  storeId: string;
  storeName: string;
  styleNo: string;
  name: string;
  category: string;
  purity: string;
  pieceCount: number;
  weightGram: number;
  unitCost: number;
  status: InventoryStatus;
  source: "商品目录" | "手动入库" | "文件导入";
  createdAt: string;
}

interface MaterialLedgerItem {
  id: string;
  storeId: string;
  storeName: string;
  type: MaterialType;
  orderNo: string;
  customerName: string;
  category: string;
  purity: string;
  weightGram: number;
  amount: number;
  remainingWeightGram: number;
  status: string;
  dueDate: string;
  remark: string;
  createdAt: string;
}

const SESSION_STORAGE_KEY = "gold_recycle_admin_session";
const INVENTORY_STORAGE_KEY = "gold_recycle_admin_inventory_entries";
const MATERIAL_STORAGE_KEY = "gold_recycle_admin_material_entries";
const INVENTORY_TEMPLATE_COLUMNS = ["款式编号", "商品名称", "品类", "成色", "件数", "重量", "门店", "成本"];

const PAGE_META: Record<PageId, { title: string; description: string }> = {
  dashboard: {
    title: "经营首页",
    description: "查看门店今日收银、回收、订单、会员和商品概况。",
  },
  stores: {
    title: "门店管理",
    description: "维护门店名称、负责人、营业资料、状态和可见范围。",
  },
  users: {
    title: "账号管理",
    description: "维护老板和店长账号、登录手机号、密码和所属门店。",
  },
  roles: {
    title: "身份范围",
    description: "查看老板和店长两个固定身份的数据范围。",
  },
  cashier: {
    title: "收银记录",
    description: "查看收银单明细、客户、金额、商品和操作员。",
  },
  recycle: {
    title: "回收录单",
    description: "查看回收单、照片留档、金额确认和客户信息。",
  },
  orders: {
    title: "业务订单",
    description: "统一查看收银单和回收单，方便日常对账。",
  },
  members: {
    title: "会员档案",
    description: "维护会员资料、偏好、备注和门店归属。",
  },
  products: {
    title: "商品目录",
    description: "维护商品名称、分类、价格、库存、状态和适用门店。",
  },
  inventory: {
    title: "库存系统",
    description: "商品入库、款式编号、品类、成色、件数、重量和库存查询。",
  },
  materials: {
    title: "旧料管理",
    description: "回收旧料统计、成色重量金额、余料、抵押寄存和月度汇总。",
  },
  settings: {
    title: "系统设置",
    description: "维护品牌名称、客服电话、小票标题和回收拍照规则。",
  },
  appointments: {
    title: "预约管理",
    description: "查看顾客预约记录、确认到店、取消预约和配置预约规则。",
  },
  "customer-home": {
    title: "顾客端配置",
    description: "配置顾客端小程序首页 Banner、品牌文案和服务介绍。",
  },
  "customer-styles": {
    title: "款式工费管理",
    description: "维护顾客端展示的款式、图片、工费说明和上下架状态。",
  },
};

const NAV_ITEMS: Array<{ id: PageId; title: string; hint: string; ability?: AbilityCode }> = [
  { id: "dashboard", title: "首页", hint: "概况", ability: "dashboard.view" },
  { id: "stores", title: "门店", hint: "资料", ability: "store.manage" },
  { id: "users", title: "账号", hint: "人员", ability: "user.manage" },
  { id: "cashier", title: "收银", hint: "记录", ability: "order.view" },
  { id: "recycle", title: "回收", hint: "留档", ability: "recycle.view" },
  { id: "orders", title: "订单", hint: "对账", ability: "order.view" },
  { id: "members", title: "会员", hint: "客户", ability: "user.manage" },
  { id: "products", title: "商品", hint: "维护", ability: "product.manage" },
  { id: "inventory", title: "库存", hint: "入库", ability: "product.manage" },
  { id: "materials", title: "旧料", hint: "台账", ability: "recycle.view" },
  { id: "appointments", title: "预约", hint: "管理", ability: "appointment.manage" },
  { id: "customer-home", title: "顾客端", hint: "配置", ability: "customer_content.view" },
  { id: "customer-styles", title: "款式", hint: "工费", ability: "product.view" },
  { id: "settings", title: "设置", hint: "系统", ability: "system.config.manage" },
];

const QUICK_ACTIONS: Array<{ id: PageId; title: string; description: string; ability?: AbilityCode }> = [
  { id: "stores", title: "门店资料", description: "维护营业门店信息和负责人", ability: "store.manage" },
  { id: "users", title: "账号权限", description: "维护老板和店长账号", ability: "user.manage" },
  { id: "cashier", title: "查看收银", description: "核对成交金额与商品件数", ability: "order.view" },
  { id: "recycle", title: "查看回收", description: "核对客户、照片和确认金额", ability: "recycle.view" },
  { id: "members", title: "会员档案", description: "维护常客资料和回访备注", ability: "user.manage" },
  { id: "products", title: "维护商品", description: "新增商品、修改价格和状态", ability: "product.manage" },
  { id: "inventory", title: "库存系统", description: "入库、导入和查询库存", ability: "product.manage" },
  { id: "materials", title: "旧料台账", description: "统计旧料和抵押寄存", ability: "recycle.view" },
  { id: "settings", title: "基础设置", description: "维护客服电话、小票和拍照规则", ability: "system.config.manage" },
];

const loginForm = reactive({
  phone: "",
  password: "",
  roleKey: "boss" as RoleKey,
});

const bootstrap = ref<ConsoleBootstrap | null>(null);
const sourceMode = ref<DataSource>("api");
const loading = ref(false);
const loginPending = ref(false);
const actionPending = ref(false);
const sessionToken = ref("");
const activePage = ref<PageId>("dashboard");
const loginError = ref("");
const notice = ref<{ tone: NoticeTone; text: string } | null>(null);

const selectedStoreId = ref("");
const selectedUserId = ref("");
const selectedRoleId = ref<number | null>(null);
const selectedCashierId = ref("");
const selectedRecycleId = ref("");
const selectedMemberId = ref("");
const selectedProductId = ref("");

const storeItems = ref<StoreRecord[]>([]);
const userItems = ref<UserAccount[]>([]);
const roleItems = ref<RoleTemplate[]>([]);
const memberItems = ref<MemberProfile[]>([]);
const productItems = ref<ProductRecord[]>([]);
const cashierItems = ref<CashierOrderView[]>([]);
const recycleItems = ref<RecycleOrderView[]>([]);
const orderItems = ref<AdminOrderRow[]>([]);

const selectedCashierDetail = ref<CashierOrderView | null>(null);
const selectedRecycleDetail = ref<RecycleOrderView | null>(null);

// --- 预约管理状态 ---
const appointmentItems = ref<AppointmentRecord[]>([]);
const appointmentFilters = reactive<AppointmentListQuery>({
  status: "",
  storeId: "",
  phone: "",
  serviceType: "",
  dateFrom: "",
  dateTo: "",
  page: 1,
  pageSize: 50,
});
const selectedAppointmentId = ref("");
const selectedAppointmentDetail = ref<AppointmentRecord | null>(null);
const staffNoteDraft = ref("");
const cancelReasonDraft = ref("");
const cancelDialogOpen = ref(false);
const appointmentRules = ref<AppointmentRules | null>(null);
const appointmentRulesDraft = ref<AppointmentRules | null>(null);
const rulesEditing = ref(false);

// --- 顾客端内容管理 ---
const customerHomeConfig = ref<CustomerHomeConfig | null>(null);
const homeConfigDraft = ref<Omit<CustomerHomeConfig, "banners"> | null>(null);
const homeBanners = ref<CustomerHomeBanner[]>([]);
const bannerDialogOpen = ref(false);
const bannerEditingId = ref("");
const bannerDraft = ref<CustomerHomeBannerInput | null>(null);
const recycleInfoDraft = ref<CustomerRecycleInfoRecord | null>(null);
const styleItems = ref<CustomerStyleRecord[]>([]);
const styleTotal = ref(0);
const styleCategories = ref<string[]>([]);
const styleFilters = reactive<CustomerStyleListQuery>({ category: "", status: "", keyword: "", page: 1, pageSize: 20 });
const styleDialogOpen = ref(false);
const styleEditingId = ref("");
const styleDraft = ref<Partial<CustomerStyleRecord> | null>(null);
const uploadPending = ref(false);

const storeFilters = reactive<ListQuery>({ page: 1, pageSize: 50, status: "", keyword: "" });
const userFilters = reactive<ListQuery>({ page: 1, pageSize: 50, storeId: "", status: "", keyword: "" });
const cashierFilters = reactive<ListQuery>({ page: 1, pageSize: 50, storeId: "", status: "", dateFrom: "", dateTo: "", keyword: "" });
const recycleFilters = reactive<ListQuery>({ page: 1, pageSize: 50, storeId: "", status: "", dateFrom: "", dateTo: "", keyword: "" });
const orderFilters = reactive<ListQuery>({ page: 1, pageSize: 50, storeId: "", status: "", dateFrom: "", dateTo: "", keyword: "" });
const memberFilters = reactive<ListQuery>({ page: 1, pageSize: 50, storeId: "", status: "", keyword: "" });
const productFilters = reactive<ListQuery>({ page: 1, pageSize: 50, storeId: "", status: "", keyword: "" });

const recyclePreview = reactive({
  open: false,
  orderNo: "",
  assets: [] as AttachmentAsset[],
  index: 0,
});

const storeDraft = ref<StoreRecord | null>(null);
const storeCustomerDraft = ref<{
  imageUrl: string;
  appointmentEnabled: boolean;
  serviceTagsText: string;
  sortOrder: number;
} | null>(null);
const userDraft = ref<UserAccount | null>(null);
const roleDraft = ref<RoleTemplate | null>(null);
const memberDraft = ref<MemberProfile | null>(null);
const productDraft = ref<ProductRecord | null>(null);
const systemDraft = ref<SystemProfile | null>(null);
const productStoreSearch = ref("");
const productStorePickerOpen = ref(false);
const productTagsText = ref("");
const selectedProductIds = ref<string[]>([]);
const selectedMaterialIds = ref<string[]>([]);
const productImportStoreId = ref("");
const productImportFile = ref<File | null>(null);
const productImportResult = ref<ProductImportResult | null>(null);
const productImportInputKey = ref(0);
const inventoryEntries = ref<InventoryStockItem[]>([]);
const materialEntries = ref<MaterialLedgerItem[]>([]);
const inventoryImportInputKey = ref(0);
const inventoryFilters = reactive({
  storeId: "",
  keyword: "",
  category: "",
  purity: "",
});
const materialFilters = reactive({
  storeId: "",
  keyword: "",
  type: "",
  dateFrom: "",
  dateTo: "",
});
const inventoryDraft = reactive({
  storeId: "",
  styleNo: "",
  name: "",
  category: "",
  purity: "",
  pieceCount: 1,
  weightGram: 0,
  unitCost: 0,
  status: "normal" as InventoryStatus,
});
const materialDraft = reactive({
  storeId: "",
  type: "pledge" as MaterialType,
  customerName: "",
  category: "黄金",
  purity: "足金9999",
  weightGram: 0,
  amount: 0,
  remainingWeightGram: 0,
  status: "在库",
  dueDate: "",
  remark: "",
});

const pageMeta = computed(() => PAGE_META[activePage.value]);
const currentUser = computed(() => bootstrap.value?.currentUser ?? null);
const stores = computed(() => storeItems.value);
const users = computed(() => userItems.value);
const roles = computed(() => roleItems.value);
const products = computed(() => productItems.value);
const members = computed(() => memberItems.value);
const cashierOrders = computed(() => cashierItems.value);
const recycleOrders = computed(() => recycleItems.value);
const abilityGroups = computed<AbilityGroup[]>(() => bootstrap.value?.abilityGroups ?? []);
const systemProfile = computed(() => bootstrap.value?.systemProfile ?? null);
const productImportLogs = computed(() => bootstrap.value?.importLogs ?? []);

const visibleStoreNames = computed(() => {
  const names = stores.value.map((store) => store.name).filter(Boolean);
  return names.length ? names.join("、") : "暂无门店";
});

const cashTotal = computed(() => cashierOrders.value.reduce((sum, order) => sum + order.totalAmount, 0));
const recycleTotal = computed(() =>
  recycleOrders.value.reduce((sum, order) => sum + (order.confirmedAmount || order.estimatedAmount), 0),
);
const memberTotalAmount = computed(() =>
  members.value.reduce((sum, member) => sum + Number(member.totalRecycleAmount || 0), 0),
);

const dashboardMetrics = computed(() => [
  {
    label: "收银单",
    value: String(cashierOrders.value.length),
    hint: `金额 ${formatCurrency(cashTotal.value)}`,
    tone: "gold" as MetricTone,
  },
  {
    label: "回收单",
    value: String(recycleOrders.value.length),
    hint: `金额 ${formatCurrency(recycleTotal.value)}`,
    tone: "emerald" as MetricTone,
  },
  {
    label: "会员档案",
    value: String(members.value.length),
    hint: `累计回收 ${formatCurrency(memberTotalAmount.value)}`,
    tone: "slate" as MetricTone,
  },
  {
    label: "商品目录",
    value: String(products.value.length),
    hint: `在售 ${activeProducts.value.length} 个`,
    tone: "slate" as MetricTone,
  },
]);

const activeProducts = computed(() => products.value.filter((product) => product.status === "active"));
const selectedStore = computed(() => stores.value.find((item) => item.id === selectedStoreId.value) ?? stores.value[0] ?? null);
const selectedUser = computed(() => users.value.find((item) => item.id === selectedUserId.value) ?? users.value[0] ?? null);
const selectedRole = computed(() => roles.value.find((item) => item.id === selectedRoleId.value) ?? roles.value[0] ?? null);
const selectedCashier = computed(() => cashierOrders.value.find((item) => item.id === selectedCashierId.value) ?? cashierOrders.value[0] ?? null);
const selectedRecycle = computed(() => recycleOrders.value.find((item) => item.id === selectedRecycleId.value) ?? recycleOrders.value[0] ?? null);
const selectedMember = computed(() => members.value.find((item) => item.id === selectedMemberId.value) ?? members.value[0] ?? null);
const selectedProduct = computed(() => products.value.find((item) => item.id === selectedProductId.value) ?? products.value[0] ?? null);
const selectedAppointment = computed(() => appointmentItems.value.find((item) => item.id === selectedAppointmentId.value) ?? appointmentItems.value[0] ?? null);
const selectedProductStores = computed(() => {
  const ids = new Set(productDraft.value?.storeIds || []);
  return stores.value.filter((store) => ids.has(store.id));
});
const productStoreOptions = computed(() => {
  const keyword = productStoreSearch.value.trim().toLowerCase();
  if (!keyword) return stores.value;
  return stores.value.filter((store) =>
    [store.name, store.code, store.city, store.address].some((field) => String(field || "").toLowerCase().includes(keyword)),
  );
});
const currentCashierDetail = computed(() => selectedCashierDetail.value ?? selectedCashier.value ?? null);
const currentRecycleDetail = computed(() => selectedRecycleDetail.value ?? selectedRecycle.value ?? null);
const previewAsset = computed(() => recyclePreview.assets[recyclePreview.index] ?? null);
const visibleNavItems = computed(() => NAV_ITEMS.filter((item) => !item.ability || currentUser.value?.abilities.includes(item.ability)));
const visibleQuickActions = computed(() => QUICK_ACTIONS.filter((item) => !item.ability || currentUser.value?.abilities.includes(item.ability)));
const allVisibleProductsSelected = computed(() =>
  products.value.length > 0 && products.value.every((product) => selectedProductIds.value.includes(product.id)),
);
const deletableMaterialRows = computed(() => filteredMaterialRows.value.filter((item) => !item.id.startsWith("recycle-")));
const allVisibleMaterialsSelected = computed(() =>
  deletableMaterialRows.value.length > 0 && deletableMaterialRows.value.every((item) => selectedMaterialIds.value.includes(item.id)),
);

const productStats = computed(() => ({
  total: products.value.length,
  active: activeProducts.value.length,
  inventory: products.value.reduce((sum, product) => sum + Number(product.inventory || 0), 0),
}));

const productInventoryRows = computed<InventoryStockItem[]>(() =>
  products.value
    .filter((product) => product.status !== "disabled")
    .map((product) => {
      const storeId = getProductStoreIds(product)[0] || stores.value[0]?.id || "";
      const storeName = stores.value.find((store) => store.id === storeId)?.name || product.storeNames[0] || "未分配门店";
      const purity = (product as ProductRecord & { purity?: string }).purity || product.categoryTab || "客户自填";
      const pieceCount = Number(product.inventory || 0);
      return {
        id: `product-${product.id}`,
        storeId,
        storeName,
        styleNo: product.sku,
        name: product.name,
        category: product.category || "未分类",
        purity,
        pieceCount,
        weightGram: Number(product.gramWeight || 0) * Math.max(pieceCount, 1),
        unitCost: Number(product.price || 0),
        status: product.stockStatus === "out" || pieceCount <= 0 ? "out" : product.stockStatus === "low" ? "low" : product.stockStatus === "review" ? "review" : "normal",
        source: "商品目录",
        createdAt: "",
      };
    }),
);

const inventoryRows = computed(() => [...inventoryEntries.value, ...productInventoryRows.value]);

const filteredInventoryRows = computed(() => {
  const keyword = inventoryFilters.keyword.trim().toLowerCase();
  return inventoryRows.value.filter((item) => {
    const matchesStore = !inventoryFilters.storeId || item.storeId === inventoryFilters.storeId;
    const matchesCategory = !inventoryFilters.category || item.category.includes(inventoryFilters.category);
    const matchesPurity = !inventoryFilters.purity || item.purity.includes(inventoryFilters.purity);
    const haystack = [item.styleNo, item.name, item.category, item.purity, item.storeName, item.source].join(" ").toLowerCase();
    return matchesStore && matchesCategory && matchesPurity && (!keyword || haystack.includes(keyword));
  });
});

const inventorySummary = computed(() => ({
  styles: filteredInventoryRows.value.length,
  pieces: filteredInventoryRows.value.reduce((sum, item) => sum + Number(item.pieceCount || 0), 0),
  weight: filteredInventoryRows.value.reduce((sum, item) => sum + Number(item.weightGram || 0), 0),
  amount: filteredInventoryRows.value.reduce((sum, item) => sum + Number(item.unitCost || 0) * Number(item.pieceCount || 0), 0),
}));

const inventoryCategories = computed(() => Array.from(new Set(inventoryRows.value.map((item) => item.category).filter(Boolean))));
const inventoryPurities = computed(() => Array.from(new Set(inventoryRows.value.map((item) => item.purity).filter(Boolean))));

const recycleMaterialRows = computed<MaterialLedgerItem[]>(() =>
  recycleOrders.value.flatMap((order) => {
    const fallbackItems = parseRecycleItemSummary(order.itemSummary);
    const sourceItems = order.items?.length ? order.items : fallbackItems;
    const itemCount = Math.max(sourceItems.length || 1, 1);
    return sourceItems.map((item, index) => ({
      id: `recycle-${order.id}-${index}`,
      storeId: order.storeId,
      storeName: order.storeName,
      type: "recycle" as MaterialType,
      orderNo: order.orderNo,
      customerName: order.customerName || "未填客户",
      category: item.category || "旧料",
      purity: item.purity || "客户自填",
      weightGram: Number(item.weightGram || 0),
      amount: Number((order.confirmedAmount || order.estimatedAmount || 0) / itemCount),
      remainingWeightGram: order.status === "cancelled" ? 0 : Number(item.weightGram || 0),
      status: recycleStatusLabel(order.status),
      dueDate: "",
      remark: order.remark || "",
      createdAt: order.createdAt,
    }));
  }),
);

const materialRows = computed(() => [...materialEntries.value, ...recycleMaterialRows.value]);

const filteredMaterialRows = computed(() => {
  const keyword = materialFilters.keyword.trim().toLowerCase();
  const start = materialFilters.dateFrom ? new Date(`${materialFilters.dateFrom}T00:00:00`).getTime() : 0;
  const end = materialFilters.dateTo ? new Date(`${materialFilters.dateTo}T23:59:59`).getTime() : Number.POSITIVE_INFINITY;
  return materialRows.value.filter((item) => {
    const timestamp = item.createdAt ? new Date(item.createdAt).getTime() : 0;
    const haystack = [item.orderNo, item.customerName, item.category, item.purity, item.storeName, item.remark].join(" ").toLowerCase();
    return (
      (!materialFilters.storeId || item.storeId === materialFilters.storeId) &&
      (!materialFilters.type || item.type === materialFilters.type) &&
      timestamp >= start &&
      timestamp <= end &&
      (!keyword || haystack.includes(keyword))
    );
  });
});

const materialSummary = computed(() => {
  const todayKey = toLocalDateKey();
  const monthKey = todayKey.slice(0, 7);
  const todayItems = materialRows.value.filter((item) => item.createdAt.startsWith(todayKey));
  const monthItems = materialRows.value.filter((item) => item.createdAt.startsWith(monthKey));
  const pledgeItems = materialRows.value.filter((item) => item.type === "pledge");
  return {
    todayWeight: todayItems.reduce((sum, item) => sum + Number(item.weightGram || 0), 0),
    todayAmount: todayItems.reduce((sum, item) => sum + Number(item.amount || 0), 0),
    monthWeight: monthItems.reduce((sum, item) => sum + Number(item.weightGram || 0), 0),
    monthAmount: monthItems.reduce((sum, item) => sum + Number(item.amount || 0), 0),
    remainingWeight: materialRows.value.reduce((sum, item) => sum + Number(item.remainingWeightGram || 0), 0),
    pledgeCount: pledgeItems.length,
    pledgeAmount: pledgeItems.reduce((sum, item) => sum + Number(item.amount || 0), 0),
  };
});

const materialPurityStats = computed(() => {
  const grouped = new Map<string, { purity: string; weight: number; amount: number; count: number }>();
  materialRows.value.forEach((item) => {
    const key = item.purity || "未填成色";
    const current = grouped.get(key) || { purity: key, weight: 0, amount: 0, count: 0 };
    current.weight += Number(item.weightGram || 0);
    current.amount += Number(item.amount || 0);
    current.count += 1;
    grouped.set(key, current);
  });
  return Array.from(grouped.values()).sort((left, right) => right.weight - left.weight);
});

const memberStats = computed(() => ({
  total: members.value.length,
  vip: members.value.filter((member) => member.level === "VIP").length,
  verified: members.value.filter((member) => member.idVerified).length,
}));

const settingsRows = computed(() => [
  { label: "登录账号", value: currentUser.value ? `${currentUser.value.name} · ${currentUser.value.account}` : "-" },
  { label: "登录身份", value: currentUser.value?.roleName || "-" },
  { label: "门店范围", value: visibleStoreNames.value },
  { label: "客服电话", value: systemProfile.value?.servicePhone || "-" },
  { label: "小票标题", value: systemProfile.value?.receiptTitle || "-" },
  {
    label: "照片规则",
    value: systemProfile.value ? `${systemProfile.value.minPhotoCount}-${systemProfile.value.maxPhotoCount} 张` : "-",
  },
  { label: "打印状态", value: systemProfile.value?.printerStatus || "-" },
]);

const orderRows = computed<OrderRow[]>(() => {
  return orderItems.value.map((order) => ({
    id: order.id,
    type: order.type,
    orderNo: order.orderNo,
    customer: formatCustomerName(order.customerName, order.customerPhone),
    storeName: order.storeName,
    amount: order.amount,
    statusLabel: order.type === "cashier" ? orderStatusLabel(order.status as CashierOrderView["status"]) : recycleStatusLabel(order.status as RecycleOrderView["status"]),
    createdAt: order.createdAt,
    meta: `${order.itemSummary || (order.type === "cashier" ? "商品明细" : "回收品类")} · ${order.createdBy}`,
  }));
});

function seedModuleData(data: ConsoleBootstrap) {
  storeItems.value = [...data.stores];
  userItems.value = data.users.map((item) => cloneUser(item)).filter((item): item is UserAccount => Boolean(item));
  roleItems.value = [...data.roles];
  memberItems.value = [...data.members];
  productItems.value = [...data.products];
  cashierItems.value = [...data.cashierOrders];
  recycleItems.value = [...data.recycleOrders];
  orderItems.value = [
    ...data.cashierOrders.map((item) => ({
      id: item.id,
      type: "cashier" as const,
      orderNo: item.orderNo,
      storeId: item.storeId,
      storeName: item.storeName,
      status: item.status,
      customerName: item.customerName,
      customerPhone: item.customerPhone,
      amount: item.totalAmount,
      itemSummary: item.itemSummary,
      createdBy: item.createdBy,
      createdAt: item.createdAt,
    })),
    ...data.recycleOrders.map((item) => ({
      id: item.id,
      type: "recycle" as const,
      orderNo: item.orderNo,
      storeId: item.storeId,
      storeName: item.storeName,
      status: item.status,
      customerName: item.customerName,
      customerPhone: item.customerPhone,
      amount: item.confirmedAmount || item.estimatedAmount,
      itemSummary: item.itemSummary,
      createdBy: item.createdBy,
      createdAt: item.createdAt,
    })),
  ].sort((left, right) => new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime());
}

function cloneStore(store: StoreRecord | null | undefined): StoreRecord | null {
  return store ? JSON.parse(JSON.stringify(store)) as StoreRecord : null;
}

function cloneUser(user: UserAccount | null | undefined): UserAccount | null {
  if (!user) return null;
  const cloned = JSON.parse(JSON.stringify(user)) as UserAccount;
  const roleKey = normalizeRoleKey(cloned.roleKey);
  return {
    ...cloned,
    name: cloned.name || "新账号",
    account: cloned.account || "",
    phone: cloned.phone || "",
    password: cloned.password || "",
    roleKey,
    roleName: roleKey === "boss" ? "老板" : "店长",
    dataScope: roleKey === "boss" ? "all_stores" : "assigned_store",
    storeIds: Array.isArray(cloned.storeIds) ? cloned.storeIds : [],
    storeNames: Array.isArray(cloned.storeNames) ? cloned.storeNames : [],
    status: cloned.status || "invited",
    abilities: Array.isArray(cloned.abilities) ? cloned.abilities : [],
    lastLoginAt: cloned.lastLoginAt || "",
  };
}

function normalizeRoleKey(roleKey: string): RoleKey {
  return roleKey === "boss" || roleKey === "owner" ? "boss" : "shop_manager";
}

function cloneRole(role: RoleTemplate | null | undefined): RoleTemplate | null {
  return role ? JSON.parse(JSON.stringify(role)) as RoleTemplate : null;
}

function cloneMember(member: MemberProfile | null | undefined): MemberProfile | null {
  return member ? JSON.parse(JSON.stringify(member)) as MemberProfile : null;
}

function productStoreIdsFromNames(storeNames: string[] | undefined) {
  const names = new Set((storeNames || []).map((name) => name.trim()).filter(Boolean));
  return stores.value.filter((store) => names.has(store.name)).map((store) => store.id);
}

function getProductStoreIds(product: ProductRecord) {
  const ids = Array.isArray(product.storeIds) ? product.storeIds.filter(Boolean) : [];
  return ids.length ? ids : productStoreIdsFromNames(product.storeNames);
}

function cloneProduct(product: ProductRecord | null | undefined): ProductRecord | null {
  if (!product) return null;
  return {
    ...product,
    categoryTab: product.categoryTab || product.category,
    imageUrl: product.imageUrl || "",
    price: Number(product.price || 0),
    gramWeight: Number(product.gramWeight || 0),
    inventory: Number(product.inventory || 0),
    stockStatus: product.stockStatus || "normal",
    storeIds: getProductStoreIds(product),
    storeNames: [...(product.storeNames || [])],
    tags: [...(product.tags || [])],
  };
}

function cloneSystemProfile(profile: SystemProfile | null | undefined): SystemProfile | null {
  return profile ? JSON.parse(JSON.stringify(profile)) as SystemProfile : null;
}

function syncDrafts(data?: ConsoleBootstrap | null) {
  selectedStoreId.value = stores.value.find((item) => item.id === selectedStoreId.value)?.id || stores.value[0]?.id || "";
  selectedUserId.value = users.value.find((item) => item.id === selectedUserId.value)?.id || users.value[0]?.id || "";
  selectedRoleId.value = roles.value.find((item) => item.id === selectedRoleId.value)?.id || roles.value[0]?.id || null;
  selectedCashierId.value = cashierOrders.value.find((item) => item.id === selectedCashierId.value)?.id || cashierOrders.value[0]?.id || "";
  selectedRecycleId.value = recycleOrders.value.find((item) => item.id === selectedRecycleId.value)?.id || recycleOrders.value[0]?.id || "";
  selectedMemberId.value = members.value.find((item) => item.id === selectedMemberId.value)?.id || members.value[0]?.id || "";
  selectedProductId.value = products.value.find((item) => item.id === selectedProductId.value)?.id || products.value[0]?.id || "";

  storeDraft.value = cloneStore(stores.value.find((item) => item.id === selectedStoreId.value) ?? null);
  if (selectedStoreId.value) {
    void loadStoreCustomerConfig(selectedStoreId.value);
  }
  userDraft.value = cloneUser(users.value.find((item) => item.id === selectedUserId.value) ?? null);
  roleDraft.value = cloneRole(roles.value.find((item) => item.id === selectedRoleId.value) ?? null);
  memberDraft.value = cloneMember(members.value.find((item) => item.id === selectedMemberId.value) ?? null);
  const product = products.value.find((item) => item.id === selectedProductId.value) ?? null;
  productDraft.value = cloneProduct(product);
  productStoreSearch.value = "";
  productStorePickerOpen.value = false;
  productTagsText.value = product?.tags?.join("、") || "";
  productImportStoreId.value = stores.value.find((item) => item.id === productImportStoreId.value)?.id || stores.value[0]?.id || "";
  inventoryDraft.storeId = stores.value.find((item) => item.id === inventoryDraft.storeId)?.id || stores.value[0]?.id || "";
  materialDraft.storeId = stores.value.find((item) => item.id === materialDraft.storeId)?.id || stores.value[0]?.id || "";
  systemDraft.value = cloneSystemProfile(data?.systemProfile ?? systemProfile.value);
  selectedCashierDetail.value = selectedCashier.value ?? null;
  selectedRecycleDetail.value = selectedRecycle.value ?? null;
}

function splitTextList(text: string) {
  return text
    .split(/[、,，\n]/)
    .map((item) => item.trim())
    .filter(Boolean);
}

function parseRecycleItemSummary(summary: string | undefined) {
  const parts = String(summary || "")
    .split(/[·,，、/]/)
    .map((item) => item.trim())
    .filter(Boolean);
  const weightPart = parts.find((item) => /g$/i.test(item) || /克$/.test(item));
  const weightGram = weightPart ? Number(weightPart.replace(/[^\d.]/g, "")) : 0;
  const purity = parts.find((item) => /(足金|K金|黄金|铂|银|钻|999|9999|Au)/i.test(item) && !/g$/i.test(item) && !/克$/.test(item)) || "客户自填";
  const category =
    parts.find((item) => item !== purity && item !== weightPart && /(金饰|黄金|K金|钻石|铂金|银饰|旧料|首饰)/i.test(item)) ||
    parts.find((item) => item !== purity && item !== weightPart) ||
    "旧料";
  return [{ category, purity, weightGram }];
}

function toLocalDateKey(value = new Date()) {
  const year = value.getFullYear();
  const month = String(value.getMonth() + 1).padStart(2, "0");
  const day = String(value.getDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function toLocalDateTime(value = new Date()) {
  const hours = String(value.getHours()).padStart(2, "0");
  const minutes = String(value.getMinutes()).padStart(2, "0");
  const seconds = String(value.getSeconds()).padStart(2, "0");
  return `${toLocalDateKey(value)}T${hours}:${minutes}:${seconds}`;
}

function formatCurrency(amount: number) {
  return `¥${Math.round(amount || 0).toLocaleString("zh-CN")}`;
}

function formatWeight(weight: number) {
  return `${Number(weight || 0).toLocaleString("zh-CN", { maximumFractionDigits: 2 })}g`;
}

function inventoryStatusLabel(status: InventoryStatus | string | undefined) {
  return {
    normal: "库存正常",
    low: "库存较低",
    out: "无库存",
    review: "待盘点",
  }[status || ""] || "未设置";
}

function materialTypeLabel(type: MaterialType | string) {
  return {
    recycle: "回收旧料",
    leftover: "余料",
    pledge: "抵押寄存",
  }[type] || type;
}

function normalizeInventoryEntry(entry: Record<string, any>): InventoryStockItem {
  const storeId = entry.storeId || stores.value[0]?.id || "";
  const store = stores.value.find((item) => item.id === storeId);
  const rawStatus = String(entry.status || "");
  const rawSource = String(entry.source || "");
  return {
    id: entry.id || `inv-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    storeId,
    storeName: store?.name || entry.storeName || "未分配门店",
    styleNo: String(entry.styleNo || "").trim(),
    name: String(entry.name || "").trim(),
    category: String(entry.category || "黄金").trim(),
    purity: String(entry.purity || "客户自填").trim(),
    pieceCount: Number(entry.pieceCount || 0),
    weightGram: Number(entry.weightGram || 0),
    unitCost: Number(entry.unitCost || (entry as Partial<ApiInventoryLedgerItem>).costAmount || 0),
    status: ((rawStatus === "in_stock" ? "normal" : rawStatus) || "normal") as InventoryStatus,
    source: rawSource === "file" || rawSource === "文件导入" ? "文件导入" : (rawSource === "商品目录" ? "商品目录" : "手动入库"),
    createdAt: entry.createdAt || toLocalDateTime(),
  };
}

function normalizeMaterialEntry(entry: Record<string, any>): MaterialLedgerItem {
  const storeId = entry.storeId || stores.value[0]?.id || "";
  const store = stores.value.find((item) => item.id === storeId);
  return {
    id: entry.id || `mat-${Date.now()}-${Math.random().toString(36).slice(2, 8)}`,
    storeId,
    storeName: store?.name || entry.storeName || "未分配门店",
    type: entry.type || "pledge",
    orderNo: entry.orderNo || `JD-${toLocalDateKey().replaceAll("-", "")}-${String(materialEntries.value.length + 1).padStart(3, "0")}`,
    customerName: String(entry.customerName || "未填客户").trim(),
    category: String(entry.category || "黄金").trim(),
    purity: String(entry.purity || "客户自填").trim(),
    weightGram: Number(entry.weightGram || 0),
    amount: Number(entry.amount || 0),
    remainingWeightGram: Number(entry.remainingWeightGram ?? entry.weightGram ?? 0),
    status: String(entry.status === "in_stock" ? "在库" : (entry.status === "outbound" ? "已出库" : (entry.status || "在库"))).trim(),
    dueDate: entry.dueDate || "",
    remark: entry.remark || "",
    createdAt: entry.createdAt || toLocalDateTime(),
  };
}

function loadLocalLedgers() {
  try {
    const inventory = JSON.parse(window.localStorage.getItem(INVENTORY_STORAGE_KEY) || "[]") as Partial<InventoryStockItem>[];
    inventoryEntries.value = Array.isArray(inventory) ? inventory.map((item) => normalizeInventoryEntry(item)).filter((item) => item.styleNo && item.name) : [];
  } catch {
    inventoryEntries.value = [];
  }
  try {
    const materials = JSON.parse(window.localStorage.getItem(MATERIAL_STORAGE_KEY) || "[]") as Partial<MaterialLedgerItem>[];
    materialEntries.value = Array.isArray(materials) ? materials.map((item) => normalizeMaterialEntry(item)) : [];
  } catch {
    materialEntries.value = [];
  }
}

function saveInventoryEntries() {
  window.localStorage.setItem(INVENTORY_STORAGE_KEY, JSON.stringify(inventoryEntries.value));
}

function saveMaterialEntries() {
  window.localStorage.setItem(MATERIAL_STORAGE_KEY, JSON.stringify(materialEntries.value));
}

function resetInventoryDraft() {
  inventoryDraft.storeId = stores.value[0]?.id || "";
  inventoryDraft.styleNo = "";
  inventoryDraft.name = "";
  inventoryDraft.category = "";
  inventoryDraft.purity = "";
  inventoryDraft.pieceCount = 1;
  inventoryDraft.weightGram = 0;
  inventoryDraft.unitCost = 0;
  inventoryDraft.status = "normal";
}

function resetMaterialDraft() {
  materialDraft.storeId = stores.value[0]?.id || "";
  materialDraft.type = "pledge";
  materialDraft.customerName = "";
  materialDraft.category = "黄金";
  materialDraft.purity = "足金9999";
  materialDraft.weightGram = 0;
  materialDraft.amount = 0;
  materialDraft.remainingWeightGram = 0;
  materialDraft.status = "在库";
  materialDraft.dueDate = "";
  materialDraft.remark = "";
}

function addInventoryEntry() {
  if (!sessionToken.value) return;
  if (!inventoryDraft.styleNo.trim() || !inventoryDraft.name.trim()) {
    showNotice("warning", "请填写款式编号和商品名称。");
    return;
  }
  void runAction(async () => {
    await createInventoryLedgerItem(sessionToken.value, {
      storeId: inventoryDraft.storeId,
      styleNo: inventoryDraft.styleNo,
      name: inventoryDraft.name,
      category: inventoryDraft.category,
      purity: inventoryDraft.purity,
      pieceCount: inventoryDraft.pieceCount,
      weightGram: inventoryDraft.weightGram,
      costAmount: inventoryDraft.unitCost,
      status: inventoryDraft.status === "out" ? "outbound" : "in_stock",
      source: "admin",
    });
    resetInventoryDraft();
  }, "商品已入库，已写入后端台账。", "inventory");
}

function addMaterialEntry() {
  if (!sessionToken.value) return;
  if (!materialDraft.category.trim() || !materialDraft.purity.trim()) {
    showNotice("warning", "请填写旧料品类和成色。");
    return;
  }
  void runAction(async () => {
    await createMaterialLedgerItem(sessionToken.value, {
      storeId: materialDraft.storeId,
      type: materialDraft.type,
      customerName: materialDraft.customerName,
      category: materialDraft.category,
      purity: materialDraft.purity,
      weightGram: materialDraft.weightGram,
      amount: materialDraft.amount,
      remainingWeightGram: materialDraft.remainingWeightGram || materialDraft.weightGram,
      status: "in_stock",
      dueDate: materialDraft.dueDate,
      remark: materialDraft.remark,
      source: "admin",
    });
    resetMaterialDraft();
  }, "旧料/寄存记录已登记，已写入后端台账。", "materials");
}

function removeMaterialEntry(id: string) {
  if (!sessionToken.value) return;
  void runAction(async () => {
    await outboundMaterialLedgerItem(sessionToken.value, id, "后台旧料出库");
  }, "旧料已出库，余料统计已更新。", "materials");
}

function deleteMaterialEntry(id: string) {
  if (!sessionToken.value) return;
  void runAction(async () => {
    await deleteMaterialLedgerItem(sessionToken.value, id);
  }, "旧料记录已删除。", "materials");
}

function batchDeleteMaterials() {
  if (!sessionToken.value || !selectedMaterialIds.value.length) return;
  void runAction(async () => {
    await Promise.all(selectedMaterialIds.value.map((id) => deleteMaterialLedgerItem(sessionToken.value, id)));
    selectedMaterialIds.value = [];
  }, "已批量删除旧料记录。", "materials");
}

function downloadInventoryTemplate() {
  const csv = `${INVENTORY_TEMPLATE_COLUMNS.join(",")}\nKJ-0001,素圈手镯,黄金,足金9999,1,32.5,${stores.value[0]?.name || "总店"},0\n`;
  const blob = new Blob([`\uFEFF${csv}`], { type: "text/csv;charset=utf-8" });
  const url = window.URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "库存导入模板.csv";
  link.click();
  window.URL.revokeObjectURL(url);
}

function parseDelimitedRows(text: string) {
  const lines = text.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
  if (lines.length <= 1) return [];
  const delimiter = lines[0].includes("\t") ? "\t" : ",";
  const headers = lines[0].replace(/^\uFEFF/, "").split(delimiter).map((item) => item.trim());
  return lines.slice(1).map((line) => {
    const values = line.split(delimiter).map((item) => item.trim());
    return headers.reduce<Record<string, string>>((row, header, index) => {
      row[header] = values[index] || "";
      return row;
    }, {});
  });
}

function handleInventoryImportFile(event: Event) {
  if (!sessionToken.value) return;
  const input = event.target as HTMLInputElement;
  const file = input.files?.[0];
  if (!file) return;
  const reader = new FileReader();
  reader.onload = () => {
    const rows = parseDelimitedRows(String(reader.result || ""));
    const imported = rows
      .map((row) => {
        const store = stores.value.find((item) => item.name === row["门店"] || item.code === row["门店"]) || stores.value[0];
        return normalizeInventoryEntry({
          storeId: store?.id,
          storeName: store?.name || row["门店"],
          styleNo: row["款式编号"],
          name: row["商品名称"],
          category: row["品类"],
          purity: row["成色"],
          pieceCount: Number(row["件数"] || 0),
          weightGram: Number(row["重量"] || 0),
          unitCost: Number(row["成本"] || 0),
          status: "normal",
          source: "文件导入",
        });
      })
      .filter((item) => item.styleNo && item.name);
    if (!imported.length) {
      showNotice("warning", "没有识别到有效库存行，请按模板填写后再导入。");
      return;
    }
    void runAction(async () => {
      await Promise.all(imported.map((item) => createInventoryLedgerItem(sessionToken.value, {
        storeId: item.storeId,
        styleNo: item.styleNo,
        name: item.name,
        category: item.category,
        purity: item.purity,
        pieceCount: item.pieceCount,
        weightGram: item.weightGram,
        costAmount: item.unitCost,
        status: "in_stock",
        source: "file",
      })));
      inventoryImportInputKey.value += 1;
    }, `已通过接口导入 ${imported.length} 条库存记录。`, "inventory");
  };
  reader.readAsText(file, "utf-8");
}

function formatDateLabel(value: string) {
  if (!value) return "-";
  return value.includes("T") ? value.replace("T", " ").slice(0, 16) : value;
}

function formatCustomerName(name?: string, phone?: string) {
  const displayName = name?.trim() || "未填客户";
  return phone?.trim() ? `${displayName} · ${phone}` : displayName;
}

function productStatusLabel(status: ProductRecord["status"]) {
  return { active: "上架中", draft: "待完善", disabled: "已删除" }[status] || status;
}

function stockStatusLabel(status: string | undefined) {
  return {
    normal: "库存正常",
    low: "库存较低",
    review: "待确认",
    disabled: "已删除",
    out: "无库存",
  }[status || ""] || "未设置";
}

function storeStatusLabel(status: StoreRecord["status"]) {
  return { active: "营业中", pending: "待开业", disabled: "已删除" }[status] || status;
}

function userStatusLabel(status: UserAccount["status"]) {
  return { enabled: "启用中", disabled: "已删除", invited: "待激活" }[status] || status;
}

function memberStatusLabel(status: string) {
  return { active: "正常", inactive: "已删除", disabled: "已删除" }[status] || status || "未设置";
}

function orderStatusLabel(status: CashierOrderView["status"]) {
  return { paid: "已完成", pending: "待复核", refunded: "已取消" }[status] || status;
}

function recycleStatusLabel(status: RecycleOrderView["status"]) {
  return { draft: "待确认", confirmed: "已确认", cancelled: "已作废" }[status] || status;
}

function statusTone(status: string): MetricTone {
  if (["已完成", "已确认", "上架中", "营业中", "正常", "库存正常", "启用中"].includes(status)) return "emerald";
  if (["待复核", "待确认", "待完善", "待开业", "库存较低", "待激活"].includes(status)) return "gold";
  if (["已取消", "已作废", "已删除", "无库存"].includes(status)) return "danger";
  return "slate";
}

function showNotice(tone: NoticeTone, text: string) {
  notice.value = { tone, text };
}

function clearNotice() {
  notice.value = null;
}

function syncDocumentMeta() {
  const pageTitle = bootstrap.value ? `${pageMeta.value.title}｜金匠倌收银系统后台` : "金匠倌收银系统后台";
  document.title = pageTitle;
  const meta = document.querySelector('meta[name="description"]');
  if (meta) {
    meta.setAttribute(
      "content",
      bootstrap.value
        ? pageMeta.value.description
        : "金匠倌收银系统后台，用于维护门店、账号、会员、商品、收银、回收和系统设置。",
    );
  }
}

function goPage(page: PageId) {
  activePage.value = page;
  window.scrollTo({ top: 0, behavior: "smooth" });
  if (sessionToken.value) {
    void loadPageData(page);
  }
}

function canAccess(ability: AbilityCode) {
  return currentUser.value?.abilities.includes(ability) ?? false;
}

function selectStore(item: StoreRecord) {
  selectedStoreId.value = item.id;
  storeDraft.value = cloneStore(item);
  void loadStoreCustomerConfig(item.id);
}

async function loadStoreCustomerConfig(storeId: string) {
  if (!sessionToken.value || !storeId) return;
  try {
    const result = await fetchStoreCustomerConfig(sessionToken.value, storeId);
    storeCustomerDraft.value = {
      imageUrl: result.data.imageUrl || "",
      appointmentEnabled: result.data.appointmentEnabled ?? true,
      serviceTagsText: (result.data.serviceTags || []).join("、"),
      sortOrder: result.data.sortOrder || 0,
    };
  } catch {
    storeCustomerDraft.value = {
      imageUrl: "",
      appointmentEnabled: true,
      serviceTagsText: "",
      sortOrder: 0,
    };
  }
}

async function saveStoreCustomerAction() {
  if (!sessionToken.value || !selectedStoreId.value || !storeCustomerDraft.value) return;
  await runAction(async () => {
    await saveStoreCustomerConfig(sessionToken.value, selectedStoreId.value, {
      imageUrl: storeCustomerDraft.value!.imageUrl,
      appointmentEnabled: storeCustomerDraft.value!.appointmentEnabled,
      serviceTags: storeCustomerDraft.value!.serviceTagsText
        .split(/[、,，\s]+/)
        .map((tag) => tag.trim())
        .filter(Boolean),
      sortOrder: Number(storeCustomerDraft.value!.sortOrder) || 0,
    });
  }, "门店顾客端展示配置已保存。", "stores");
}

function selectUser(item: UserAccount) {
  selectedUserId.value = item.id;
  userDraft.value = cloneUser(item);
}

function selectRole(item: RoleTemplate) {
  selectedRoleId.value = item.id;
  roleDraft.value = cloneRole(item);
}

function selectCashier(item: CashierOrderView) {
  selectedCashierId.value = item.id;
  void loadSelectedCashierDetail(item.id);
}

function selectRecycle(item: RecycleOrderView) {
  selectedRecycleId.value = item.id;
  void loadSelectedRecycleDetail(item.id);
}

function selectMember(item: MemberProfile) {
  selectedMemberId.value = item.id;
  memberDraft.value = cloneMember(item);
}

function selectProduct(item: ProductRecord) {
  selectedProductId.value = item.id;
  productDraft.value = cloneProduct(item);
  productStoreSearch.value = "";
  productStorePickerOpen.value = false;
  productTagsText.value = item.tags.join("、");
}

function toggleUserStore(storeId: string) {
  if (!userDraft.value) return;
  if (userDraft.value.roleKey === "boss") {
    userDraft.value.storeIds = [];
    userDraft.value.storeNames = ["全部门店"];
    return;
  }
  const store = stores.value.find((item) => item.id === storeId);
  userDraft.value.storeIds = store ? [store.id] : [];
  userDraft.value.storeNames = store ? [store.name] : [];
}

function onUserRoleChange() {
  if (!userDraft.value) return;
  userDraft.value.roleKey = normalizeRoleKey(userDraft.value.roleKey);
  userDraft.value.roleName = userDraft.value.roleKey === "boss" ? "老板" : "店长";
  userDraft.value.dataScope = userDraft.value.roleKey === "boss" ? "all_stores" : "assigned_store";
  if (userDraft.value.roleKey === "boss") {
    userDraft.value.storeIds = [];
    userDraft.value.storeNames = ["全部门店"];
    return;
  }
  const firstStore = stores.value.find((store) => userDraft.value?.storeIds.includes(store.id)) ?? stores.value[0];
  userDraft.value.storeIds = firstStore ? [firstStore.id] : [];
  userDraft.value.storeNames = firstStore ? [firstStore.name] : [];
}

function syncProductStoreNames() {
  if (!productDraft.value) return;
  const current = new Set(productDraft.value.storeIds || []);
  productDraft.value.storeNames = stores.value.filter((item) => current.has(item.id)).map((item) => item.name);
}

function toggleProductStore(storeId: string) {
  if (!productDraft.value) return;
  const current = new Set(productDraft.value.storeIds || []);
  if (current.has(storeId)) {
    current.delete(storeId);
  } else {
    current.add(storeId);
  }
  productDraft.value.storeIds = Array.from(current);
  syncProductStoreNames();
  productStoreSearch.value = "";
  productStorePickerOpen.value = true;
}

function removeProductStore(storeId: string) {
  if (!productDraft.value) return;
  productDraft.value.storeIds = (productDraft.value.storeIds || []).filter((id) => id !== storeId);
  syncProductStoreNames();
}

function closeProductStorePickerSoon() {
  window.setTimeout(() => {
    productStorePickerOpen.value = false;
  }, 120);
}

function productStoreMeta(store: StoreRecord) {
  return [store.code, store.city, store.address].filter(Boolean).join(" · ");
}

function toggleRoleAbility(code: AbilityCode) {
  if (!roleDraft.value) return;
  const current = new Set(roleDraft.value.abilities);
  if (current.has(code)) {
    current.delete(code);
  } else {
    current.add(code);
  }
  roleDraft.value.abilities = Array.from(current);
}

function toggleProductSelection(productId: string) {
  const current = new Set(selectedProductIds.value);
  if (current.has(productId)) {
    current.delete(productId);
  } else {
    current.add(productId);
  }
  selectedProductIds.value = Array.from(current);
}

function toggleAllVisibleProducts() {
  if (allVisibleProductsSelected.value) {
    selectedProductIds.value = selectedProductIds.value.filter((id) => !products.value.some((product) => product.id === id));
    return;
  }
  const current = new Set(selectedProductIds.value);
  products.value.forEach((product) => current.add(product.id));
  selectedProductIds.value = Array.from(current);
}

function toggleMaterialSelection(materialId: string) {
  const current = new Set(selectedMaterialIds.value);
  if (current.has(materialId)) {
    current.delete(materialId);
  } else {
    current.add(materialId);
  }
  selectedMaterialIds.value = Array.from(current);
}

function toggleAllVisibleMaterials() {
  if (allVisibleMaterialsSelected.value) {
    selectedMaterialIds.value = selectedMaterialIds.value.filter((id) => !deletableMaterialRows.value.some((item) => item.id === id));
    return;
  }
  const current = new Set(selectedMaterialIds.value);
  deletableMaterialRows.value.forEach((item) => current.add(item.id));
  selectedMaterialIds.value = Array.from(current);
}

function handleProductImportFile(event: Event) {
  const input = event.target as HTMLInputElement;
  productImportFile.value = input.files?.[0] ?? null;
}

async function downloadImportTemplate() {
  if (!sessionToken.value) return;
  if (!productImportStoreId.value) {
    showNotice("warning", "请先选择导入门店，再下载对应模板。");
    return;
  }
  actionPending.value = true;
  try {
    const blob = await downloadProductImportTemplate(sessionToken.value, productImportStoreId.value);
    const url = window.URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    const store = stores.value.find((item) => item.id === productImportStoreId.value);
    link.download = `${store?.name || "门店"}商品导入模板.xlsx`;
    link.click();
    window.URL.revokeObjectURL(url);
    showNotice("success", "商品导入模板已下载。");
  } catch (error) {
    showNotice("warning", error instanceof Error ? error.message : "模板下载失败。");
  } finally {
    actionPending.value = false;
  }
}

async function importProductExcel() {
  if (!sessionToken.value) return;
  if (!productImportStoreId.value) {
    showNotice("warning", "请先选择要导入的门店。");
    return;
  }
  if (!productImportFile.value) {
    showNotice("warning", "请先选择 Excel 文件。");
    return;
  }
  await runAction(async () => {
    const result = await importProducts(sessionToken.value, productImportStoreId.value, productImportFile.value as File);
    productImportResult.value = result.data;
    productImportFile.value = null;
    productImportInputKey.value += 1;
  }, "商品导入完成，列表已刷新。", "products");
}

async function loadConsole(token: string, nextPage?: PageId) {
  loading.value = true;
  try {
    const result = await fetchAdminConsole(token);
    bootstrap.value = result.data;
    sourceMode.value = result.source;
    sessionToken.value = token;
    activePage.value = nextPage || activePage.value || "dashboard";
    seedModuleData(result.data);
    syncDrafts(result.data);
    window.sessionStorage.setItem(SESSION_STORAGE_KEY, token);
    await loadPageData(activePage.value);
  } catch (error) {
    window.sessionStorage.removeItem(SESSION_STORAGE_KEY);
    sessionToken.value = "";
    bootstrap.value = null;
    showNotice("warning", error instanceof Error ? error.message : "后台数据加载失败，请重新登录。");
    throw error;
  } finally {
    loading.value = false;
  }
}

async function runAction(action: () => Promise<void>, successText: string, page = activePage.value) {
  actionPending.value = true;
  clearNotice();
  try {
    await action();
    await loadPageData(page);
    syncDrafts(bootstrap.value);
    showNotice("success", successText);
  } catch (error) {
    showNotice("warning", error instanceof Error ? error.message : "操作失败，请稍后重试。");
  } finally {
    actionPending.value = false;
  }
}

async function loadSelectedCashierDetail(id = selectedCashierId.value) {
  if (!sessionToken.value || !id) return;
  try {
    const result = await fetchCashierOrderDetail(sessionToken.value, id);
    selectedCashierDetail.value = result.data;
  } catch {
    selectedCashierDetail.value = selectedCashier.value ?? null;
  }
}

async function loadSelectedRecycleDetail(id = selectedRecycleId.value) {
  if (!sessionToken.value || !id) return;
  try {
    const result = await fetchRecycleOrderDetail(sessionToken.value, id);
    selectedRecycleDetail.value = result.data;
  } catch {
    selectedRecycleDetail.value = selectedRecycle.value ?? null;
  }
}

async function loadPageData(page = activePage.value) {
  if (!sessionToken.value) return;
  switch (page) {
    case "stores": {
      const result = await fetchStoreRecords(sessionToken.value, storeFilters);
      sourceMode.value = result.source;
      storeItems.value = result.data.items;
      break;
    }
    case "users": {
      const result = await fetchUserAccounts(sessionToken.value, userFilters);
      sourceMode.value = result.source;
      userItems.value = result.data.items.map((item) => cloneUser(item)).filter((item): item is UserAccount => Boolean(item));
      break;
    }
    case "roles": {
      const result = await fetchRoleTemplates(sessionToken.value);
      sourceMode.value = result.source;
      roleItems.value = result.data;
      break;
    }
    case "cashier": {
      const result = await fetchCashierOrders(sessionToken.value, cashierFilters);
      sourceMode.value = result.source;
      cashierItems.value = result.data.items;
      selectedCashierId.value = cashierItems.value.find((item) => item.id === selectedCashierId.value)?.id || cashierItems.value[0]?.id || "";
      await loadSelectedCashierDetail();
      break;
    }
    case "recycle": {
      const result = await fetchRecycleOrders(sessionToken.value, recycleFilters);
      sourceMode.value = result.source;
      recycleItems.value = result.data.items;
      selectedRecycleId.value = recycleItems.value.find((item) => item.id === selectedRecycleId.value)?.id || recycleItems.value[0]?.id || "";
      await loadSelectedRecycleDetail();
      break;
    }
    case "orders": {
      const result = await fetchOrderRows(sessionToken.value, orderFilters);
      sourceMode.value = result.source;
      orderItems.value = result.data.items;
      break;
    }
    case "members": {
      const result = await fetchMemberProfiles(sessionToken.value, memberFilters);
      sourceMode.value = result.source;
      memberItems.value = result.data.items;
      break;
    }
    case "products": {
      const result = await fetchProductRecords(sessionToken.value, productFilters);
      sourceMode.value = result.source;
      productItems.value = result.data.items;
      selectedProductIds.value = selectedProductIds.value.filter((id) => productItems.value.some((product) => product.id === id));
      break;
    }
    case "inventory": {
      const result = await fetchInventoryLedger(sessionToken.value);
      sourceMode.value = result.source;
      inventoryEntries.value = result.data.items.map((item) => normalizeInventoryEntry(item));
      break;
    }
    case "materials": {
      const result = await fetchMaterialLedger(sessionToken.value);
      sourceMode.value = result.source;
      materialEntries.value = result.data.items.map((item) => normalizeMaterialEntry(item));
      selectedMaterialIds.value = selectedMaterialIds.value.filter((id) => materialEntries.value.some((item) => item.id === id));
      break;
    }
    case "appointments": {
      const result = await fetchAppointments(sessionToken.value, appointmentFilters);
      sourceMode.value = result.source;
      appointmentItems.value = result.data.items;
      selectedAppointmentId.value = appointmentItems.value.find((item) => item.id === selectedAppointmentId.value)?.id || appointmentItems.value[0]?.id || "";
      await loadSelectedAppointmentDetail();
      await loadAppointmentRules();
      break;
    }
    case "customer-home": {
      const result = await fetchCustomerHomeConfig(sessionToken.value);
      sourceMode.value = result.source;
      customerHomeConfig.value = result.data;
      homeBanners.value = result.data.banners || [];
      const { banners: _banners, ...copyDraft } = result.data;
      homeConfigDraft.value = {
        brandName: copyDraft.brandName || "",
        brandSlogan1: copyDraft.brandSlogan1 || "",
        brandSlogan2: copyDraft.brandSlogan2 || "",
        serviceCopy: copyDraft.serviceCopy || "",
        entryStyleText: copyDraft.entryStyleText || "",
        entryFeeText: copyDraft.entryFeeText || "",
        servicePhone: copyDraft.servicePhone || "",
        appointmentNotes: copyDraft.appointmentNotes || "",
        serviceIntro: copyDraft.serviceIntro || "",
        locationPermissionNote: copyDraft.locationPermissionNote || "",
      };
      const recycleResult = await fetchCustomerRecycleInfo(sessionToken.value);
      recycleInfoDraft.value = {
        title: recycleResult.data.title || "",
        intro: recycleResult.data.intro || "",
        imageUrl: recycleResult.data.imageUrl || "",
        process: recycleResult.data.process?.length ? recycleResult.data.process.map((p) => ({ step: p.step, title: p.title, desc: p.desc || "" })) : [],
        services: recycleResult.data.services?.length ? recycleResult.data.services.map((s) => ({ icon: s.icon || "", title: s.title, desc: s.desc || "" })) : [],
        notices: recycleResult.data.notices?.length ? [...recycleResult.data.notices] : [],
      };
      break;
    }
    case "customer-styles": {
      const result = await fetchCustomerStyles(sessionToken.value, styleFilters);
      sourceMode.value = result.source;
      styleItems.value = result.data.items || [];
      styleTotal.value = result.data.total || 0;
      if (styleCategories.value.length === 0) {
        try {
          const cats = await fetchCustomerStyleCategories(sessionToken.value);
          styleCategories.value = cats.data || [];
        } catch {
          styleCategories.value = [];
        }
      }
      break;
    }
    default:
      break;
  }
  syncDrafts(bootstrap.value);
}

function resetFilters(page: PageId) {
  if (page === "inventory") {
    inventoryFilters.storeId = "";
    inventoryFilters.keyword = "";
    inventoryFilters.category = "";
    inventoryFilters.purity = "";
    return;
  }
  if (page === "materials") {
    materialFilters.storeId = "";
    materialFilters.keyword = "";
    materialFilters.type = "";
    materialFilters.dateFrom = "";
    materialFilters.dateTo = "";
    return;
  }
  if (page === "customer-styles") {
    styleFilters.category = "";
    styleFilters.status = "";
    styleFilters.keyword = "";
    styleFilters.page = 1;
    return;
  }
  if (page === "appointments") {
    appointmentFilters.status = "";
    appointmentFilters.storeId = "";
    appointmentFilters.phone = "";
    appointmentFilters.serviceType = "";
    appointmentFilters.dateFrom = "";
    appointmentFilters.dateTo = "";
    appointmentFilters.page = 1;
    appointmentFilters.pageSize = 50;
    return;
  }
  const target = page === "stores" ? storeFilters : page === "users" ? userFilters : page === "cashier" ? cashierFilters : page === "recycle" ? recycleFilters : page === "orders" ? orderFilters : page === "members" ? memberFilters : productFilters;
  target.page = 1;
  target.pageSize = 50;
  target.storeId = "";
  target.status = "";
  target.dateFrom = "";
  target.dateTo = "";
  target.keyword = "";
  if (sessionToken.value) {
    void loadPageData(page);
  }
}

function openRecyclePreview(orderNo: string, assets: AttachmentAsset[], index: number) {
  recyclePreview.assets = assets.filter((item) => item.hasPreview && item.previewUrl);
  recyclePreview.orderNo = orderNo;
  recyclePreview.index = index;
  recyclePreview.open = recyclePreview.assets.length > 0;
}

function closeRecyclePreview() {
  recyclePreview.open = false;
}

function showNextPreview(delta: number) {
  if (!recyclePreview.assets.length) return;
  recyclePreview.index = (recyclePreview.index + delta + recyclePreview.assets.length) % recyclePreview.assets.length;
}

async function voidCurrentCashier() {
  if (!sessionToken.value || !currentCashierDetail.value || currentCashierDetail.value.status === "refunded") return;
  const reason = window.prompt("请输入收银单作废原因");
  if (!reason || !reason.trim()) return;
  await runAction(async () => {
    await voidCashierOrder(sessionToken.value, currentCashierDetail.value!.id, reason.trim());
  }, "收银单已作废。", "cashier");
}

async function cancelCurrentRecycle() {
  if (!sessionToken.value || !currentRecycleDetail.value || currentRecycleDetail.value.status !== "draft") return;
  const reason = window.prompt("请输入回收单取消原因");
  if (!reason || !reason.trim()) return;
  await runAction(async () => {
    await cancelRecycleOrder(sessionToken.value, currentRecycleDetail.value!.id, reason.trim());
  }, "回收单已取消。", "recycle");
}

async function openOrderRow(order: OrderRow) {
  if (order.type === "cashier") {
    activePage.value = "cashier";
    selectedCashierId.value = order.id;
    await loadPageData("cashier");
    await loadSelectedCashierDetail(order.id);
    return;
  }
  activePage.value = "recycle";
  selectedRecycleId.value = order.id;
  await loadPageData("recycle");
  await loadSelectedRecycleDetail(order.id);
}

async function submitLogin(roleKey: RoleKey = loginForm.roleKey) {
  loginPending.value = true;
  loginError.value = "";
  clearNotice();
  try {
    loginForm.roleKey = roleKey;
    const result = await loginAdmin(loginForm.phone.trim(), loginForm.password, roleKey);
    sourceMode.value = result.source;
    await loadConsole(result.data.token, "dashboard");
    showNotice("success", "已进入门店后台。");
  } catch (error) {
    loginError.value = error instanceof Error ? error.message : "登录失败，请检查账号密码。";
  } finally {
    loginPending.value = false;
  }
}

function logout() {
  bootstrap.value = null;
  sessionToken.value = "";
  activePage.value = "dashboard";
  window.sessionStorage.removeItem(SESSION_STORAGE_KEY);
}

async function createStore() {
  if (!sessionToken.value) return;
  await runAction(async () => {
    const result = await createStoreRecord(sessionToken.value);
    selectedStoreId.value = result.data.id;
  }, "门店已新增。", "stores");
}

async function saveStore() {
  if (!sessionToken.value || !storeDraft.value) return;
  await runAction(async () => {
    await saveStoreRecord(sessionToken.value, storeDraft.value as StoreRecord);
  }, "门店资料已保存。", "stores");
}

async function disableStore() {
  if (!sessionToken.value || !storeDraft.value) return;
  await runAction(async () => {
    await disableStoreRecord(sessionToken.value, storeDraft.value!.id);
  }, "门店已删除。", "stores");
}

// --- 预约管理 ---

function selectAppointment(item: AppointmentRecord) {
  selectedAppointmentId.value = item.id;
  void loadSelectedAppointmentDetail(item.id);
}

async function loadSelectedAppointmentDetail(id = selectedAppointmentId.value) {
  if (!sessionToken.value || !id) {
    selectedAppointmentDetail.value = null;
    return;
  }
  try {
    const result = await fetchAppointmentDetail(sessionToken.value, id);
    selectedAppointmentDetail.value = result.data;
    staffNoteDraft.value = result.data.staffNote || "";
  } catch {
    selectedAppointmentDetail.value = null;
  }
}

const APPT_STATUS_LABELS: Record<string, string> = {
  PENDING: "待确认",
  CONFIRMED: "已确认",
  ARRIVED: "已到店",
  COMPLETED: "已完成",
  CANCELLED: "已取消",
  NO_SHOW: "未到店",
  TERMINATED: "已终止",
};

const APPT_STATUS_TONES: Record<string, string> = {
  PENDING: "gold",
  CONFIRMED: "emerald",
  ARRIVED: "emerald",
  COMPLETED: "slate",
  CANCELLED: "danger",
  NO_SHOW: "danger",
  TERMINATED: "danger",
};

function apptStatusLabel(status: string): string {
  return APPT_STATUS_LABELS[status] || status;
}

function apptStatusTone(status: string): string {
  return APPT_STATUS_TONES[status] || "slate";
}

const SERVICE_TYPE_LABELS: Record<string, string> = {
  OLD_FOR_NEW: "以旧换新",
  REPAIR: "维修保养",
  CONSULT: "咨询鉴定",
  RECYCLE: "黄金回收",
};

function serviceTypeLabel(type: string): string {
  return SERVICE_TYPE_LABELS[type] || type;
}

function bannerLinkLabel(linkType: string): string {
  const labels: Record<string, string> = {
    none: "不跳转",
    styles: "款式列表",
    stores: "门店列表",
    booking: "预约页",
    custom: "自定义页面",
  };
  return labels[linkType] || linkType || "不跳转";
}

async function confirmAppointment(id: string) {
  if (!sessionToken.value) return;
  await runAction(async () => {
    await appointmentAction(sessionToken.value, id, "confirm");
  }, "预约已确认。", "appointments");
}

async function arriveAppointment(id: string) {
  if (!sessionToken.value) return;
  await runAction(async () => {
    await appointmentAction(sessionToken.value, id, "arrive");
  }, "已标记到店。", "appointments");
}

async function completeAppointment(id: string) {
  if (!sessionToken.value) return;
  await runAction(async () => {
    await appointmentAction(sessionToken.value, id, "complete");
  }, "预约已完成。", "appointments");
}

async function noShowAppointment(id: string) {
  if (!sessionToken.value) return;
  await runAction(async () => {
    await appointmentAction(sessionToken.value, id, "no-show");
  }, "已标记未到店。", "appointments");
}

function openCancelDialog() {
  cancelReasonDraft.value = "";
  cancelDialogOpen.value = true;
}

async function doCancelAppointment() {
  if (!sessionToken.value || !selectedAppointmentDetail.value) return;
  const reason = cancelReasonDraft.value.trim();
  if (!reason) {
    showNotice("warning", "请填写取消原因。");
    return;
  }
  cancelDialogOpen.value = false;
  await runAction(async () => {
    await cancelAppointment(sessionToken.value, selectedAppointmentDetail.value!.id, reason);
  }, "预约已取消。", "appointments");
}

async function saveStaffNote() {
  if (!sessionToken.value || !selectedAppointmentDetail.value) return;
  actionPending.value = true;
  clearNotice();
  try {
    await updateAppointmentStaffNote(sessionToken.value, selectedAppointmentDetail.value.id, staffNoteDraft.value);
    showNotice("success", "内部备注已保存。");
  } catch (error) {
    showNotice("warning", error instanceof Error ? error.message : "保存失败，请稍后重试。");
  } finally {
    actionPending.value = false;
  }
}

async function loadAppointmentRules() {
  if (!sessionToken.value) return;
  try {
    const result = await fetchAppointmentRules(sessionToken.value);
    appointmentRules.value = result.data;
    appointmentRulesDraft.value = { ...result.data };
  } catch {
    // 静默失败，不阻断页面加载
  }
}

function startEditRules() {
  if (appointmentRules.value) {
    appointmentRulesDraft.value = { ...appointmentRules.value };
    rulesEditing.value = true;
  }
}

function cancelEditRules() {
  rulesEditing.value = false;
  if (appointmentRules.value) {
    appointmentRulesDraft.value = { ...appointmentRules.value };
  }
}

function toggleServiceType(type: string) {
  if (!appointmentRulesDraft.value) return;
  const list = appointmentRulesDraft.value.bookableServiceTypes;
  const idx = list.indexOf(type);
  if (idx >= 0) {
    list.splice(idx, 1);
  } else {
    list.push(type);
  }
}

async function saveAppointmentRulesAction() {
  if (!sessionToken.value || !appointmentRulesDraft.value) return;
  actionPending.value = true;
  clearNotice();
  try {
    const { updatedAt: _, ...rules } = appointmentRulesDraft.value;
    void _;
    const result = await saveAppointmentRules(sessionToken.value, rules);
    appointmentRules.value = result.data;
    appointmentRulesDraft.value = { ...result.data };
    rulesEditing.value = false;
    showNotice("success", "预约规则已保存。");
  } catch (error) {
    showNotice("warning", error instanceof Error ? error.message : "保存失败，请稍后重试。");
  } finally {
    actionPending.value = false;
  }
}

// --- 顾客端内容管理：首页配置 / Banner / 款式工费 / 图片上传 ---

function pickCustomerImage(scene: UploadScene, maxMB: number, onDone: (url: string) => void) {
  const input = document.createElement("input");
  input.type = "file";
  input.accept = "image/jpeg,image/png,image/webp";
  input.onchange = async () => {
    const file = input.files?.[0];
    if (!file) return;
    if (file.size > maxMB * 1024 * 1024) {
      showNotice("warning", `图片不能超过 ${maxMB}MB。`);
      return;
    }
    if (!sessionToken.value) return;
    uploadPending.value = true;
    clearNotice();
    try {
      const result = await uploadCustomerImage(sessionToken.value, file, scene);
      onDone(result.data.url);
      showNotice("success", "图片已上传。");
    } catch (error) {
      showNotice("warning", error instanceof Error ? error.message : "图片上传失败。");
    } finally {
      uploadPending.value = false;
    }
  };
  input.click();
}

async function saveHomeConfigAction() {
  if (!sessionToken.value || !homeConfigDraft.value) return;
  actionPending.value = true;
  clearNotice();
  try {
    const result = await saveCustomerHomeConfig(sessionToken.value, homeConfigDraft.value);
    customerHomeConfig.value = result.data;
    homeBanners.value = result.data.banners || [];
    showNotice("success", "顾客端首页配置已保存。");
  } catch (error) {
    showNotice("warning", error instanceof Error ? error.message : "保存失败，请稍后重试。");
  } finally {
    actionPending.value = false;
  }
}

// --- 回收服务介绍编辑 ---

async function saveRecycleInfoAction() {
  if (!sessionToken.value || !recycleInfoDraft.value) return;
  if (!recycleInfoDraft.value.title.trim()) {
    showNotice("warning", "回收服务标题不能为空。");
    return;
  }
  actionPending.value = true;
  clearNotice();
  try {
    const result = await saveCustomerRecycleInfo(sessionToken.value, {
      title: recycleInfoDraft.value.title,
      intro: recycleInfoDraft.value.intro || "",
      imageUrl: recycleInfoDraft.value.imageUrl || "",
      process: recycleInfoDraft.value.process.map((p, i) => ({ step: i + 1, title: p.title, desc: p.desc || "" })),
      services: recycleInfoDraft.value.services.map((s) => ({ icon: s.icon || "", title: s.title, desc: s.desc || "" })),
      notices: recycleInfoDraft.value.notices.filter((n) => n.trim() !== ""),
    });
    recycleInfoDraft.value = result.data;
    showNotice("success", "回收服务介绍已保存。");
  } catch (error) {
    showNotice("warning", error instanceof Error ? error.message : "保存失败，请稍后重试。");
  } finally {
    actionPending.value = false;
  }
}

function addRecycleProcessStep() {
  if (!recycleInfoDraft.value) return;
  recycleInfoDraft.value.process.push({ step: recycleInfoDraft.value.process.length + 1, title: "", desc: "" });
}
function removeRecycleProcessStep(index: number) {
  if (!recycleInfoDraft.value) return;
  recycleInfoDraft.value.process.splice(index, 1);
  recycleInfoDraft.value.process.forEach((p, i) => (p.step = i + 1));
}
function moveRecycleProcessStep(index: number, dir: -1 | 1) {
  if (!recycleInfoDraft.value) return;
  const target = index + dir;
  if (target < 0 || target >= recycleInfoDraft.value.process.length) return;
  const arr = recycleInfoDraft.value.process;
  [arr[index], arr[target]] = [arr[target], arr[index]];
  arr.forEach((p, i) => (p.step = i + 1));
}
function addRecycleServiceItem() {
  if (!recycleInfoDraft.value) return;
  recycleInfoDraft.value.services.push({ icon: "", title: "", desc: "" });
}
function removeRecycleServiceItem(index: number) {
  if (!recycleInfoDraft.value) return;
  recycleInfoDraft.value.services.splice(index, 1);
}
function moveRecycleServiceItem(index: number, dir: -1 | 1) {
  if (!recycleInfoDraft.value) return;
  const target = index + dir;
  if (target < 0 || target >= recycleInfoDraft.value.services.length) return;
  const arr = recycleInfoDraft.value.services;
  [arr[index], arr[target]] = [arr[target], arr[index]];
}
function addRecycleNotice() {
  if (!recycleInfoDraft.value) return;
  recycleInfoDraft.value.notices.push("");
}
function removeRecycleNotice(index: number) {
  if (!recycleInfoDraft.value) return;
  recycleInfoDraft.value.notices.splice(index, 1);
}
function moveRecycleNotice(index: number, dir: -1 | 1) {
  if (!recycleInfoDraft.value) return;
  const target = index + dir;
  if (target < 0 || target >= recycleInfoDraft.value.notices.length) return;
  const arr = recycleInfoDraft.value.notices;
  [arr[index], arr[target]] = [arr[target], arr[index]];
}

function openCreateBanner() {
  bannerEditingId.value = "";
  bannerDraft.value = {
    title: "",
    subtitle: "",
    imageUrl: "",
    linkType: "none",
    linkTarget: "",
    sortOrder: homeBanners.value.length + 1,
    enabled: true,
  };
  bannerDialogOpen.value = true;
}

function openEditBanner(banner: CustomerHomeBanner) {
  bannerEditingId.value = banner.id;
  bannerDraft.value = {
    title: banner.title,
    subtitle: banner.subtitle || "",
    imageUrl: banner.imageUrl,
    linkType: banner.linkType || "none",
    linkTarget: banner.linkTarget || "",
    sortOrder: banner.sortOrder,
    enabled: banner.enabled,
  };
  bannerDialogOpen.value = true;
}

async function saveBannerAction() {
  if (!sessionToken.value || !bannerDraft.value) return;
  if (!bannerDraft.value.imageUrl.trim()) {
    showNotice("warning", "请先上传 Banner 图片。");
    return;
  }
  await runAction(async () => {
    if (bannerEditingId.value) {
      await saveCustomerHomeBanner(sessionToken.value, bannerEditingId.value, bannerDraft.value!);
    } else {
      await createCustomerHomeBanner(sessionToken.value, bannerDraft.value!);
    }
    bannerDialogOpen.value = false;
  }, "Banner 已保存。", "customer-home");
}

async function removeBanner(banner: CustomerHomeBanner) {
  if (!sessionToken.value) return;
  if (!window.confirm(`确定删除 Banner「${banner.title || banner.imageUrl}」吗？`)) return;
  await runAction(async () => {
    await deleteCustomerHomeBanner(sessionToken.value, banner.id);
  }, "Banner 已删除。", "customer-home");
}

async function toggleBannerEnabled(banner: CustomerHomeBanner) {
  if (!sessionToken.value) return;
  await runAction(async () => {
    await saveCustomerHomeBanner(sessionToken.value, banner.id, {
      title: banner.title,
      subtitle: banner.subtitle || "",
      imageUrl: banner.imageUrl,
      linkType: banner.linkType || "none",
      linkTarget: banner.linkTarget || "",
      sortOrder: banner.sortOrder,
      enabled: !banner.enabled,
    });
  }, banner.enabled ? "Banner 已停用。" : "Banner 已启用。", "customer-home");
}

function styleStatusLabel(status: string) {
  return status === "active" ? "上架" : "下架";
}

function openCreateStyle() {
  styleEditingId.value = "";
  styleDraft.value = {
    name: "",
    sku: "",
    category: styleCategories.value[0] || "",
    imageUrl: "",
    images: [],
    detailImages: [],
    purity: "",
    retailPrice: 0,
    gramWeight: 0,
    laborFeeRef: "",
    description: "",
    laborFeeNote: "",
    applicableServiceTypes: ["OLD_FOR_NEW", "REPAIR", "CONSULT", "RECYCLE"],
    recommendedStoreRule: "nearest",
    sortOrder: 0,
    isRecommended: false,
    isHot: false,
    status: "active",
    tags: [],
    storeIds: [],
  };
  styleDialogOpen.value = true;
}

function openEditStyle(record: CustomerStyleRecord) {
  styleEditingId.value = record.id;
  styleDraft.value = {
    ...record,
    images: [...(record.images || [])],
    detailImages: [...(record.detailImages || [])],
    applicableServiceTypes: [...(record.applicableServiceTypes || [])],
    tags: [...(record.tags || [])],
    storeIds: [...(record.storeIds || [])],
  };
  styleDialogOpen.value = true;
}

async function saveStyleAction() {
  if (!sessionToken.value || !styleDraft.value) return;
  if (!String(styleDraft.value.name || "").trim() || !String(styleDraft.value.imageUrl || "").trim()) {
    showNotice("warning", "款式名称和主图必填。");
    return;
  }
  const id = styleEditingId.value;
  await runAction(async () => {
    await saveCustomerStyle(sessionToken.value, id, styleDraft.value!);
    styleDialogOpen.value = false;
  }, id ? "款式已更新。" : "款式已新增。", "customer-styles");
}

async function removeStyle(record: CustomerStyleRecord) {
  if (!sessionToken.value) return;
  if (!window.confirm(`确定删除款式「${record.name}」吗？删除后顾客端将不再展示。`)) return;
  await runAction(async () => {
    await deleteCustomerStyle(sessionToken.value, record.id);
  }, "款式已删除。", "customer-styles");
}

function toggleStyleServiceType(type: string) {
  if (!styleDraft.value) return;
  const list = styleDraft.value.applicableServiceTypes || (styleDraft.value.applicableServiceTypes = []);
  const idx = list.indexOf(type);
  if (idx >= 0) {
    list.splice(idx, 1);
  } else {
    list.push(type);
  }
}

function moveStyleImage(index: number, offset: number) {
  if (!styleDraft.value) return;
  const list = styleDraft.value.images || [];
  const target = index + offset;
  if (target < 0 || target >= list.length) return;
  [list[index], list[target]] = [list[target], list[index]];
  styleDraft.value.images = [...list];
}

function moveStyleDetailImage(index: number, offset: number) {
  if (!styleDraft.value) return;
  const list = styleDraft.value.detailImages || [];
  const target = index + offset;
  if (target < 0 || target >= list.length) return;
  [list[index], list[target]] = [list[target], list[index]];
  styleDraft.value.detailImages = [...list];
}

function removeStyleDetailImage(index: number) {
  if (!styleDraft.value) return;
  const list = styleDraft.value.detailImages || [];
  list.splice(index, 1);
  styleDraft.value.detailImages = [...list];
}

function removeStyleImage(index: number) {
  if (!styleDraft.value) return;
  const list = styleDraft.value.images || [];
  list.splice(index, 1);
  styleDraft.value.images = [...list];
  if (styleDraft.value.images[0]) {
    styleDraft.value.imageUrl = styleDraft.value.images[0];
  }
}

async function createUser() {
  if (!sessionToken.value) return;
  await runAction(async () => {
    const result = await createUserAccount(sessionToken.value);
    selectedUserId.value = result.data.id;
  }, "账号已新增。", "users");
}

async function saveUser() {
  if (!sessionToken.value || !userDraft.value) return;
  userDraft.value.roleKey = normalizeRoleKey(userDraft.value.roleKey);
  userDraft.value.roleName = userDraft.value.roleKey === "boss" ? "老板" : "店长";
  userDraft.value.dataScope = userDraft.value.roleKey === "boss" ? "all_stores" : "assigned_store";
  if (!userDraft.value.phone.trim()) {
    showNotice("warning", "请填写登录手机号。");
    return;
  }
  if (userDraft.value.id.startsWith("user-new-") && !String(userDraft.value.password || "").trim()) {
    showNotice("warning", "新账号请填写初始登录密码。");
    return;
  }
  if (userDraft.value.roleKey === "shop_manager" && userDraft.value.storeIds.length !== 1) {
    showNotice("warning", "一个店长账号必须且只能绑定一个门店。");
    return;
  }
  await runAction(async () => {
    await saveUserAccount(sessionToken.value, userDraft.value as UserAccount);
  }, "账号资料已保存。", "users");
}

async function disableUser() {
  if (!sessionToken.value || !userDraft.value) return;
  await runAction(async () => {
    await disableUserAccount(sessionToken.value, userDraft.value!.id);
  }, "账号已删除，历史订单保留。", "users");
}

async function saveRole() {
  if (!sessionToken.value || !roleDraft.value) return;
  await runAction(async () => {
    await saveRoleTemplate(sessionToken.value, roleDraft.value as RoleTemplate);
  }, "身份范围已保存。", "roles");
}

async function createMember() {
  if (!sessionToken.value) return;
  await runAction(async () => {
    const result = await createMemberProfile(sessionToken.value);
    selectedMemberId.value = result.data.id;
  }, "会员已新增。", "members");
}

async function saveMember() {
  if (!sessionToken.value || !memberDraft.value) return;
  await runAction(async () => {
    await saveMemberProfile(sessionToken.value, memberDraft.value as MemberProfile);
  }, "会员资料已保存。", "members");
}

async function disableMember() {
  if (!sessionToken.value || !memberDraft.value) return;
  await runAction(async () => {
    await disableMemberProfile(sessionToken.value, memberDraft.value!.id);
  }, "会员已删除。", "members");
}

async function createProduct() {
  if (!sessionToken.value) return;
  await runAction(async () => {
    const result = await createProductRecord(sessionToken.value);
    selectedProductId.value = result.data.id;
  }, "商品已新增。", "products");
}

async function saveProduct() {
  if (!sessionToken.value || !productDraft.value) return;
  if (!productDraft.value.storeIds.length) {
    showNotice("warning", "请先选择至少一个适用门店。");
    productStorePickerOpen.value = true;
    return;
  }
  syncProductStoreNames();
  const payload: ProductRecord = {
    ...productDraft.value,
    price: Number(productDraft.value.price || 0),
    gramWeight: Number(productDraft.value.gramWeight || 0),
    inventory: Number(productDraft.value.inventory || 0),
    storeIds: [...productDraft.value.storeIds],
    storeNames: [...productDraft.value.storeNames],
    tags: splitTextList(productTagsText.value),
  };
  await runAction(async () => {
    await saveProductRecord(sessionToken.value, payload);
  }, "商品资料已保存。", "products");
}

async function disableProduct() {
  if (!sessionToken.value || !productDraft.value) return;
  await runAction(async () => {
    await disableProductRecord(sessionToken.value, productDraft.value!.id);
  }, "商品已删除，历史订单保留。", "products");
}

async function batchDeleteProducts() {
  if (!sessionToken.value || !selectedProductIds.value.length) return;
  await runAction(async () => {
    await deleteProductRecords(sessionToken.value, selectedProductIds.value);
    selectedProductIds.value = [];
  }, "已批量删除商品，历史订单保留。", "products");
}

async function saveSettings() {
  if (!sessionToken.value || !systemDraft.value) return;
  await runAction(async () => {
    await saveSystemProfile(sessionToken.value, systemDraft.value as SystemProfile);
  }, "系统设置已保存。", "settings");
}

onMounted(() => {
  loadLocalLedgers();
  syncDocumentMeta();
  const token = window.sessionStorage.getItem(SESSION_STORAGE_KEY);
  if (token) {
    void loadConsole(token, activePage.value);
  }
});

watchEffect(() => {
  syncDocumentMeta();
});
</script>

<template>
  <div class="admin-app">
    <section v-if="!bootstrap" class="login-shell">
      <div class="login-hero panel">
        <div>
          <p class="eyebrow">门店后台</p>
          <h1>金匠倌收银系统后台</h1>
          <p class="hero-copy">
            老板查看全部门店，店长只管理自己的门店。商品、收银、回收和订单数据实时同步。
          </p>
        </div>
      </div>

      <form class="login-panel panel" @submit.prevent="submitLogin(loginForm.roleKey)">
        <div class="panel-head">
          <div>
            <p class="eyebrow">登录</p>
            <h2>后台登录</h2>
            <p class="header-copy">请输入手机号和密码，再选择登录身份。</p>
          </div>
        </div>

        <label class="field">
          <span>手机号</span>
          <input v-model="loginForm.phone" autocomplete="username" inputmode="tel" placeholder="请输入手机号" />
        </label>

        <label class="field">
          <span>密码</span>
          <input v-model="loginForm.password" autocomplete="current-password" placeholder="请输入密码" type="password" />
        </label>

        <div class="login-role-actions">
          <button
            class="primary-btn"
            :class="{ muted: loginForm.roleKey !== 'boss' }"
            :disabled="loginPending"
            type="button"
            @click="submitLogin('boss')"
          >
            {{ loginPending && loginForm.roleKey === "boss" ? "登录中..." : "老板登录" }}
          </button>
          <button
            class="secondary-btn"
            :class="{ active: loginForm.roleKey === 'shop_manager' }"
            :disabled="loginPending"
            type="button"
            @click="submitLogin('shop_manager')"
          >
            {{ loginPending && loginForm.roleKey === "shop_manager" ? "登录中..." : "店长登录" }}
          </button>
        </div>

        <p v-if="loginError" class="state-text error">{{ loginError }}</p>
      </form>
    </section>

    <section v-else class="workspace-shell">
      <aside class="sidebar panel">
        <div class="sidebar-brand">
          <div class="brand-mark">金</div>
          <div>
            <strong>金匠倌收银系统后台</strong>
            <p>门店管理后台</p>
          </div>
        </div>

        <div class="sidebar-user">
          <span class="badge gold">{{ currentUser?.roleName || "门店账号" }}</span>
          <h3>{{ currentUser?.name }}</h3>
          <p>{{ currentUser?.account }} · {{ currentUser?.dataScope === "all_stores" ? "全部门店" : "门店范围" }}</p>
        </div>

        <nav class="nav-group">
          <p class="nav-group-title">功能菜单</p>
          <button
            v-for="item in visibleNavItems"
            :key="item.id"
            class="nav-item"
            :class="{ active: activePage === item.id }"
            type="button"
            @click="goPage(item.id)"
          >
            <span>{{ item.title }}</span>
            <small>{{ item.hint }}</small>
          </button>
        </nav>

        <div class="sidebar-note">
          <p>可见门店</p>
          <span>{{ visibleStoreNames }}</span>
          <span class="mode-source">当前身份：{{ currentUser?.roleName }}</span>
        </div>
      </aside>

      <main class="workspace-main">
        <header class="workspace-header">
          <div>
            <p class="eyebrow">后台管理</p>
            <h2>{{ pageMeta.title }}</h2>
            <p class="header-copy">{{ pageMeta.description }}</p>
          </div>
          <div class="header-actions">
            <span class="badge slate">更新于 {{ formatDateLabel(bootstrap.updatedAt) }}</span>
            <button class="secondary-btn" type="button" @click="logout">退出</button>
          </div>
        </header>

        <div v-if="loading" class="notice panel info">
          <strong>正在加载后台数据...</strong>
        </div>

        <div v-if="notice" class="notice panel" :class="notice.tone">
          <strong>{{ notice.text }}</strong>
          <button class="text-btn" type="button" @click="clearNotice">关闭</button>
        </div>

        <section v-if="activePage === 'dashboard'" class="page-grid">
          <div class="hero-banner panel">
            <div>
              <p class="eyebrow">经营概况</p>
              <h3>门店业务总览</h3>
              <p class="header-copy">按同一套业务数据源查看收银、回收、会员和商品情况。</p>
            </div>
            <div class="hero-summary">
              <span>门店范围</span>
              <strong>{{ visibleStoreNames }}</strong>
            </div>
          </div>

          <div class="stat-grid">
            <div v-for="metric in dashboardMetrics" :key="metric.label" class="stat-card panel" :class="metric.tone">
              <span>{{ metric.label }}</span>
              <strong>{{ metric.value }}</strong>
              <p>{{ metric.hint }}</p>
            </div>
          </div>

          <div class="quick-grid-simple">
            <button
              v-for="action in visibleQuickActions"
              :key="action.id"
              class="shortcut-card"
              type="button"
              @click="goPage(action.id)"
            >
              <strong>{{ action.title }}</strong>
              <p>{{ action.description }}</p>
            </button>
          </div>

          <div class="content-grid two-up">
            <div class="panel">
              <div class="panel-head">
                <div>
                  <h3>最近订单</h3>
                  <p class="header-copy">收银单和回收单合并展示，方便快速对账。</p>
                </div>
                <button class="text-btn" type="button" @click="goPage('orders')">查看全部</button>
              </div>
              <div class="stack-list">
                <div v-for="order in orderRows.slice(0, 6)" :key="`${order.type}-${order.id}`" class="list-card simple-row">
                  <div>
                    <span>{{ order.type === "cashier" ? "收银单" : "回收单" }} · {{ formatDateLabel(order.createdAt) }}</span>
                    <strong>{{ order.orderNo }}</strong>
                    <p>{{ order.customer }} · {{ order.storeName }} · {{ order.meta }}</p>
                  </div>
                  <div class="simple-row-end">
                    <strong>{{ formatCurrency(order.amount) }}</strong>
                    <span class="badge" :class="statusTone(order.statusLabel)">{{ order.statusLabel }}</span>
                  </div>
                </div>
              </div>
            </div>

            <div class="panel">
              <div class="panel-head">
                <div>
                  <h3>业务管理</h3>
                  <p class="header-copy">常用门店经营资料集中维护。</p>
                </div>
              </div>
              <div class="stack-list">
                <div class="list-card"><strong>门店</strong><p>新增、编辑、删除，维护营业门店资料。</p></div>
                <div class="list-card"><strong>账号与门店</strong><p>维护账号所属门店和可见范围。</p></div>
                <div class="list-card"><strong>会员与商品</strong><p>维护会员档案、商品目录、价格和库存。</p></div>
                <div class="list-card"><strong>收银与回收</strong><p>业务单据、客户、明细和回收照片统一归档。</p></div>
              </div>
            </div>
          </div>
        </section>

        <section v-if="activePage === 'stores'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">门店管理</p>
              <h3>门店资料</h3>
              <p class="header-copy">新增门店、编辑门店资料、删除门店。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending" @click="createStore">新增门店</button>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="storeFilters.status">
                <option value="">全部</option>
                <option value="active">营业中</option>
                <option value="pending">待开业</option>
              </select>
            </label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="storeFilters.keyword" placeholder="门店名 / 编码 / 负责人" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('stores')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('stores')">重置</button>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>门店</th>
                    <th>城市</th>
                    <th>负责人</th>
                    <th>状态</th>
                    <th>今日收银</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in stores"
                    :key="item.id"
                    :class="{ selected: selectedStore?.id === item.id }"
                    @click="selectStore(item)"
                  >
                    <td><strong>{{ item.name }}</strong><span>{{ item.code }}</span></td>
                    <td>{{ item.city }}</td>
                    <td>{{ item.managerName }}</td>
                    <td><span class="badge" :class="statusTone(storeStatusLabel(item.status))">{{ storeStatusLabel(item.status) }}</span></td>
                    <td>{{ formatCurrency(item.todayAmount) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="storeDraft" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">门店编辑</p>
                  <h3>{{ storeDraft.name || "新门店" }}</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field"><span>门店编码</span><input v-model="storeDraft.code" /></label>
                <label class="field"><span>门店名称</span><input v-model="storeDraft.name" /></label>
                <label class="field"><span>负责人</span><input v-model="storeDraft.managerName" /></label>
                <label class="field"><span>城市</span><input v-model="storeDraft.city" /></label>
                <label class="field"><span>地址</span><textarea v-model="storeDraft.address" rows="2"></textarea></label>
                <label class="field"><span>联系电话</span><input v-model="storeDraft.contactPhone" placeholder="顾客端显示的门店电话" /></label>
                <label class="field"><span>营业时间</span><input v-model="storeDraft.businessHours" placeholder="如 09:30-21:30" /></label>
                <label class="field"><span>经度</span><input v-model.number="storeDraft.longitude" type="number" step="0.000001" placeholder="如 116.397428" /></label>
                <label class="field"><span>纬度</span><input v-model.number="storeDraft.latitude" type="number" step="0.000001" placeholder="如 39.90923" /></label>
                <label class="field">
                  <span>状态</span>
                  <select v-model="storeDraft.status">
                    <option value="active">营业中</option>
                    <option value="pending">待开业</option>
                  </select>
                </label>
              </div>
              <div class="detail-actions">
                <button class="primary-btn" type="button" :disabled="actionPending" @click="saveStore">保存门店</button>
                <button class="secondary-btn" type="button" :disabled="actionPending" @click="disableStore">删除门店</button>
              </div>

              <!-- 顾客端展示配置（门店图片 / 预约开关 / 服务标签） -->
              <div v-if="storeCustomerDraft" class="form-card" style="margin-top: 20px; border-top: 1px dashed var(--border-color, #e2e8f0); padding-top: 16px;">
                <p class="eyebrow">顾客端展示配置</p>
                <label class="field">
                  <span>门店图片（顾客端列表/详情展示）</span>
                  <div v-if="storeCustomerDraft.imageUrl" style="margin-bottom: 8px;">
                    <img :src="storeCustomerDraft.imageUrl" alt="门店图片" style="max-width: 180px; max-height: 100px; object-fit: cover; border-radius: 8px; display: block;" />
                  </div>
                  <button class="secondary-btn" type="button" :disabled="uploadPending" @click="pickCustomerImage('store', 5, (url) => { if (storeCustomerDraft) storeCustomerDraft.imageUrl = url; })">
                    {{ uploadPending ? "上传中…" : (storeCustomerDraft.imageUrl ? "更换图片" : "上传图片") }}
                  </button>
                </label>
                <label class="field">
                  <span>预约开关（关闭后顾客端该门店不可预约）</span>
                  <select v-model="storeCustomerDraft.appointmentEnabled">
                    <option :value="true">开启预约</option>
                    <option :value="false">关闭预约</option>
                  </select>
                </label>
                <label class="field">
                  <span>服务标签（顿号/逗号分隔，如：以旧换新、维修保养）</span>
                  <input v-model="storeCustomerDraft.serviceTagsText" placeholder="以旧换新、维修保养" />
                </label>
                <label class="field">
                  <span>顾客端排序（数字越小越靠前）</span>
                  <input v-model.number="storeCustomerDraft.sortOrder" type="number" />
                </label>
                <button class="secondary-btn" type="button" :disabled="actionPending" @click="saveStoreCustomerAction">保存顾客端配置</button>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'users'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">账号管理</p>
              <h3>后台账号</h3>
              <p class="header-copy">新增老板或店长账号，填写手机号、密码和所属门店。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending" @click="createUser">新增账号</button>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="userFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="userFilters.status">
                <option value="">全部</option>
                <option value="enabled">启用中</option>
                <option value="invited">待激活</option>
              </select>
            </label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="userFilters.keyword" placeholder="姓名 / 账号 / 手机号" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('users')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('users')">重置</button>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>账号</th>
                    <th>身份</th>
                    <th>门店</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in users"
                    :key="item.id"
                    :class="{ selected: selectedUser?.id === item.id }"
                    @click="selectUser(item)"
                  >
                    <td><strong>{{ item.name }}</strong><span>{{ item.account }}</span></td>
                    <td>{{ item.roleName }}</td>
                    <td>{{ item.storeNames.join("、") || "未分配" }}</td>
                    <td><span class="badge" :class="statusTone(userStatusLabel(item.status))">{{ userStatusLabel(item.status) }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="userDraft" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">账号编辑</p>
                  <h3>{{ userDraft.name || "新账号" }}</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field"><span>姓名</span><input v-model="userDraft.name" /></label>
                <label class="field"><span>登录账号</span><input v-model="userDraft.account" /></label>
                <label class="field"><span>登录手机号</span><input v-model="userDraft.phone" inputmode="tel" /></label>
                <label class="field"><span>登录密码</span><input v-model="userDraft.password" autocomplete="new-password" placeholder="不修改可留空" type="password" /></label>
                <label class="field">
                  <span>登录身份</span>
                  <select v-model="userDraft.roleKey" @change="onUserRoleChange">
                    <option value="boss">老板</option>
                    <option value="shop_manager">店长</option>
                  </select>
                </label>
                <label class="field">
                  <span>状态</span>
                  <select v-model="userDraft.status">
                    <option value="enabled">启用中</option>
                    <option value="invited">待激活</option>
                  </select>
                </label>
              </div>
              <div class="ability-group">
                <header>
                  <strong>所属门店</strong>
                  <p>{{ userDraft.roleKey === "boss" ? "老板默认管理全部门店。" : "一个店长账号只能绑定一个门店。" }}</p>
                </header>
                <label v-for="store in stores" :key="store.id" class="ability-item">
                  <input
                    :checked="userDraft.roleKey === 'boss' || userDraft.storeIds.includes(store.id)"
                    :disabled="userDraft.roleKey === 'boss' || store.status === 'disabled'"
                    :type="userDraft.roleKey === 'boss' ? 'checkbox' : 'radio'"
                    name="user-store"
                    @change="toggleUserStore(store.id)"
                  />
                  <span>{{ store.name }}</span>
                </label>
              </div>
              <div class="detail-actions">
                <button class="primary-btn" type="button" :disabled="actionPending" @click="saveUser">保存账号</button>
                <button class="secondary-btn" type="button" :disabled="actionPending" @click="disableUser">删除账号</button>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'roles'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">身份范围</p>
              <h3>固定身份配置</h3>
              <p class="header-copy">系统只保留老板和店长两个身份。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending || !roleDraft" @click="saveRole">保存身份范围</button>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="stack-list">
              <button
                v-for="item in roles"
                :key="item.id"
                class="role-card"
                :class="{ active: selectedRole?.id === item.id }"
                type="button"
                @click="selectRole(item)"
              >
                <strong>{{ item.name }}</strong>
                <span>{{ item.description }}</span>
              </button>
            </div>

            <aside v-if="roleDraft" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">身份编辑</p>
                  <h3>{{ roleDraft.name }}</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field"><span>身份名称</span><input v-model="roleDraft.name" /></label>
                <p class="state-text">老板管理全部门店，店长只管理自己绑定的门店。</p>
                <label class="field"><span>身份说明</span><textarea v-model="roleDraft.description" rows="2"></textarea></label>
                <label class="field">
                  <span>数据范围</span>
                  <select v-model="roleDraft.dataScope" :disabled="roleDraft.key === 'boss'">
                    <option value="all_stores">全部门店</option>
                    <option value="assigned_store">所属门店</option>
                  </select>
                </label>
              </div>
              <div class="ability-stack">
                <div v-for="group in abilityGroups" :key="group.key" class="ability-group">
                  <header>
                    <strong>{{ group.label }}</strong>
                    <p>{{ group.description }}</p>
                  </header>
                  <label v-for="item in group.items" :key="item.code" class="ability-item">
                    <input :checked="roleDraft.abilities.includes(item.code)" type="checkbox" @change="toggleRoleAbility(item.code)" />
                    <span>{{ item.label }} · {{ item.description }}</span>
                  </label>
                </div>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'cashier'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">收银记录</p>
              <h3>收银单明细</h3>
              <p class="header-copy">查看客户、商品明细、金额、收款方式和操作员。</p>
            </div>
            <span class="badge gold">{{ cashierOrders.length }} 笔</span>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="cashierFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="cashierFilters.status">
                <option value="">全部</option>
                <option value="paid">已完成</option>
                <option value="pending">待复核</option>
                <option value="refunded">已取消</option>
              </select>
            </label>
            <label class="field compact-field"><span>开始日期</span><input v-model="cashierFilters.dateFrom" type="date" /></label>
            <label class="field compact-field"><span>结束日期</span><input v-model="cashierFilters.dateTo" type="date" /></label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="cashierFilters.keyword" placeholder="订单号 / 客户 / 手机号" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('cashier')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('cashier')">重置</button>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>收银单</th>
                    <th>门店</th>
                    <th>客户</th>
                    <th>金额</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in cashierOrders"
                    :key="item.id"
                    :class="{ selected: selectedCashier?.id === item.id }"
                    @click="selectCashier(item)"
                  >
                    <td><strong>{{ item.orderNo }}</strong><span>{{ formatDateLabel(item.createdAt) }}</span></td>
                    <td>{{ item.storeName }}</td>
                    <td>{{ formatCustomerName(item.customerName, item.customerPhone) }}</td>
                    <td>{{ formatCurrency(item.totalAmount) }}</td>
                    <td><span class="badge" :class="statusTone(orderStatusLabel(item.status))">{{ orderStatusLabel(item.status) }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="currentCashierDetail" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">收银单详情</p>
                  <h3>{{ currentCashierDetail.orderNo }}</h3>
                </div>
                <button
                  v-if="canAccess('order.view')"
                  class="secondary-btn"
                  type="button"
                  :disabled="actionPending || currentCashierDetail.status === 'refunded'"
                  @click="voidCurrentCashier"
                >
                  {{ currentCashierDetail.status === "refunded" ? "已作废" : "作废收银单" }}
                </button>
              </div>
              <div class="detail-grid">
                <p><strong>门店：</strong>{{ currentCashierDetail.storeName }}</p>
                <p><strong>客户：</strong>{{ formatCustomerName(currentCashierDetail.customerName, currentCashierDetail.customerPhone) }}</p>
                <p><strong>金额：</strong>{{ formatCurrency(currentCashierDetail.totalAmount) }}</p>
                <p><strong>实收：</strong>{{ formatCurrency(currentCashierDetail.paidAmount || currentCashierDetail.totalAmount) }}</p>
                <p><strong>收款方式：</strong>{{ currentCashierDetail.paymentMethod || "未设置" }}</p>
                <p><strong>操作员：</strong>{{ currentCashierDetail.createdBy }}</p>
              </div>
              <div class="list-card">
                <strong>商品明细</strong>
                <p v-if="currentCashierDetail.items?.length">
                  <span v-for="item in currentCashierDetail.items" :key="`${item.sku || item.name}-${item.quantity}`">
                    {{ item.name }} x{{ item.quantity }} · {{ formatCurrency(item.amount) }}<br />
                  </span>
                </p>
                <p v-else>{{ currentCashierDetail.itemSummary || "暂无商品明细" }}</p>
              </div>
              <div v-if="currentCashierDetail.voidReason" class="list-card warning-card">
                <strong>作废信息</strong>
                <p>{{ currentCashierDetail.voidReason }}<br />{{ currentCashierDetail.voidedBy || "-" }} · {{ currentCashierDetail.voidedAt || "-" }}</p>
              </div>
              <div class="list-card">
                <strong>备注</strong>
                <p>{{ currentCashierDetail.remark || "无备注" }}</p>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'recycle'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">回收录单</p>
              <h3>回收单与照片留档</h3>
              <p class="header-copy">查看客户、回收明细、确认金额和照片留档入口。</p>
            </div>
            <span class="badge emerald">{{ recycleOrders.length }} 笔</span>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="recycleFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="recycleFilters.status">
                <option value="">全部</option>
                <option value="draft">待确认</option>
                <option value="confirmed">已确认</option>
                <option value="cancelled">已作废</option>
              </select>
            </label>
            <label class="field compact-field"><span>开始日期</span><input v-model="recycleFilters.dateFrom" type="date" /></label>
            <label class="field compact-field"><span>结束日期</span><input v-model="recycleFilters.dateTo" type="date" /></label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="recycleFilters.keyword" placeholder="订单号 / 客户 / 手机号" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('recycle')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('recycle')">重置</button>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>回收单</th>
                    <th>门店</th>
                    <th>客户</th>
                    <th>照片</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in recycleOrders"
                    :key="item.id"
                    :class="{ selected: selectedRecycle?.id === item.id }"
                    @click="selectRecycle(item)"
                  >
                    <td><strong>{{ item.orderNo }}</strong><span>{{ formatDateLabel(item.createdAt) }}</span></td>
                    <td>{{ item.storeName }}</td>
                    <td>{{ formatCustomerName(item.customerName, item.customerPhone) }}</td>
                    <td>{{ item.photoCount }} 张</td>
                    <td><span class="badge" :class="statusTone(recycleStatusLabel(item.status))">{{ recycleStatusLabel(item.status) }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="currentRecycleDetail" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">回收单详情</p>
                  <h3>{{ currentRecycleDetail.orderNo }}</h3>
                </div>
                <button
                  class="secondary-btn"
                  type="button"
                  :disabled="actionPending || currentRecycleDetail.status !== 'draft'"
                  @click="cancelCurrentRecycle"
                >
                  {{ currentRecycleDetail.status === "draft" ? "取消回收单" : "不可取消" }}
                </button>
              </div>
              <div class="detail-grid">
                <p><strong>门店：</strong>{{ currentRecycleDetail.storeName }}</p>
                <p><strong>客户：</strong>{{ formatCustomerName(currentRecycleDetail.customerName, currentRecycleDetail.customerPhone) }}</p>
                <p><strong>预估金额：</strong>{{ formatCurrency(currentRecycleDetail.estimatedAmount) }}</p>
                <p><strong>确认金额：</strong>{{ formatCurrency(currentRecycleDetail.confirmedAmount || currentRecycleDetail.estimatedAmount) }}</p>
                <p><strong>照片数：</strong>{{ currentRecycleDetail.photoCount }} 张</p>
                <p><strong>操作员：</strong>{{ currentRecycleDetail.createdBy }}</p>
              </div>
              <div class="list-card">
                <strong>回收明细</strong>
                <p v-if="currentRecycleDetail.items?.length">
                  <span v-for="item in currentRecycleDetail.items" :key="`${item.category}-${item.purity}-${item.weightGram}`">
                    {{ item.category }} · {{ item.purity }} · {{ item.weightGram }}g<br />
                  </span>
                </p>
                <p v-else>{{ currentRecycleDetail.itemSummary || "暂无回收明细" }}</p>
              </div>
              <div class="list-card">
                <strong>照片留档</strong>
                <div v-if="currentRecycleDetail.attachments?.length" class="photo-grid">
                  <button
                    v-for="(asset, index) in currentRecycleDetail.attachments"
                    :key="asset.id"
                    class="photo-card"
                    :class="{ missing: !asset.hasPreview || !asset.previewUrl }"
                    type="button"
                    @click="asset.hasPreview && asset.previewUrl ? openRecyclePreview(currentRecycleDetail.orderNo, currentRecycleDetail.attachments || [], index) : undefined"
                  >
                    <img v-if="asset.hasPreview && asset.previewUrl" :src="asset.previewUrl" :alt="asset.fileName || '回收照片'" />
                    <div v-else class="photo-missing">历史图片缺失</div>
                    <span>{{ asset.fileName || "未命名图片" }}</span>
                  </button>
                </div>
                <p v-else>当前回收单暂无图片留档。</p>
              </div>
              <div v-if="currentRecycleDetail.cancelReason" class="list-card warning-card">
                <strong>取消信息</strong>
                <p>{{ currentRecycleDetail.cancelReason }}<br />{{ currentRecycleDetail.cancelledBy || "-" }} · {{ currentRecycleDetail.cancelledAt || "-" }}</p>
              </div>
              <div class="list-card">
                <strong>备注</strong>
                <p>{{ currentRecycleDetail.remark || "无备注" }}</p>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'orders'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">业务订单</p>
              <h3>订单对账</h3>
              <p class="header-copy">统一查看收银单和回收单，用于业务对账。</p>
            </div>
            <span class="badge slate">{{ orderRows.length }} 条</span>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="orderFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="orderFilters.status">
                <option value="">全部</option>
                <option value="paid">收银已完成</option>
                <option value="pending">收银待复核</option>
                <option value="refunded">收银已取消</option>
                <option value="draft">回收待确认</option>
                <option value="confirmed">回收已确认</option>
                <option value="cancelled">回收已作废</option>
              </select>
            </label>
            <label class="field compact-field"><span>开始日期</span><input v-model="orderFilters.dateFrom" type="date" /></label>
            <label class="field compact-field"><span>结束日期</span><input v-model="orderFilters.dateTo" type="date" /></label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="orderFilters.keyword" placeholder="订单号 / 客户 / 手机号" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('orders')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('orders')">重置</button>
          </div>

          <div class="stack-list">
            <button
              v-for="order in orderRows"
              :key="`${order.type}-${order.id}`"
              class="list-card simple-row order-link-card"
              type="button"
              @click="openOrderRow(order)"
            >
              <div class="order-row-main">
                <span>{{ order.type === "cashier" ? "收银单" : "回收单" }} · {{ formatDateLabel(order.createdAt) }}</span>
                <strong>{{ order.orderNo }}</strong>
                <p>{{ order.customer }} · {{ order.storeName }} · {{ order.meta }}</p>
              </div>
              <div class="simple-row-end">
                <strong>{{ formatCurrency(order.amount) }}</strong>
                <span class="badge" :class="statusTone(order.statusLabel)">{{ order.statusLabel }}</span>
              </div>
            </button>
          </div>
        </section>

        <section v-if="activePage === 'members'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">会员档案</p>
              <h3>会员管理</h3>
              <p class="header-copy">新增会员、修改门店归属、等级、偏好、备注，也可以删除会员。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending" @click="createMember">新增会员</button>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="memberFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="memberFilters.status">
                <option value="">全部</option>
                <option value="active">正常</option>
              </select>
            </label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="memberFilters.keyword" placeholder="姓名 / 手机号 / 等级" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('members')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('members')">重置</button>
          </div>

          <div class="inline-metrics">
            <div class="summary-pill"><span>会员数</span><strong>{{ memberStats.total }}</strong></div>
            <div class="summary-pill"><span>VIP</span><strong>{{ memberStats.vip }}</strong></div>
            <div class="summary-pill"><span>已实名</span><strong>{{ memberStats.verified }}</strong></div>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>会员</th>
                    <th>门店</th>
                    <th>等级</th>
                    <th>最近到店</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in members"
                    :key="item.id"
                    :class="{ selected: selectedMember?.id === item.id }"
                    @click="selectMember(item)"
                  >
                    <td><strong>{{ item.name }}</strong><span>{{ item.phone }}</span></td>
                    <td>{{ item.storeName }}</td>
                    <td>{{ item.level }}</td>
                    <td>{{ formatDateLabel(item.lastVisitAt) }}</td>
                    <td><span class="badge" :class="statusTone(memberStatusLabel(item.status))">{{ memberStatusLabel(item.status) }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="memberDraft" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">会员编辑</p>
                  <h3>{{ memberDraft.name || "新会员" }}</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field"><span>姓名</span><input v-model="memberDraft.name" /></label>
                <label class="field"><span>手机号</span><input v-model="memberDraft.phone" /></label>
                <label class="field">
                  <span>所属门店</span>
                  <select v-model="memberDraft.storeId">
                    <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                  </select>
                </label>
                <label class="field"><span>会员等级</span><input v-model="memberDraft.level" /></label>
                <label class="field"><span>偏好成色</span><input v-model="memberDraft.preferredPurity" /></label>
                <label class="field"><span>来源渠道</span><input v-model="memberDraft.sourceChannel" /></label>
                <label class="field"><span>跟进人</span><input v-model="memberDraft.managerName" /></label>
                <label class="field">
                  <span>状态</span>
                  <select v-model="memberDraft.status">
                    <option value="active">正常</option>
                  </select>
                </label>
                <label class="toggle-field">
                  <input v-model="memberDraft.idVerified" type="checkbox" />
                  <span>已核验身份</span>
                </label>
                <label class="field"><span>回访备注</span><textarea v-model="memberDraft.notes" rows="3"></textarea></label>
              </div>
              <div class="detail-actions">
                <button class="primary-btn" type="button" :disabled="actionPending" @click="saveMember">保存会员</button>
                <button class="secondary-btn" type="button" :disabled="actionPending" @click="disableMember">删除会员</button>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'products'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">商品目录</p>
              <h3>商品维护</h3>
              <p class="header-copy">新增商品、导入 Excel，维护价格、库存、状态和适用门店。</p>
            </div>
            <div class="header-actions">
              <button class="secondary-btn" type="button" :disabled="actionPending || !selectedProductIds.length" @click="batchDeleteProducts">
                批量删除 {{ selectedProductIds.length ? selectedProductIds.length : "" }}
              </button>
              <button class="primary-btn" type="button" :disabled="actionPending" @click="createProduct">新增商品</button>
            </div>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="productFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="productFilters.status">
                <option value="">全部</option>
                <option value="active">上架中</option>
                <option value="draft">待完善</option>
              </select>
            </label>
            <label class="field compact-field grow-field">
              <span>关键字</span>
              <input v-model="productFilters.keyword" placeholder="商品名 / SKU / 分类" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('products')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('products')">重置</button>
          </div>

          <div class="panel import-panel">
            <div>
              <p class="eyebrow">商品导入</p>
              <h3>按门店导入商品</h3>
              <p class="header-copy">先选择门店并下载对应模板，再上传 Excel。最后一列可填门店名称或留空。</p>
            </div>
            <div class="import-controls">
              <label class="field compact-field">
                <span>导入门店</span>
                <select v-model="productImportStoreId">
                  <option value="">请选择门店</option>
                  <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                </select>
              </label>
              <button class="secondary-btn" type="button" :disabled="actionPending || !productImportStoreId" @click="downloadImportTemplate">下载模板</button>
              <label class="file-picker">
                <input :key="productImportInputKey" accept=".xlsx" type="file" @change="handleProductImportFile" />
                <span>{{ productImportFile?.name || "选择 Excel" }}</span>
              </label>
              <button class="primary-btn" type="button" :disabled="actionPending || !productImportStoreId || !productImportFile" @click="importProductExcel">
                上传导入
              </button>
            </div>
            <div v-if="productImportResult" class="import-result">
              <strong>{{ productImportResult.storeName }} 导入完成</strong>
              <span>成功 {{ productImportResult.successCount }} 条，失败 {{ productImportResult.failureCount }} 条</span>
              <p v-for="failure in productImportResult.failures.slice(0, 4)" :key="`${failure.row}-${failure.reason}`">
                第 {{ failure.row }} 行：{{ failure.reason }}
              </p>
            </div>
          </div>

          <div class="inline-metrics">
            <div class="summary-pill"><span>全部商品</span><strong>{{ productStats.total }}</strong></div>
            <div class="summary-pill"><span>上架中</span><strong>{{ productStats.active }}</strong></div>
            <div class="summary-pill"><span>库存数量</span><strong>{{ productStats.inventory }}</strong></div>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th class="select-cell">
                      <input :checked="allVisibleProductsSelected" type="checkbox" @change="toggleAllVisibleProducts" />
                    </th>
                    <th>商品</th>
                    <th>分类</th>
                    <th>价格</th>
                    <th>库存</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in products"
                    :key="item.id"
                    :class="{ selected: selectedProduct?.id === item.id }"
                    @click="selectProduct(item)"
                  >
                    <td class="select-cell" @click.stop>
                      <input :checked="selectedProductIds.includes(item.id)" type="checkbox" @change="toggleProductSelection(item.id)" />
                    </td>
                    <td><strong>{{ item.name }}</strong><span>{{ item.sku }}</span></td>
                    <td>{{ item.category }}</td>
                    <td>{{ formatCurrency(item.price) }}</td>
                    <td>{{ item.inventory || 0 }}</td>
                    <td><span class="badge" :class="statusTone(productStatusLabel(item.status))">{{ productStatusLabel(item.status) }}</span></td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="productDraft" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">商品编辑</p>
                  <h3>{{ productDraft.name || "新商品" }}</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field"><span>商品名称</span><input v-model="productDraft.name" /></label>
                <label class="field"><span>商品编号</span><input v-model="productDraft.sku" /></label>
                <label class="field"><span>商品分类</span><input v-model="productDraft.category" /></label>
                <label class="field"><span>页面分类</span><input v-model="productDraft.categoryTab" /></label>
                <label class="field"><span>销售价格</span><input v-model.number="productDraft.price" min="0" type="number" /></label>
                <label class="field"><span>克重</span><input v-model.number="productDraft.gramWeight" min="0" step="0.01" type="number" /></label>
                <label class="field"><span>库存数量</span><input v-model.number="productDraft.inventory" min="0" type="number" /></label>
                <label class="field">
                  <span>商品状态</span>
                  <select v-model="productDraft.status">
                    <option value="active">上架中</option>
                    <option value="draft">待完善</option>
                  </select>
                </label>
                <label class="field">
                  <span>库存状态</span>
                  <select v-model="productDraft.stockStatus">
                    <option value="normal">库存正常</option>
                    <option value="low">库存较低</option>
                    <option value="review">待确认</option>
                    <option value="out">无库存</option>
                  </select>
                </label>
                <div class="field store-picker-field">
                  <span>适用门店</span>
                  <div class="store-picker">
                    <div v-if="selectedProductStores.length" class="store-chips">
                      <span v-for="store in selectedProductStores" :key="store.id" class="store-chip">
                        {{ store.name }}
                        <button type="button" :aria-label="`移除${store.name}`" @click="removeProductStore(store.id)">×</button>
                      </span>
                    </div>
                    <input
                      v-model="productStoreSearch"
                      placeholder="点击选择门店"
                      @focus="productStorePickerOpen = true"
                      @input="productStorePickerOpen = true"
                      @blur="closeProductStorePickerSoon"
                      @keydown.escape="productStorePickerOpen = false"
                    />
                    <div v-if="productStorePickerOpen" class="store-dropdown">
                      <button
                        v-for="store in productStoreOptions"
                        :key="store.id"
                        type="button"
                        :class="{ selected: productDraft.storeIds.includes(store.id) }"
                        @mousedown.prevent="toggleProductStore(store.id)"
                      >
                        <strong>{{ store.name }}</strong>
                        <span>{{ productStoreMeta(store) }}</span>
                        <em v-if="productDraft.storeIds.includes(store.id)">已选择</em>
                      </button>
                      <p v-if="!productStoreOptions.length" class="empty-dropdown">没有匹配的门店</p>
                    </div>
                  </div>
                  <small>保存后，该商品只在已选门店可用。</small>
                </div>
                <label class="field"><span>商品标签</span><textarea v-model="productTagsText" rows="2"></textarea></label>
              </div>
              <div class="detail-actions">
                <button class="primary-btn" type="button" :disabled="actionPending" @click="saveProduct">保存商品</button>
                <button class="secondary-btn" type="button" :disabled="actionPending" @click="disableProduct">删除商品</button>
                <span class="state-text">{{ stockStatusLabel(productDraft.stockStatus) }}</span>
              </div>
            </aside>
          </div>
        </section>

        <section v-if="activePage === 'inventory'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">库存系统</p>
              <h3>商品入库与库存查询</h3>
              <p class="header-copy">维护款式编号、品类、成色、件数、重量，支持手动入库和 CSV 文件导入。</p>
            </div>
            <span class="badge gold">{{ filteredInventoryRows.length }} 款</span>
          </div>

          <div class="stat-grid">
            <div class="stat-card panel gold"><span>款式数量</span><strong>{{ inventorySummary.styles }}</strong><p>当前筛选范围</p></div>
            <div class="stat-card panel emerald"><span>库存件数</span><strong>{{ inventorySummary.pieces }}</strong><p>商品目录 + 入库记录</p></div>
            <div class="stat-card panel slate"><span>库存重量</span><strong>{{ formatWeight(inventorySummary.weight) }}</strong><p>按克重汇总</p></div>
            <div class="stat-card panel slate"><span>库存成本</span><strong>{{ formatCurrency(inventorySummary.amount) }}</strong><p>按成本/售价估算</p></div>
          </div>

          <div class="content-grid two-up">
            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">商品入库</p>
                  <h3>新增入库记录</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field">
                  <span>门店</span>
                  <select v-model="inventoryDraft.storeId">
                    <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                  </select>
                </label>
                <label class="field"><span>款式编号</span><input v-model="inventoryDraft.styleNo" placeholder="例如 KJ-HJ-0001" /></label>
                <label class="field"><span>商品名称</span><input v-model="inventoryDraft.name" placeholder="例如 素圈手镯" /></label>
                <label class="field"><span>品类</span><input v-model="inventoryDraft.category" placeholder="黄金 / K金 / 钻石" /></label>
                <label class="field"><span>成色</span><input v-model="inventoryDraft.purity" placeholder="客户可自填，例如 足金9999" /></label>
                <label class="field"><span>件数</span><input v-model.number="inventoryDraft.pieceCount" min="0" type="number" /></label>
                <label class="field"><span>重量(g)</span><input v-model.number="inventoryDraft.weightGram" min="0" step="0.01" type="number" /></label>
                <label class="field"><span>成本/参考价</span><input v-model.number="inventoryDraft.unitCost" min="0" type="number" /></label>
                <label class="field">
                  <span>库存状态</span>
                  <select v-model="inventoryDraft.status">
                    <option value="normal">库存正常</option>
                    <option value="low">库存较低</option>
                    <option value="review">待盘点</option>
                    <option value="out">无库存</option>
                  </select>
                </label>
              </div>
              <div class="detail-actions">
                <button class="primary-btn" type="button" @click="addInventoryEntry">确认入库</button>
                <button class="secondary-btn" type="button" @click="resetInventoryDraft">清空</button>
              </div>
            </article>

            <article class="panel import-panel inventory-import-panel">
              <div>
                <p class="eyebrow">文件导入</p>
                <h3>库存 CSV 导入</h3>
                <p class="header-copy">导入字段：款式编号、商品名称、品类、成色、件数、重量、门店、成本。</p>
              </div>
              <div class="import-controls">
                <button class="secondary-btn" type="button" @click="downloadInventoryTemplate">下载模板</button>
                <label class="file-picker">
                  <input :key="inventoryImportInputKey" accept=".csv,.txt" type="file" @change="handleInventoryImportFile" />
                  <span>选择 CSV</span>
                </label>
              </div>
            </article>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="inventoryFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>品类</span>
              <select v-model="inventoryFilters.category">
                <option value="">全部品类</option>
                <option v-for="category in inventoryCategories" :key="category" :value="category">{{ category }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>成色</span>
              <select v-model="inventoryFilters.purity">
                <option value="">全部成色</option>
                <option v-for="purity in inventoryPurities" :key="purity" :value="purity">{{ purity }}</option>
              </select>
            </label>
            <label class="field compact-field grow-field"><span>关键字</span><input v-model="inventoryFilters.keyword" placeholder="款式编号 / 商品 / 门店" /></label>
            <button class="text-btn" type="button" @click="resetFilters('inventory')">重置</button>
          </div>

          <div class="table-card">
            <table class="data-table">
              <thead>
                <tr>
                  <th>款式编号</th>
                  <th>商品</th>
                  <th>品类/成色</th>
                  <th>件数</th>
                  <th>重量</th>
                  <th>门店</th>
                  <th>状态</th>
                  <th>来源</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in filteredInventoryRows" :key="item.id">
                  <td><strong>{{ item.styleNo }}</strong><span>{{ item.createdAt ? formatDateLabel(item.createdAt) : "商品目录" }}</span></td>
                  <td>{{ item.name }}</td>
                  <td>{{ item.category }}<br /><span>{{ item.purity }}</span></td>
                  <td>{{ item.pieceCount }}</td>
                  <td>{{ formatWeight(item.weightGram) }}</td>
                  <td>{{ item.storeName }}</td>
                  <td><span class="badge" :class="statusTone(inventoryStatusLabel(item.status))">{{ inventoryStatusLabel(item.status) }}</span></td>
                  <td>{{ item.source }}</td>
                  <td>
                    <span v-if="item.source !== '商品目录'" class="muted-text">后端台账</span>
                    <span v-else>随商品维护</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section v-if="activePage === 'materials'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">旧料管理</p>
              <h3>回收旧料、余料、抵押寄存</h3>
              <p class="header-copy">从回收单自动汇总旧料，也可登记余料和抵押寄存，查看今日/月度重量金额。</p>
            </div>
            <div class="header-actions">
              <button class="secondary-btn" type="button" :disabled="actionPending || !selectedMaterialIds.length" @click="batchDeleteMaterials">
                批量删除 {{ selectedMaterialIds.length ? selectedMaterialIds.length : "" }}
              </button>
              <span class="badge emerald">{{ materialRows.length }} 条</span>
            </div>
          </div>

          <div class="stat-grid">
            <div class="stat-card panel emerald"><span>今日旧料</span><strong>{{ formatWeight(materialSummary.todayWeight) }}</strong><p>{{ formatCurrency(materialSummary.todayAmount) }}</p></div>
            <div class="stat-card panel gold"><span>月度旧料</span><strong>{{ formatWeight(materialSummary.monthWeight) }}</strong><p>{{ formatCurrency(materialSummary.monthAmount) }}</p></div>
            <div class="stat-card panel slate"><span>余料在库</span><strong>{{ formatWeight(materialSummary.remainingWeight) }}</strong><p>未出库重量</p></div>
            <div class="stat-card panel slate"><span>抵押寄存</span><strong>{{ materialSummary.pledgeCount }}</strong><p>{{ formatCurrency(materialSummary.pledgeAmount) }}</p></div>
          </div>

          <div class="content-grid two-up">
            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">登记台账</p>
                  <h3>余料 / 抵押寄存</h3>
                </div>
              </div>
              <div class="form-card">
                <label class="field">
                  <span>门店</span>
                  <select v-model="materialDraft.storeId">
                    <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                  </select>
                </label>
                <label class="field">
                  <span>类型</span>
                  <select v-model="materialDraft.type">
                    <option value="pledge">抵押寄存</option>
                    <option value="leftover">余料</option>
                    <option value="recycle">回收旧料</option>
                  </select>
                </label>
                <label class="field"><span>客户</span><input v-model="materialDraft.customerName" placeholder="可留空" /></label>
                <label class="field"><span>品类</span><input v-model="materialDraft.category" /></label>
                <label class="field"><span>成色</span><input v-model="materialDraft.purity" /></label>
                <label class="field"><span>重量(g)</span><input v-model.number="materialDraft.weightGram" min="0" step="0.01" type="number" /></label>
                <label class="field"><span>金额</span><input v-model.number="materialDraft.amount" min="0" type="number" /></label>
                <label class="field"><span>剩余重量(g)</span><input v-model.number="materialDraft.remainingWeightGram" min="0" step="0.01" type="number" /></label>
                <label class="field"><span>状态</span><input v-model="materialDraft.status" placeholder="在库 / 已取回 / 已处理" /></label>
                <label class="field"><span>到期日</span><input v-model="materialDraft.dueDate" type="date" /></label>
                <label class="field store-picker-field"><span>备注</span><textarea v-model="materialDraft.remark" rows="2"></textarea></label>
              </div>
              <div class="detail-actions">
                <button class="primary-btn" type="button" @click="addMaterialEntry">登记</button>
                <button class="secondary-btn" type="button" @click="resetMaterialDraft">清空</button>
              </div>
            </article>

            <article class="panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">成色统计</p>
                  <h3>重量与金额</h3>
                </div>
              </div>
              <div class="stack-list compact-stack">
                <div v-for="item in materialPurityStats" :key="item.purity" class="list-card simple-row">
                  <div>
                    <strong>{{ item.purity }}</strong>
                    <p>{{ item.count }} 条记录 · {{ formatWeight(item.weight) }}</p>
                  </div>
                  <strong>{{ formatCurrency(item.amount) }}</strong>
                </div>
              </div>
            </article>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="materialFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>类型</span>
              <select v-model="materialFilters.type">
                <option value="">全部</option>
                <option value="recycle">回收旧料</option>
                <option value="leftover">余料</option>
                <option value="pledge">抵押寄存</option>
              </select>
            </label>
            <label class="field compact-field"><span>开始日期</span><input v-model="materialFilters.dateFrom" type="date" /></label>
            <label class="field compact-field"><span>结束日期</span><input v-model="materialFilters.dateTo" type="date" /></label>
            <label class="field compact-field grow-field"><span>关键字</span><input v-model="materialFilters.keyword" placeholder="单号 / 客户 / 品类 / 成色" /></label>
            <button class="text-btn" type="button" @click="resetFilters('materials')">重置</button>
          </div>

          <div class="table-card">
            <table class="data-table">
              <thead>
                <tr>
                  <th class="select-cell">
                    <input :checked="allVisibleMaterialsSelected" type="checkbox" @change="toggleAllVisibleMaterials" />
                  </th>
                  <th>单号</th>
                  <th>类型</th>
                  <th>门店/客户</th>
                  <th>品类成色</th>
                  <th>重量</th>
                  <th>金额</th>
                  <th>余料</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="item in filteredMaterialRows" :key="item.id">
                  <td class="select-cell">
                    <input
                      v-if="!item.id.startsWith('recycle-')"
                      :checked="selectedMaterialIds.includes(item.id)"
                      type="checkbox"
                      @change="toggleMaterialSelection(item.id)"
                    />
                  </td>
                  <td><strong>{{ item.orderNo }}</strong><span>{{ formatDateLabel(item.createdAt) }}</span></td>
                  <td>{{ materialTypeLabel(item.type) }}</td>
                  <td>{{ item.storeName }}<br /><span>{{ item.customerName }}</span></td>
                  <td>{{ item.category }}<br /><span>{{ item.purity }}</span></td>
                  <td>{{ formatWeight(item.weightGram) }}</td>
                  <td>{{ formatCurrency(item.amount) }}</td>
                  <td>{{ formatWeight(item.remainingWeightGram) }}</td>
                  <td><span class="badge" :class="statusTone(item.status)">{{ item.status }}</span></td>
                  <td>
                    <button v-if="!item.id.startsWith('recycle-') && item.status !== '已出库'" class="text-btn" type="button" @click="removeMaterialEntry(item.id)">出库</button>
                    <button v-if="!item.id.startsWith('recycle-') && item.status === '已出库'" class="text-btn danger-text" type="button" @click="deleteMaterialEntry(item.id)">删除</button>
                    <span v-else>来自回收单</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <!-- 预约管理 -->
        <section v-if="activePage === 'appointments'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">预约管理</p>
              <h3>顾客预约</h3>
              <p class="header-copy">查看顾客预约记录、确认到店、取消预约和配置预约规则。</p>
            </div>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="appointmentFilters.status">
                <option value="">全部</option>
                <option value="PENDING">待确认</option>
                <option value="CONFIRMED">已确认</option>
                <option value="ARRIVED">已到店</option>
                <option value="COMPLETED">已完成</option>
                <option value="CANCELLED">已取消</option>
                <option value="NO_SHOW">未到店</option>
                <option value="TERMINATED">已终止</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>门店</span>
              <select v-model="appointmentFilters.storeId">
                <option value="">全部门店</option>
                <option v-for="s in stores" :key="s.id" :value="s.id">{{ s.name }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>服务类型</span>
              <select v-model="appointmentFilters.serviceType">
                <option value="">全部</option>
                <option value="OLD_FOR_NEW">以旧换新</option>
                <option value="REPAIR">维修保养</option>
                <option value="CONSULT">咨询鉴定</option>
                <option value="RECYCLE">黄金回收</option>
              </select>
            </label>
            <label class="field compact-field grow-field">
              <span>手机号</span>
              <input v-model="appointmentFilters.phone" placeholder="顾客手机号" />
            </label>
            <button class="secondary-btn" type="button" @click="loadPageData('appointments')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('appointments')">重置</button>
          </div>

          <div class="content-grid sidebar-layout">
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>预约号</th>
                    <th>顾客</th>
                    <th>门店</th>
                    <th>服务</th>
                    <th>时间</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr
                    v-for="item in appointmentItems"
                    :key="item.id"
                    :class="{ selected: selectedAppointment?.id === item.id }"
                    @click="selectAppointment(item)"
                  >
                    <td><strong>{{ item.appointmentNo }}</strong></td>
                    <td>{{ item.customerName }}<span>{{ item.customerPhone }}</span></td>
                    <td>{{ item.storeName }}</td>
                    <td>{{ item.serviceTypeText || serviceTypeLabel(item.serviceType) }}</td>
                    <td>{{ item.appointmentDate }} {{ item.appointmentTime }}</td>
                    <td><span class="badge" :class="apptStatusTone(item.status)">{{ apptStatusLabel(item.status) }}</span></td>
                  </tr>
                  <tr v-if="appointmentItems.length === 0">
                    <td colspan="6" class="empty-row">暂无预约记录</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <aside v-if="selectedAppointmentDetail" class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">预约详情</p>
                  <h3>{{ selectedAppointmentDetail.appointmentNo }}</h3>
                  <span class="badge" :class="apptStatusTone(selectedAppointmentDetail.status)">{{ apptStatusLabel(selectedAppointmentDetail.status) }}</span>
                </div>
              </div>

              <div class="info-list">
                <div class="info-row"><span class="info-label">顾客</span><span>{{ selectedAppointmentDetail.customerName }}</span></div>
                <div class="info-row"><span class="info-label">手机号</span><span>{{ selectedAppointmentDetail.customerPhone }}</span></div>
                <div class="info-row"><span class="info-label">门店</span><span>{{ selectedAppointmentDetail.storeName }}</span></div>
                <div class="info-row"><span class="info-label">服务类型</span><span>{{ selectedAppointmentDetail.serviceTypeText || serviceTypeLabel(selectedAppointmentDetail.serviceType) }}</span></div>
                <div class="info-row"><span class="info-label">预约时间</span><span>{{ selectedAppointmentDetail.appointmentDate }} {{ selectedAppointmentDetail.appointmentTime }}</span></div>
                <div class="info-row" v-if="selectedAppointmentDetail.remark"><span class="info-label">顾客备注</span><span>{{ selectedAppointmentDetail.remark }}</span></div>
                <div class="info-row" v-if="selectedAppointmentDetail.confirmedBy"><span class="info-label">确认人</span><span>{{ selectedAppointmentDetail.confirmedBy }}</span></div>
                <div class="info-row" v-if="selectedAppointmentDetail.cancelReason"><span class="info-label">取消原因</span><span>{{ selectedAppointmentDetail.cancelReason }}</span></div>
              </div>

              <div class="form-card" style="margin-top: 16px;">
                <label class="field">
                  <span>内部备注（顾客不可见）</span>
                  <textarea v-model="staffNoteDraft" rows="3" placeholder="记录内部沟通备注"></textarea>
                </label>
                <button class="secondary-btn" type="button" :disabled="actionPending" @click="saveStaffNote">保存备注</button>
              </div>

              <div class="detail-actions" style="margin-top: 16px;">
                <button v-if="selectedAppointmentDetail.status === 'PENDING'" class="primary-btn" type="button" :disabled="actionPending" @click="confirmAppointment(selectedAppointmentDetail.id)">确认预约</button>
                <button v-if="selectedAppointmentDetail.status === 'CONFIRMED'" class="primary-btn" type="button" :disabled="actionPending" @click="arriveAppointment(selectedAppointmentDetail.id)">标记到店</button>
                <button v-if="selectedAppointmentDetail.status === 'ARRIVED'" class="primary-btn" type="button" :disabled="actionPending" @click="completeAppointment(selectedAppointmentDetail.id)">完成服务</button>
                <button v-if="selectedAppointmentDetail.status === 'ARRIVED'" class="secondary-btn" type="button" :disabled="actionPending" @click="noShowAppointment(selectedAppointmentDetail.id)">标记未到</button>
                <button v-if="['PENDING', 'CONFIRMED'].includes(selectedAppointmentDetail.status)" class="text-btn danger-text" type="button" :disabled="actionPending" @click="openCancelDialog">取消预约</button>
              </div>
            </aside>
          </div>

          <!-- 预约规则配置 -->
          <div class="panel" style="margin-top: 24px;">
            <div class="panel-head">
              <div>
                <p class="eyebrow">预约规则</p>
                <h3>规则配置</h3>
                <p class="header-copy">配置可预约天数、时段粒度、营业时间和取消提前量。</p>
              </div>
              <div v-if="!rulesEditing && appointmentRules">
                <button class="secondary-btn" type="button" @click="startEditRules">编辑规则</button>
              </div>
              <div v-if="rulesEditing" style="display: flex; gap: 8px;">
                <button class="primary-btn" type="button" :disabled="actionPending" @click="saveAppointmentRulesAction">保存</button>
                <button class="text-btn" type="button" @click="cancelEditRules">取消</button>
              </div>
            </div>

            <div v-if="appointmentRulesDraft" class="form-card" style="grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));">
              <label class="field">
                <span>可预约天数</span>
                <input v-if="rulesEditing" v-model.number="appointmentRulesDraft.bookableDays" type="number" min="1" max="30" />
                <span v-else class="readonly-value">{{ appointmentRulesDraft.bookableDays }} 天</span>
              </label>
              <label class="field">
                <span>时段粒度（分钟）</span>
                <select v-if="rulesEditing" v-model.number="appointmentRulesDraft.slotMinutes">
                  <option :value="15">15</option>
                  <option :value="30">30</option>
                  <option :value="60">60</option>
                </select>
                <span v-else class="readonly-value">{{ appointmentRulesDraft.slotMinutes }} 分钟</span>
              </label>
              <label class="field">
                <span>营业开始</span>
                <input v-if="rulesEditing" v-model="appointmentRulesDraft.openTime" type="time" />
                <span v-else class="readonly-value">{{ appointmentRulesDraft.openTime }}</span>
              </label>
              <label class="field">
                <span>营业结束</span>
                <input v-if="rulesEditing" v-model="appointmentRulesDraft.closeTime" type="time" />
                <span v-else class="readonly-value">{{ appointmentRulesDraft.closeTime }}</span>
              </label>
              <label class="field">
                <span>当天提前量（分钟）</span>
                <input v-if="rulesEditing" v-model.number="appointmentRulesDraft.sameDayLeadMinutes" type="number" min="0" />
                <span v-else class="readonly-value">{{ appointmentRulesDraft.sameDayLeadMinutes }} 分钟</span>
              </label>
              <label class="field">
                <span>取消提前量（分钟）</span>
                <input v-if="rulesEditing" v-model.number="appointmentRulesDraft.cancelLeadMinutes" type="number" min="0" />
                <span v-else class="readonly-value">{{ appointmentRulesDraft.cancelLeadMinutes }} 分钟</span>
              </label>
              <label class="field">
                <span>单时段容量</span>
                <input v-if="rulesEditing" v-model.number="appointmentRulesDraft.slotCapacity" type="number" min="1" max="20" />
                <span v-else class="readonly-value">{{ appointmentRulesDraft.slotCapacity }} 单</span>
              </label>
            </div>
            <div v-if="appointmentRulesDraft && rulesEditing" class="form-card" style="margin-top: 12px;">
              <label class="field">
                <span>可预约服务类型</span>
              </label>
              <div style="display: flex; gap: 16px; flex-wrap: wrap; padding: 8px 0;">
                <label v-for="st in [{v:'OLD_FOR_NEW',l:'以旧换新'},{v:'REPAIR',l:'维修保养'},{v:'CONSULT',l:'咨询鉴定'},{v:'RECYCLE',l:'黄金回收'}]" :key="st.v" style="display: flex; align-items: center; gap: 6px;">
                  <input type="checkbox" :value="st.v" :checked="appointmentRulesDraft.bookableServiceTypes.includes(st.v)" @change="toggleServiceType(st.v)" />
                  <span>{{ st.l }}</span>
                </label>
              </div>
            </div>
          </div>
        </section>

        <!-- 取消预约弹窗 -->
        <div v-if="cancelDialogOpen" class="modal-overlay" @click.self="cancelDialogOpen = false">
          <div class="modal-card panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">取消预约</p>
                <h3>请填写取消原因</h3>
              </div>
            </div>
            <div class="form-card">
              <label class="field">
                <span>取消原因</span>
                <textarea v-model="cancelReasonDraft" rows="3" placeholder="后台取消不受提前量限制，但需填写原因"></textarea>
              </label>
            </div>
            <div class="detail-actions">
              <button class="primary-btn" type="button" :disabled="actionPending || !cancelReasonDraft.trim()" @click="doCancelAppointment">确认取消</button>
              <button class="text-btn" type="button" @click="cancelDialogOpen = false">关闭</button>
            </div>
          </div>
        </div>

        <section v-if="activePage === 'customer-home'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">顾客端配置</p>
              <h3>首页内容管理</h3>
              <p class="header-copy">配置顾客端小程序首页的品牌文案、服务介绍和轮播 Banner。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending" @click="saveHomeConfigAction">保存文案配置</button>
          </div>

          <!-- 品牌文案 -->
          <div class="panel" v-if="homeConfigDraft">
            <div class="panel-head">
              <div>
                <p class="eyebrow">品牌与服务文案</p>
                <h3>首页文案</h3>
                <p class="header-copy">这些文案会展示在顾客端首页，修改后立即生效。</p>
              </div>
            </div>
            <div class="form-card" style="grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));">
              <label class="field"><span>品牌名称（必填）</span><input v-model="homeConfigDraft.brandName" placeholder="金匠馆" /></label>
              <label class="field"><span>品牌标语一</span><input v-model="homeConfigDraft.brandSlogan1" placeholder="专业黄金服务" /></label>
              <label class="field"><span>品牌标语二</span><input v-model="homeConfigDraft.brandSlogan2" placeholder="值得信赖" /></label>
              <label class="field"><span>服务文案</span><input v-model="homeConfigDraft.serviceCopy" placeholder="以旧换新 · 维修保养 · 咨询鉴定" /></label>
              <label class="field"><span>款式入口文案</span><input v-model="homeConfigDraft.entryStyleText" placeholder="查看款式" /></label>
              <label class="field"><span>工费入口文案</span><input v-model="homeConfigDraft.entryFeeText" placeholder="工费说明" /></label>
              <label class="field"><span>客服电话</span><input v-model="homeConfigDraft.servicePhone" placeholder="400-000-0000" /></label>
              <label class="field"><span>预约须知</span><input v-model="homeConfigDraft.appointmentNotes" placeholder="预约后请提前 10 分钟到店" /></label>
              <label class="field"><span>定位权限说明</span><input v-model="homeConfigDraft.locationPermissionNote" placeholder="用于推荐附近门店" /></label>
              <label class="field" style="grid-column: 1 / -1;">
                <span>服务介绍（多行文本，V1 不支持富文本）</span>
                <textarea v-model="homeConfigDraft.serviceIntro" rows="4" placeholder="门店提供的服务介绍"></textarea>
              </label>
            </div>
          </div>

          <!-- Banner 管理 -->
          <div class="panel" style="margin-top: 24px;">
            <div class="panel-head">
              <div>
                <p class="eyebrow">首页 Banner</p>
                <h3>轮播图管理（最多 8 张启用）</h3>
                <p class="header-copy">建议尺寸 750×300，单张不超过 2MB。</p>
              </div>
              <button class="primary-btn" type="button" :disabled="actionPending" @click="openCreateBanner">新增 Banner</button>
            </div>
            <div class="table-card">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>预览</th>
                    <th>标题</th>
                    <th>跳转</th>
                    <th>排序</th>
                    <th>状态</th>
                    <th>操作</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="banner in homeBanners" :key="banner.id">
                    <td><img :src="banner.imageUrl" alt="banner" style="width: 120px; height: 48px; object-fit: cover; border-radius: 6px;" /></td>
                    <td>
                      <strong>{{ banner.title || "（无标题）" }}</strong>
                      <span v-if="banner.subtitle">{{ banner.subtitle }}</span>
                    </td>
                    <td>{{ bannerLinkLabel(banner.linkType) }}{{ banner.linkTarget ? ` · ${banner.linkTarget}` : "" }}</td>
                    <td>{{ banner.sortOrder }}</td>
                    <td><span class="badge" :class="banner.enabled ? 'emerald' : 'slate'">{{ banner.enabled ? "启用" : "停用" }}</span></td>
                    <td>
                      <div style="display: flex; gap: 8px;">
                        <button class="text-btn" type="button" @click="openEditBanner(banner)">编辑</button>
                        <button class="text-btn" type="button" @click="toggleBannerEnabled(banner)">{{ banner.enabled ? "停用" : "启用" }}</button>
                        <button class="text-btn danger-text" type="button" @click="removeBanner(banner)">删除</button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="homeBanners.length === 0">
                    <td colspan="6" class="empty-row">暂无 Banner，点击右上角新增</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- 回收服务介绍 -->
          <div class="panel" style="margin-top: 24px;">
            <div class="panel-head">
              <div>
                <p class="eyebrow">顾客端配置</p>
                <h3>回收服务介绍</h3>
                <p class="header-copy">配置顾客端「回收介绍」页的标题、简介、配图、流程步骤、服务项和须知。</p>
              </div>
              <button class="primary-btn" type="button" :disabled="actionPending" @click="saveRecycleInfoAction">保存回收介绍</button>
            </div>
            <div v-if="recycleInfoDraft" class="form-card" style="grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));">
              <label class="field"><span>回收服务标题（必填）</span><input v-model="recycleInfoDraft.title" placeholder="黄金回收服务" /></label>
              <label class="field" style="grid-column: 1 / -1;"><span>简介</span><textarea v-model="recycleInfoDraft.intro" rows="3" placeholder="页首简介文字"></textarea></label>
              <label class="field" style="grid-column: 1 / -1;">
                <span>顶部配图</span>
                <div style="display: flex; align-items: center; gap: 12px; flex-wrap: wrap;">
                  <img v-if="recycleInfoDraft.imageUrl" :src="recycleInfoDraft.imageUrl" alt="回收介绍配图" style="max-width: 240px; max-height: 90px; object-fit: cover; border-radius: 8px; display: block;" />
                  <button class="secondary-btn" type="button" :disabled="uploadPending" @click="pickCustomerImage('service_intro', 5, (url) => { if (recycleInfoDraft) recycleInfoDraft.imageUrl = url; })">
                    {{ uploadPending ? "上传中…" : (recycleInfoDraft.imageUrl ? "更换配图" : "上传配图") }}
                  </button>
                  <button v-if="recycleInfoDraft.imageUrl" class="text-btn danger-text" type="button" @click="recycleInfoDraft.imageUrl = ''">移除</button>
                </div>
              </label>

              <!-- 流程步骤 -->
              <div class="sub-block" style="grid-column: 1 / -1;">
                <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;">
                  <strong>回收流程步骤（按顺序展示）</strong>
                  <button class="secondary-btn" type="button" @click="addRecycleProcessStep">添加步骤</button>
                </div>
                <div v-for="(step, idx) in recycleInfoDraft.process" :key="idx" class="sub-item">
                  <span class="step-no">{{ idx + 1 }}</span>
                  <input v-model="step.title" placeholder="步骤标题（如：到店咨询）" style="flex: 1.2;" />
                  <input v-model="step.desc" placeholder="步骤说明" style="flex: 2;" />
                  <button class="text-btn" type="button" :disabled="idx === 0" @click="moveRecycleProcessStep(idx, -1)">↑</button>
                  <button class="text-btn" type="button" :disabled="idx === recycleInfoDraft.process.length - 1" @click="moveRecycleProcessStep(idx, 1)">↓</button>
                  <button class="text-btn danger-text" type="button" @click="removeRecycleProcessStep(idx)">删</button>
                </div>
                <div v-if="recycleInfoDraft.process.length === 0" class="empty-row">暂无流程步骤，点击「添加步骤」</div>
              </div>

              <!-- 服务项 -->
              <div class="sub-block" style="grid-column: 1 / -1;">
                <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;">
                  <strong>服务项目（图标 + 标题 + 说明）</strong>
                  <button class="secondary-btn" type="button" @click="addRecycleServiceItem">添加服务项</button>
                </div>
                <div v-for="(svc, idx) in recycleInfoDraft.services" :key="idx" class="sub-item">
                  <select v-model="svc.icon" style="width: 130px;">
                    <option value="">无图标</option>
                    <option value="recycle">♻ 回收</option>
                    <option value="repair">🔧 维修</option>
                    <option value="consult">💎 咨询</option>
                    <option value="custom">✨ 定制</option>
                  </select>
                  <input v-model="svc.title" placeholder="服务名称（如：旧金换新）" style="flex: 1.2;" />
                  <input v-model="svc.desc" placeholder="服务说明" style="flex: 2;" />
                  <button class="text-btn" type="button" :disabled="idx === 0" @click="moveRecycleServiceItem(idx, -1)">↑</button>
                  <button class="text-btn" type="button" :disabled="idx === recycleInfoDraft.services.length - 1" @click="moveRecycleServiceItem(idx, 1)">↓</button>
                  <button class="text-btn danger-text" type="button" @click="removeRecycleServiceItem(idx)">删</button>
                </div>
                <div v-if="recycleInfoDraft.services.length === 0" class="empty-row">暂无服务项目，点击「添加服务项」</div>
              </div>

              <!-- 须知 -->
              <div class="sub-block" style="grid-column: 1 / -1;">
                <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 8px;">
                  <strong>回收须知（逐条展示）</strong>
                  <button class="secondary-btn" type="button" @click="addRecycleNotice">添加须知</button>
                </div>
                <div v-for="(notice, idx) in recycleInfoDraft.notices" :key="idx" class="sub-item">
                  <span class="step-no">{{ idx + 1 }}</span>
                  <input v-model="recycleInfoDraft.notices[idx]" placeholder="须知内容" style="flex: 3;" />
                  <button class="text-btn" type="button" :disabled="idx === 0" @click="moveRecycleNotice(idx, -1)">↑</button>
                  <button class="text-btn" type="button" :disabled="idx === recycleInfoDraft.notices.length - 1" @click="moveRecycleNotice(idx, 1)">↓</button>
                  <button class="text-btn danger-text" type="button" @click="removeRecycleNotice(idx)">删</button>
                </div>
                <div v-if="recycleInfoDraft.notices.length === 0" class="empty-row">暂无须知，点击「添加须知」</div>
              </div>
            </div>
          </div>
        </section>

        <section v-if="activePage === 'customer-styles'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">款式工费管理</p>
              <h3>顾客端款式</h3>
              <p class="header-copy">维护顾客端展示的款式、图片、工费说明和上下架状态。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending" @click="openCreateStyle">新增款式</button>
          </div>

          <div class="panel toolbar">
            <label class="field compact-field">
              <span>分类</span>
              <select v-model="styleFilters.category">
                <option value="">全部分类</option>
                <option v-for="cat in styleCategories" :key="cat" :value="cat">{{ cat }}</option>
              </select>
            </label>
            <label class="field compact-field">
              <span>状态</span>
              <select v-model="styleFilters.status">
                <option value="">全部</option>
                <option value="active">上架</option>
                <option value="inactive">下架</option>
              </select>
            </label>
            <label class="field compact-field grow-field">
              <span>关键词</span>
              <input v-model="styleFilters.keyword" placeholder="款式名称 / 纯度" @keyup.enter="styleFilters.page = 1; loadPageData('customer-styles')" />
            </label>
            <button class="secondary-btn" type="button" @click="styleFilters.page = 1; loadPageData('customer-styles')">查询</button>
            <button class="text-btn" type="button" @click="resetFilters('customer-styles')">重置</button>
          </div>

          <div class="table-card">
            <table class="data-table">
              <thead>
                <tr>
                  <th>主图</th>
                  <th>款式</th>
                  <th>分类 / 纯度</th>
                  <th>工费参考</th>
                  <th>排序</th>
                  <th>状态</th>
                  <th>操作</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="record in styleItems" :key="record.id">
                  <td><img :src="record.imageUrl" alt="款式主图" style="width: 56px; height: 56px; object-fit: cover; border-radius: 8px;" /></td>
                  <td>
                    <strong>{{ record.name }}</strong>
                    <span v-if="record.isHot">🔥 热门</span>
                  </td>
                  <td>{{ record.category || "—" }} / {{ record.purity || "—" }}</td>
                  <td>{{ record.laborFeeRef || "—" }}</td>
                  <td>{{ record.sortOrder }}</td>
                  <td><span class="badge" :class="record.status === 'active' ? 'emerald' : 'slate'">{{ styleStatusLabel(record.status) }}</span></td>
                  <td>
                    <div style="display: flex; gap: 8px;">
                      <button class="text-btn" type="button" @click="openEditStyle(record)">编辑</button>
                      <button class="text-btn danger-text" type="button" @click="removeStyle(record)">删除</button>
                    </div>
                  </td>
                </tr>
                <tr v-if="styleItems.length === 0">
                  <td colspan="7" class="empty-row">暂无款式记录</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="styleTotal > (styleFilters.pageSize || 20)" class="panel toolbar" style="justify-content: flex-end;">
            <button class="secondary-btn" type="button" :disabled="(styleFilters.page || 1) <= 1" @click="styleFilters.page = (styleFilters.page || 1) - 1; loadPageData('customer-styles')">上一页</button>
            <span style="align-self: center; font-size: 13px;">第 {{ styleFilters.page }} 页 / 共 {{ Math.ceil(styleTotal / (styleFilters.pageSize || 20)) }} 页（{{ styleTotal }} 条）</span>
            <button class="secondary-btn" type="button" :disabled="(styleFilters.page || 1) >= Math.ceil(styleTotal / (styleFilters.pageSize || 20))" @click="styleFilters.page = (styleFilters.page || 1) + 1; loadPageData('customer-styles')">下一页</button>
          </div>
        </section>

        <!-- Banner 编辑弹窗 -->
        <div v-if="bannerDialogOpen && bannerDraft" class="modal-overlay" @click.self="bannerDialogOpen = false">
          <div class="modal-card panel">
            <div class="panel-head">
              <div>
                <p class="eyebrow">首页 Banner</p>
                <h3>{{ bannerEditingId ? "编辑 Banner" : "新增 Banner" }}</h3>
              </div>
            </div>
            <div class="form-card">
              <label class="field">
                <span>Banner 图片（750×300，≤2MB，必填）</span>
                <div v-if="bannerDraft.imageUrl" style="margin-bottom: 8px;">
                  <img :src="bannerDraft.imageUrl" alt="banner" style="max-width: 260px; border-radius: 8px; display: block;" />
                </div>
                <button class="secondary-btn" type="button" :disabled="uploadPending" @click="pickCustomerImage('banner', 2, (url) => { if (bannerDraft) bannerDraft.imageUrl = url; })">
                  {{ uploadPending ? "上传中…" : (bannerDraft.imageUrl ? "更换图片" : "上传图片") }}
                </button>
              </label>
              <label class="field"><span>标题</span><input v-model="bannerDraft.title" placeholder="活动标题" /></label>
              <label class="field"><span>副标题</span><input v-model="bannerDraft.subtitle" placeholder="可选" /></label>
              <label class="field">
                <span>跳转类型</span>
                <select v-model="bannerDraft.linkType">
                  <option value="none">不跳转</option>
                  <option value="styles">款式列表</option>
                  <option value="stores">门店列表</option>
                  <option value="booking">预约页</option>
                  <option value="custom">自定义页面</option>
                </select>
              </label>
              <label class="field" v-if="bannerDraft.linkType === 'custom'"><span>跳转目标（页面路径）</span><input v-model="bannerDraft.linkTarget" placeholder="/pkg-customer/pages/xxx" /></label>
              <label class="field"><span>排序（数字越小越靠前）</span><input v-model.number="bannerDraft.sortOrder" type="number" /></label>
              <label class="field">
                <span>启用状态</span>
                <select v-model="bannerDraft.enabled">
                  <option :value="true">启用</option>
                  <option :value="false">停用</option>
                </select>
              </label>
            </div>
            <div class="detail-actions">
              <button class="primary-btn" type="button" :disabled="actionPending || !bannerDraft.imageUrl.trim()" @click="saveBannerAction">保存</button>
              <button class="text-btn" type="button" @click="bannerDialogOpen = false">取消</button>
            </div>
          </div>
        </div>

        <!-- 款式编辑弹窗 -->
        <div v-if="styleDialogOpen && styleDraft" class="modal-overlay" @click.self="styleDialogOpen = false">
          <div class="modal-card panel" style="max-width: 720px; max-height: 86vh; overflow-y: auto;">
            <div class="panel-head">
              <div>
                <p class="eyebrow">款式工费</p>
                <h3>{{ styleEditingId ? "编辑款式" : "新增款式" }}</h3>
              </div>
            </div>
            <div class="form-card" style="grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));">
              <label class="field"><span>款式名称（必填）</span><input v-model="styleDraft.name" placeholder="如：足金传承手镯" /></label>
              <label class="field"><span>分类</span><input v-model="styleDraft.category" placeholder="如：手镯" list="style-category-options" /></label>
              <datalist id="style-category-options">
                <option v-for="cat in styleCategories" :key="cat" :value="cat" />
              </datalist>
              <label class="field"><span>纯度</span><input v-model="styleDraft.purity" placeholder="如：足金999" /></label>
              <label class="field"><span>工费参考</span><input v-model="styleDraft.laborFeeRef" placeholder="如：35元/克 起" /></label>
              <label class="field"><span>零售价（元）</span><input v-model.number="styleDraft.retailPrice" type="number" min="0" /></label>
              <label class="field"><span>克重（克）</span><input v-model.number="styleDraft.gramWeight" type="number" min="0" step="0.01" /></label>
              <label class="field"><span>推荐门店规则</span>
                <select v-model="styleDraft.recommendedStoreRule">
                  <option value="nearest">离顾客最近</option>
                  <option value="product_stores">仅指定门店</option>
                  <option value="all">全部门店</option>
                </select>
              </label>
              <label class="field"><span>排序（数字越小越靠前）</span><input v-model.number="styleDraft.sortOrder" type="number" /></label>
              <label class="field"><span>上架状态</span>
                <select v-model="styleDraft.status">
                  <option value="active">上架</option>
                  <option value="inactive">下架</option>
                </select>
              </label>
              <label class="field" style="grid-column: 1 / -1;"><span>款式说明</span><textarea v-model="styleDraft.description" rows="3" placeholder="展示在款式详情页"></textarea></label>
              <label class="field" style="grid-column: 1 / -1;"><span>工费说明</span><textarea v-model="styleDraft.laborFeeNote" rows="2" placeholder="如：工费按工艺复杂度浮动，以门店报价为准"></textarea></label>
            </div>

            <div class="form-card" style="margin-top: 12px;">
              <label class="field">
                <span>款式图片（第一张为主图，必填）</span>
              </label>
              <div style="display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 10px;">
                <div v-for="(img, index) in styleDraft.images" :key="img + index" style="position: relative;">
                  <img :src="img" alt="款式图" style="width: 96px; height: 96px; object-fit: cover; border-radius: 8px; display: block;" />
                  <div style="display: flex; gap: 4px; justify-content: center; margin-top: 4px;">
                    <button class="text-btn" type="button" :disabled="index === 0" @click="moveStyleImage(index, -1)">←</button>
                    <button class="text-btn" type="button" :disabled="index === (styleDraft.images?.length || 0) - 1" @click="moveStyleImage(index, 1)">→</button>
                    <button class="text-btn danger-text" type="button" @click="removeStyleImage(index)">删</button>
                  </div>
                  <span v-if="index === 0" style="position: absolute; top: 4px; left: 4px; font-size: 11px; background: rgba(0,0,0,0.6); color: #fff; padding: 1px 6px; border-radius: 4px;">主图</span>
                </div>
              </div>
              <button class="secondary-btn" type="button" :disabled="uploadPending" @click="pickCustomerImage('style_main', 5, (url) => { if (styleDraft) { styleDraft.images = [...(styleDraft.images || []), url]; if (!styleDraft.imageUrl) styleDraft.imageUrl = url; } })">
                {{ uploadPending ? "上传中…" : "添加图片" }}
              </button>
            </div>

            <div class="form-card" style="margin-top: 12px;">
              <label class="field">
                <span>款式详情图（展示在款式详情页，可多张）</span>
              </label>
              <div style="display: flex; gap: 12px; flex-wrap: wrap; margin-bottom: 10px;">
                <div v-for="(img, index) in styleDraft.detailImages" :key="img + index" style="position: relative;">
                  <img :src="img" alt="款式详情图" style="width: 96px; height: 96px; object-fit: cover; border-radius: 8px; display: block;" />
                  <div style="display: flex; gap: 4px; justify-content: center; margin-top: 4px;">
                    <button class="text-btn" type="button" :disabled="index === 0" @click="moveStyleDetailImage(index, -1)">←</button>
                    <button class="text-btn" type="button" :disabled="index === (styleDraft.detailImages?.length || 0) - 1" @click="moveStyleDetailImage(index, 1)">→</button>
                    <button class="text-btn danger-text" type="button" @click="removeStyleDetailImage(index)">删</button>
                  </div>
                  <span style="position: absolute; top: 4px; left: 4px; font-size: 11px; background: rgba(0,0,0,0.6); color: #fff; padding: 1px 6px; border-radius: 4px;">详情{{ index + 1 }}</span>
                </div>
              </div>
              <button class="secondary-btn" type="button" :disabled="uploadPending" @click="pickCustomerImage('style_detail', 5, (url) => { if (styleDraft) { styleDraft.detailImages = [...(styleDraft.detailImages || []), url]; } })">
                {{ uploadPending ? "上传中…" : "添加详情图" }}
              </button>
            </div>

            <div class="form-card" style="margin-top: 12px;">
              <label class="field"><span>适用服务类型</span></label>
              <div style="display: flex; gap: 16px; flex-wrap: wrap; padding: 8px 0;">
                <label v-for="st in [{v:'OLD_FOR_NEW',l:'以旧换新'},{v:'REPAIR',l:'维修保养'},{v:'CONSULT',l:'咨询鉴定'},{v:'RECYCLE',l:'黄金回收'}]" :key="st.v" style="display: flex; align-items: center; gap: 6px;">
                  <input type="checkbox" :value="st.v" :checked="(styleDraft.applicableServiceTypes || []).includes(st.v)" @change="toggleStyleServiceType(st.v)" />
                  <span>{{ st.l }}</span>
                </label>
                <label style="display: flex; align-items: center; gap: 6px;">
                  <input type="checkbox" :checked="styleDraft.isHot" @change="styleDraft.isHot = ($event.target as HTMLInputElement).checked" />
                  <span>热门款式</span>
                </label>
                <label style="display: flex; align-items: center; gap: 6px;">
                  <input type="checkbox" :checked="styleDraft.isRecommended" @change="styleDraft.isRecommended = ($event.target as HTMLInputElement).checked" />
                  <span>推荐款式</span>
                </label>
              </div>
            </div>

            <div class="detail-actions">
              <button class="primary-btn" type="button" :disabled="actionPending" @click="saveStyleAction">保存</button>
              <button class="text-btn" type="button" @click="styleDialogOpen = false">取消</button>
            </div>
          </div>
        </div>

        <section v-if="activePage === 'settings'" class="page-grid">
          <div class="panel simple-page-head">
            <div>
              <p class="eyebrow">系统设置</p>
              <h3>基础资料维护</h3>
              <p class="header-copy">维护品牌、小票、客服电话和回收拍照规则。</p>
            </div>
            <button class="primary-btn" type="button" :disabled="actionPending || !systemDraft" @click="saveSettings">保存设置</button>
          </div>

          <div class="content-grid sidebar-layout">
            <article v-if="systemDraft" class="panel">
              <div class="form-card">
                <label class="field"><span>品牌名称</span><input v-model="systemDraft.brandName" /></label>
                <label class="field"><span>客服电话</span><input v-model="systemDraft.servicePhone" /></label>
                <label class="field"><span>小票标题</span><input v-model="systemDraft.receiptTitle" /></label>
                <label class="field"><span>最少照片数</span><input v-model.number="systemDraft.minPhotoCount" min="0" type="number" /></label>
                <label class="field"><span>最多照片数</span><input v-model.number="systemDraft.maxPhotoCount" min="0" type="number" /></label>
                <label class="field"><span>打印状态</span><input v-model="systemDraft.printerStatus" /></label>
                <label class="toggle-field"><input v-model="systemDraft.requireExactThree" type="checkbox" /><span>回收单要求固定照片数量</span></label>
                <label class="toggle-field"><input v-model="systemDraft.requireIdCheck" type="checkbox" /><span>回收单要求核对客户身份</span></label>
              </div>
            </article>

            <aside class="detail-card panel">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">当前资料</p>
                  <h3>系统概况</h3>
                </div>
              </div>
              <div class="detail-grid">
                <div v-for="row in settingsRows" :key="row.label" class="list-card">
                  <span>{{ row.label }}</span>
                  <strong>{{ row.value }}</strong>
                </div>
              </div>
            </aside>
          </div>
        </section>
      </main>
    </section>

    <div v-if="recyclePreview.open && previewAsset" class="preview-modal" @click.self="closeRecyclePreview">
      <div class="preview-dialog panel">
        <div class="panel-head">
          <div>
            <p class="eyebrow">图片预览</p>
            <h3>{{ previewAsset.fileName || recyclePreview.orderNo }}</h3>
            <p class="header-copy">{{ recyclePreview.orderNo }} · 第 {{ recyclePreview.index + 1 }} / {{ recyclePreview.assets.length }} 张</p>
          </div>
          <button class="secondary-btn" type="button" @click="closeRecyclePreview">关闭</button>
        </div>
        <div class="preview-stage">
          <button class="secondary-btn" type="button" @click="showNextPreview(-1)">上一张</button>
          <img :src="previewAsset.previewUrl || previewAsset.publicUrl" :alt="previewAsset.fileName || '回收照片'" />
          <button class="secondary-btn" type="button" @click="showNextPreview(1)">下一张</button>
        </div>
      </div>
    </div>
  </div>
</template>
