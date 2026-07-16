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
  cancelRecycleOrder,
  createInventoryLedgerItem,
  createMaterialLedgerItem,
  createMemberProfile,
  createProductRecord,
  createStoreRecord,
  createUserAccount,
  deleteInventoryLedgerItem,
  deleteMaterialLedgerItem,
  deleteProductRecords,
  disableMemberProfile,
  disableProductRecord,
  disableStoreRecord,
  disableUserAccount,
  downloadProductImportTemplate,
  fetchCashierOrderDetail,
  fetchCashierOrders,
  fetchAdminConsole,
  fetchInventoryLedger,
  fetchMaterialLedger,
  fetchMemberProfiles,
  fetchOrderRows,
  fetchProductRecords,
  fetchRecycleOrderDetail,
  fetchRecycleOrders,
  fetchRoleTemplates,
  fetchStoreRecords,
  fetchUserAccounts,
  importProducts,
  loginAdmin,
  outboundMaterialLedgerItem,
  saveInventoryLedgerItem,
  saveMaterialLedgerItem,
  saveMemberProfile,
  saveProductRecord,
  saveRoleTemplate,
  saveStoreRecord,
  saveSystemProfile,
  saveUserAccount,
  voidCashierOrder,
  type AbilityCode,
  type AbilityGroup,
  type AdminOrderRow,
  type AttachmentAsset,
  type CashierOrderView,
  type ConsoleBootstrap,
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
  | "settings";
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
const inventoryActionMenuId = ref("");
const editingInventoryId = ref("");
const materialActionMenuId = ref("");
const editingMaterialId = ref("");
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
const inventoryEditDraft = reactive({
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
const materialEditDraft = reactive({
  storeId: "",
  type: "pledge" as MaterialType,
  orderNo: "",
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

function materialStatusForApi(status: string) {
  if (status === "已出库" || status === "outbound") return "outbound";
  if (!status || status === "在库" || status === "in_stock") return "in_stock";
  return status;
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

function inventoryStatusForApi(status: InventoryStatus | string) {
  if (status === "out") return "outbound";
  if (status === "normal" || status === "low" || status === "review") return "in_stock";
  return status || "in_stock";
}

function toggleInventoryActionMenu(id: string) {
  inventoryActionMenuId.value = inventoryActionMenuId.value === id ? "" : id;
}

function openInventoryEditor(item: InventoryStockItem) {
  inventoryActionMenuId.value = "";
  editingInventoryId.value = item.id;
  inventoryEditDraft.storeId = item.storeId;
  inventoryEditDraft.styleNo = item.styleNo;
  inventoryEditDraft.name = item.name;
  inventoryEditDraft.category = item.category;
  inventoryEditDraft.purity = item.purity;
  inventoryEditDraft.pieceCount = item.pieceCount;
  inventoryEditDraft.weightGram = item.weightGram;
  inventoryEditDraft.unitCost = item.unitCost;
  inventoryEditDraft.status = item.status;
}

function closeInventoryEditor() {
  editingInventoryId.value = "";
}

function saveInventoryEditor() {
  if (!sessionToken.value || !editingInventoryId.value) return;
  if (!inventoryEditDraft.styleNo.trim() || !inventoryEditDraft.name.trim()) {
    showNotice("warning", "请填写款式编号和商品名称。");
    return;
  }
  const itemId = editingInventoryId.value;
  void runAction(async () => {
    await saveInventoryLedgerItem(sessionToken.value, itemId, {
      storeId: inventoryEditDraft.storeId,
      styleNo: inventoryEditDraft.styleNo,
      name: inventoryEditDraft.name,
      category: inventoryEditDraft.category,
      purity: inventoryEditDraft.purity,
      pieceCount: inventoryEditDraft.pieceCount,
      weightGram: inventoryEditDraft.weightGram,
      costAmount: inventoryEditDraft.unitCost,
      status: inventoryStatusForApi(inventoryEditDraft.status),
      source: "admin",
    });
    editingInventoryId.value = "";
  }, "库存记录已更新。", "inventory");
}

function deleteInventoryEntry(id: string) {
  if (!sessionToken.value) return;
  inventoryActionMenuId.value = "";
  if (!window.confirm("确认删除这条库存记录吗？删除后不会在库存台账中显示。")) return;
  void runAction(async () => {
    await deleteInventoryLedgerItem(sessionToken.value, id);
  }, "库存记录已删除。", "inventory");
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
  materialActionMenuId.value = "";
  void runAction(async () => {
    await outboundMaterialLedgerItem(sessionToken.value, id, "后台旧料出库");
  }, "旧料已出库，余料统计已更新。", "materials");
}

function deleteMaterialEntry(id: string) {
  if (!sessionToken.value) return;
  materialActionMenuId.value = "";
  if (!window.confirm("确认删除这条旧料记录吗？删除后不会在台账中显示。")) return;
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

function toggleMaterialActionMenu(id: string) {
  materialActionMenuId.value = materialActionMenuId.value === id ? "" : id;
}

function openMaterialEditor(item: MaterialLedgerItem) {
  materialActionMenuId.value = "";
  editingMaterialId.value = item.id;
  materialEditDraft.storeId = item.storeId;
  materialEditDraft.type = item.type;
  materialEditDraft.orderNo = item.orderNo;
  materialEditDraft.customerName = item.customerName;
  materialEditDraft.category = item.category;
  materialEditDraft.purity = item.purity;
  materialEditDraft.weightGram = item.weightGram;
  materialEditDraft.amount = item.amount;
  materialEditDraft.remainingWeightGram = item.remainingWeightGram;
  materialEditDraft.status = item.status;
  materialEditDraft.dueDate = item.dueDate;
  materialEditDraft.remark = item.remark;
}

function closeMaterialEditor() {
  editingMaterialId.value = "";
}

function saveMaterialEditor() {
  if (!sessionToken.value || !editingMaterialId.value) return;
  if (!materialEditDraft.category.trim() || !materialEditDraft.purity.trim()) {
    showNotice("warning", "请填写旧料品类和成色。");
    return;
  }
  const itemId = editingMaterialId.value;
  void runAction(async () => {
    await saveMaterialLedgerItem(sessionToken.value, itemId, {
      storeId: materialEditDraft.storeId,
      type: materialEditDraft.type,
      orderNo: materialEditDraft.orderNo,
      customerName: materialEditDraft.customerName,
      category: materialEditDraft.category,
      purity: materialEditDraft.purity,
      weightGram: materialEditDraft.weightGram,
      amount: materialEditDraft.amount,
      remainingWeightGram: materialEditDraft.remainingWeightGram,
      status: materialStatusForApi(materialEditDraft.status),
      dueDate: materialEditDraft.dueDate,
      remark: materialEditDraft.remark,
      source: "admin",
    });
    editingMaterialId.value = "";
  }, "旧料记录已更新。", "materials");
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
  return { paid: "已完成", pending: "待复核", refunded: "已退单" }[status] || status;
}

function recycleStatusLabel(status: RecycleOrderView["status"]) {
  return { draft: "待确认", confirmed: "已确认", cancelled: "已作废" }[status] || status;
}

function statusTone(status: string): MetricTone {
  if (["已完成", "已确认", "上架中", "营业中", "正常", "库存正常", "启用中"].includes(status)) return "emerald";
  if (["待复核", "待确认", "待完善", "待开业", "库存较低", "待激活"].includes(status)) return "gold";
  if (["已取消", "已作废", "已退单", "已删除", "无库存"].includes(status)) return "danger";
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
  const reason = window.prompt("请输入退单原因");
  if (!reason || !reason.trim()) return;
  await runAction(async () => {
    await voidCashierOrder(sessionToken.value, currentCashierDetail.value!.id, reason.trim());
  }, "退单已完成，统计会扣除这笔成交。", "cashier");
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
                <option value="refunded">已退单</option>
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
                  {{ currentCashierDetail.status === "refunded" ? "已退单" : "退单" }}
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
              <div v-if="currentCashierDetail.refundReason || currentCashierDetail.voidReason" class="list-card warning-card">
                <strong>退单信息</strong>
                <p>{{ currentCashierDetail.refundReason || currentCashierDetail.voidReason }}<br />{{ currentCashierDetail.refundedBy || currentCashierDetail.voidedBy || "-" }} · {{ currentCashierDetail.refundedAt || currentCashierDetail.voidedAt || "-" }}</p>
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
                <option value="refunded">收银已退单</option>
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
                    <div v-if="item.source !== '商品目录'" class="action-menu-wrap">
                      <button class="secondary-btn small-btn" type="button" @click="toggleInventoryActionMenu(item.id)">操作</button>
                      <div v-if="inventoryActionMenuId === item.id" class="action-menu">
                        <button type="button" @click="openInventoryEditor(item)">编辑</button>
                        <button class="danger-text" type="button" @click="deleteInventoryEntry(item.id)">删除</button>
                      </div>
                    </div>
                    <span v-else>随商品维护</span>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <article v-if="editingInventoryId" class="panel editor-panel">
            <div class="section-heading">
              <div>
                <p class="eyebrow">编辑库存</p>
                <h3>{{ inventoryEditDraft.styleNo || "库存记录" }}</h3>
              </div>
              <button class="text-btn" type="button" @click="closeInventoryEditor">关闭</button>
            </div>
            <div class="form-grid">
              <label class="field">
                <span>门店</span>
                <select v-model="inventoryEditDraft.storeId">
                  <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                </select>
              </label>
              <label class="field"><span>款式编号</span><input v-model="inventoryEditDraft.styleNo" /></label>
              <label class="field"><span>商品名称</span><input v-model="inventoryEditDraft.name" /></label>
              <label class="field"><span>品类</span><input v-model="inventoryEditDraft.category" /></label>
              <label class="field"><span>成色</span><input v-model="inventoryEditDraft.purity" /></label>
              <label class="field"><span>件数</span><input v-model.number="inventoryEditDraft.pieceCount" min="0" type="number" /></label>
              <label class="field"><span>重量(g)</span><input v-model.number="inventoryEditDraft.weightGram" min="0" step="0.01" type="number" /></label>
              <label class="field"><span>成本/参考价</span><input v-model.number="inventoryEditDraft.unitCost" min="0" type="number" /></label>
              <label class="field">
                <span>库存状态</span>
                <select v-model="inventoryEditDraft.status">
                  <option value="normal">库存正常</option>
                  <option value="low">库存较低</option>
                  <option value="review">待盘点</option>
                  <option value="out">无库存</option>
                </select>
              </label>
            </div>
            <div class="actions-row">
              <button class="primary-btn" type="button" :disabled="actionPending" @click="saveInventoryEditor">保存库存</button>
            </div>
          </article>
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
                  <td class="action-cell" @click.stop>
                    <button class="secondary-btn small-btn" type="button" @click="toggleMaterialActionMenu(item.id)">操作</button>
                    <div v-if="materialActionMenuId === item.id" class="row-action-menu">
                      <template v-if="!item.id.startsWith('recycle-')">
                        <button type="button" @click="openMaterialEditor(item)">编辑</button>
                        <button v-if="item.status !== '已出库'" type="button" @click="removeMaterialEntry(item.id)">出库</button>
                        <button class="danger-text" type="button" @click="deleteMaterialEntry(item.id)">删除</button>
                      </template>
                      <span v-else>回收单生成，需到回收单处理</span>
                    </div>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>

          <div v-if="editingMaterialId" class="edit-modal" @click.self="closeMaterialEditor">
            <article class="panel edit-dialog">
              <div class="panel-head">
                <div>
                  <p class="eyebrow">编辑旧料</p>
                  <h3>{{ materialEditDraft.orderNo || "旧料记录" }}</h3>
                  <p class="header-copy">修改后会写入后端台账，后台和小程序读取同一份数据。</p>
                </div>
                <button class="secondary-btn" type="button" @click="closeMaterialEditor">关闭</button>
              </div>
              <div class="form-card edit-form-grid">
                <label class="field">
                  <span>门店</span>
                  <select v-model="materialEditDraft.storeId">
                    <option v-for="store in stores" :key="store.id" :value="store.id">{{ store.name }}</option>
                  </select>
                </label>
                <label class="field">
                  <span>类型</span>
                  <select v-model="materialEditDraft.type">
                    <option value="leftover">余料</option>
                    <option value="pledge">抵押寄存</option>
                    <option value="recycle">回收旧料</option>
                  </select>
                </label>
                <label class="field"><span>单号</span><input v-model="materialEditDraft.orderNo" /></label>
                <label class="field"><span>客户</span><input v-model="materialEditDraft.customerName" /></label>
                <label class="field"><span>品类</span><input v-model="materialEditDraft.category" /></label>
                <label class="field"><span>成色</span><input v-model="materialEditDraft.purity" /></label>
                <label class="field"><span>重量(g)</span><input v-model.number="materialEditDraft.weightGram" min="0" step="0.01" type="number" /></label>
                <label class="field"><span>金额</span><input v-model.number="materialEditDraft.amount" min="0" type="number" /></label>
                <label class="field"><span>剩余重量(g)</span><input v-model.number="materialEditDraft.remainingWeightGram" min="0" step="0.01" type="number" /></label>
                <label class="field">
                  <span>状态</span>
                  <select v-model="materialEditDraft.status">
                    <option value="在库">在库</option>
                    <option value="已出库">已出库</option>
                    <option value="已取回">已取回</option>
                    <option value="已处理">已处理</option>
                  </select>
                </label>
                <label class="field"><span>到期日</span><input v-model="materialEditDraft.dueDate" type="date" /></label>
                <label class="field store-picker-field"><span>备注</span><textarea v-model="materialEditDraft.remark" rows="3"></textarea></label>
              </div>
              <div class="form-actions">
                <button class="primary-btn" type="button" :disabled="actionPending" @click="saveMaterialEditor">保存</button>
                <button class="secondary-btn" type="button" @click="closeMaterialEditor">取消</button>
              </div>
            </article>
          </div>
        </section>

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
