/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: api.ts
 * 功能描述: 接口模块
 * 作者: 廖心慈
 * 创建日期: 2026-05-08
 */

import {
  buildLocalConsoleBootstrap,
  getLocalSessionByCredentials,
  getLocalSessionByToken,
} from "./localData";

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export type DataSource = "api" | "local";
export type AdminApiMode = "strict-api" | "local-data-enabled";
export type DataScope = "all_stores" | "assigned_store" | "self";
export type RoleKey = "boss" | "shop_manager";
export type StoreStatus = "active" | "pending" | "disabled";
export type UserStatus = "enabled" | "disabled" | "invited";
export type MetricTone = "gold" | "emerald" | "slate" | "danger";
export type TodoLevel = "high" | "medium" | "low";
export type PageId =
  | "dashboard"
  | "stores"
  | "users"
  | "roles"
  | "members"
  | "products"
  | "orders"
  | "recycle"
  | "cashier"
  | "settings"
  | "system-config"
  | "template-init"
  | "audit";

export type AbilityCode =
  | "auth.login"
  | "dashboard.view"
  | "store.view"
  | "store.manage"
  | "user.view"
  | "user.manage"
  | "role.view"
  | "role.manage"
  | "product.view"
  | "product.manage"
  | "order.view"
  | "recycle.view"
  | "system.config.view"
  | "system.config.manage"
  | "template.init"
  | "audit.view";

export interface SessionUser {
  id: string;
  name: string;
  account: string;
  roleKey: RoleKey;
  roleName: string;
  dataScope: DataScope;
  storeIds: string[];
  abilities: AbilityCode[];
  lastLoginAt: string;
}

export interface LoginResult {
  token: string;
  user: SessionUser;
  landingPage: PageId;
}

export interface DashboardMetric {
  key: string;
  label: string;
  value: string;
  delta: string;
  tone: MetricTone;
}

export interface DashboardTodo {
  id: string;
  title: string;
  description: string;
  level: TodoLevel;
  page: PageId;
}

export interface DashboardShortcut {
  id: string;
  label: string;
  hint: string;
  page: PageId;
  ability: AbilityCode;
}

export interface DashboardSummary {
  title: string;
  subtitle: string;
  metrics: DashboardMetric[];
  todos: DashboardTodo[];
  shortcuts: DashboardShortcut[];
  notices: string[];
}

export interface StoreRecord {
  id: string;
  code: string;
  name: string;
  managerName: string;
  city: string;
  address: string;
  contactPhone: string;
  businessHours: string;
  status: StoreStatus;
  cashierDevices: number;
  pendingTasks: number;
  todayAmount: number;
  todayOrders: number;
  lastSettlementAt: string;
  tags: string[];
}

export interface UserAccount {
  id: string;
  name: string;
  account: string;
  phone: string;
  roleKey: RoleKey;
  roleName: string;
  dataScope: DataScope;
  storeIds: string[];
  storeNames: string[];
  status: UserStatus;
  lastLoginAt: string;
  abilities: AbilityCode[];
  password?: string;
}

export interface AbilityOption {
  code: AbilityCode;
  label: string;
  description: string;
}

export interface AbilityGroup {
  key: string;
  label: string;
  description: string;
  items: AbilityOption[];
}

export interface RoleTemplate {
  id: number;
  key: RoleKey;
  name: string;
  description: string;
  dataScope: DataScope;
  memberCount: number;
  locked: boolean;
  abilities: AbilityCode[];
}

export interface ProductRecord {
  id: string;
  name: string;
  sku: string;
  category: string;
  categoryTab?: string;
  imageUrl?: string;
  price: number;
  gramWeight: number;
  status: "active" | "draft" | "disabled";
  inventory?: number;
  stockStatus?: "normal" | "low" | "review" | "disabled" | "out" | string;
  storeIds: string[];
  storeNames: string[];
  tags: string[];
}

export interface ProductImportFailure {
  row: number;
  reason: string;
}

export interface ProductImportResult {
  importId: string;
  storeId: string;
  storeName: string;
  successCount: number;
  failureCount: number;
  failures: ProductImportFailure[];
  importedAt: string;
}

export interface ProductImportLog extends ProductImportResult {
  id: string;
  fileName: string;
  importedBy: string;
}

export interface InventoryLedgerItem {
  id: string;
  orgId?: string;
  storeId: string;
  storeName: string;
  styleNo: string;
  name: string;
  category: string;
  purity: string;
  pieceCount: number;
  weightGram: number;
  costAmount: number;
  status: string;
  source: string;
  remark?: string;
  createdBy?: string;
  createdAt: string;
}

export interface InventoryLedgerSummary {
  orgId: string;
  visibleStoreCount: number;
  totalStyleCount: number;
  totalPieceCount: number;
  totalWeightGram: number;
  totalCostAmount: number;
  items: InventoryLedgerItem[];
}

export interface MaterialLedgerItem {
  id: string;
  orgId?: string;
  storeId: string;
  storeName: string;
  type: string;
  orderNo: string;
  customerName: string;
  category: string;
  purity: string;
  weightGram: number;
  amount: number;
  remainingWeightGram: number;
  status: string;
  dueDate?: string;
  remark?: string;
  source?: string;
  createdBy?: string;
  createdAt: string;
  outboundAt?: string;
  outboundBy?: string;
  outboundRemark?: string;
}

export interface MaterialLedgerSummary {
  orgId: string;
  visibleStoreCount: number;
  todayWeightGram: number;
  todayAmount: number;
  monthWeightGram: number;
  monthAmount: number;
  remainingWeightGram: number;
  pledgeCount: number;
  pledgeAmount: number;
  items: MaterialLedgerItem[];
  purityStats: Array<{ purity: string; count: number; weightGram: number; amount: number }>;
}

export interface MemberProfile {
  id: string;
  orgId: string;
  storeId: string;
  storeName: string;
  name: string;
  phone: string;
  level: string;
  status: string;
  totalOrders: number;
  totalRecycleAmount: number;
  lastVisitAt: string;
  preferredPurity: string;
  sourceChannel: string;
  managerName: string;
  idVerified: boolean;
  tags: string[];
  notes: string;
}

export interface CashierOrderLine {
  productId?: string;
  sku?: string;
  name: string;
  quantity: number;
  unitPrice: number;
  amount: number;
}

export interface CashierOrderView {
  id: string;
  orderNo: string;
  storeId: string;
  storeName: string;
  status: "paid" | "pending" | "refunded";
  customerName: string;
  customerPhone: string;
  paymentMethod?: string;
  totalAmount: number;
  paidAmount?: number;
  itemCount: number;
  itemSummary: string;
  items?: CashierOrderLine[];
  createdBy: string;
  createdAt: string;
  remark: string;
  voidReason?: string;
  voidedBy?: string;
  voidedAt?: string;
  refundReason?: string;
  refundedBy?: string;
  refundedAt?: string;
}

export interface RecycleItem {
  category: string;
  purity: string;
  weightGram: number;
}

export interface AttachmentAsset {
  id: string;
  orderId: string;
  storeId: string;
  category: string;
  storageProvider: string;
  objectKey: string;
  publicUrl: string;
  thumbnailUrl: string;
  fileName: string;
  contentType: string;
  sizeBytes: number;
  source: string;
  status: string;
  uploadedBy: string;
  uploadedAt: string;
  hasPreview?: boolean;
  previewUrl?: string;
}

export interface RecycleOrderView {
  id: string;
  orderNo: string;
  storeId: string;
  storeName: string;
  status: "draft" | "confirmed" | "cancelled";
  customerName: string;
  customerPhone: string;
  estimatedAmount: number;
  confirmedAmount: number;
  photoCount: number;
  itemSummary: string;
  items?: RecycleItem[];
  attachments?: AttachmentAsset[];
  createdBy: string;
  createdAt: string;
  confirmedAt?: string;
  remark: string;
  cancelReason?: string;
  cancelledBy?: string;
  cancelledAt?: string;
}

export interface ListQuery {
  page?: number;
  pageSize?: number;
  storeId?: string;
  status?: string;
  dateFrom?: string;
  dateTo?: string;
  keyword?: string;
}

export interface PagedItems<T> {
  items: T[];
  total: number;
  page: number;
  pageSize: number;
}

export interface AdminOrderRow {
  id: string;
  type: "cashier" | "recycle";
  orderNo: string;
  storeId: string;
  storeName: string;
  status: string;
  customerName: string;
  customerPhone: string;
  amount: number;
  itemSummary: string;
  createdBy: string;
  createdAt: string;
}

export interface SystemProfile {
  brandName: string;
  servicePhone: string;
  receiptTitle: string;
  minPhotoCount: number;
  maxPhotoCount: number;
  requireExactThree: boolean;
  requireIdCheck: boolean;
  domainName: string;
  domainStatus: string;
  ossStatus: string;
  appIdStatus: string;
  printerStatus: string;
}

export interface PrintTemplate {
  receipt: {
    enabled: boolean;
    paperWidth: string;
    headerTitle: string;
    footerNote: string;
    showStoreName: boolean;
    showOperatorName: boolean;
    showPhotoSummary: boolean;
    showPhotoThumbnails: boolean;
    fields: string[];
  };
  label: {
    enabled: boolean;
    size: string;
    copies: number;
    fields: string[];
    barcodeType: string;
  };
  recycle: {
    enabled: boolean;
    printPhotoSummary: boolean;
    printPhotoThumbnail: boolean;
    summaryText: string;
  };
}

export interface AuditLogRecord {
  id: string;
  module: string;
  action: string;
  operatorName: string;
  result: "success" | "warning" | "info";
  riskLevel: "high" | "medium" | "low";
  summary: string;
  createdAt: string;
}

export interface TemplateInitPlan {
  title: string;
  description: string;
  steps: Array<{
    id: string;
    title: string;
    description: string;
    status: "done" | "current" | "planned";
  }>;
  outputs: string[];
}

export interface ConsoleBootstrap {
  currentUser: SessionUser;
  dashboard: DashboardSummary;
  stores: StoreRecord[];
  users: UserAccount[];
  roles: RoleTemplate[];
  abilityGroups: AbilityGroup[];
  products: ProductRecord[];
  importLogs: ProductImportLog[];
  members: MemberProfile[];
  cashierOrders: CashierOrderView[];
  recycleOrders: RecycleOrderView[];
  printTemplate: PrintTemplate;
  systemProfile: SystemProfile;
  auditLogs: AuditLogRecord[];
  templateInit?: TemplateInitPlan;
  updatedAt: string;
}

export interface ResourceResult<T> {
  data: T;
  source: DataSource;
}

const API_BASE = import.meta.env.VITE_API_BASE || "/api";
const LOCAL_DATA_DELAY_MS = 180;
const ADMIN_LOCAL_DATA_ENABLED = ["1", "true", "yes", "on"].includes(
  String(import.meta.env.VITE_ENABLE_ADMIN_LOCAL_DATA || "").toLowerCase(),
);

export const ADMIN_API_MODE: AdminApiMode = ADMIN_LOCAL_DATA_ENABLED ? "local-data-enabled" : "strict-api";

function deepCopy<T>(value: T): T {
  if (value === undefined || value === null) return value;
  return JSON.parse(JSON.stringify(value)) as T;
}

function delay(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function authHeaders(token: string) {
  return { Authorization: `Bearer ${token}` };
}

function toQueryString(query: ListQuery = {}) {
  const params = new URLSearchParams();
  for (const [key, value] of Object.entries(query)) {
    if (value === undefined || value === null || value === "") continue;
    params.set(key, String(value));
  }
  const text = params.toString();
  return text ? `?${text}` : "";
}

class AdminApiError extends Error {
  kind: "network" | "response";
  status?: number;
  responseCode?: number;

  constructor(
    message: string,
    options: { kind: "network" | "response"; status?: number; responseCode?: number; cause?: unknown },
  ) {
    super(message, { cause: options.cause });
    this.name = "AdminApiError";
    this.kind = options.kind;
    this.status = options.status;
    this.responseCode = options.responseCode;
  }
}

function isPlainObject(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null;
}

function resolveApiErrorMessage(payload: unknown, defaultText: string): string {
  if (isPlainObject(payload) && typeof payload.message === "string" && payload.message.trim()) {
    return payload.message.trim();
  }

  if (defaultText.trim()) {
    return defaultText.trim();
  }

  return "请求失败。";
}

function canUseLocalData(error: unknown): boolean {
  return (
    ADMIN_LOCAL_DATA_ENABLED &&
    error instanceof AdminApiError &&
    (error.kind === "network" || [502, 503, 504].includes(Number(error.status || 0)))
  );
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers: {
        "Content-Type": "application/json",
        ...(options.headers || {}),
      },
    });
  } catch (error) {
    throw new AdminApiError(error instanceof Error ? error.message : "无法连接后台接口。", {
      kind: "network",
      cause: error,
    });
  }

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  let payload: ApiResponse<T> | null = null;
  if (text) {
    try {
      payload = JSON.parse(text) as ApiResponse<T>;
    } catch {
      payload = null;
    }
  } else {
    payload = { code: 0, message: "", data: undefined as T };
  }

  if (!res.ok) {
    throw new AdminApiError(resolveApiErrorMessage(payload, text || `请求失败（HTTP ${res.status}）。`), {
      kind: "response",
      status: res.status,
      responseCode: payload?.code,
    });
  }

  if (!payload) {
    throw new AdminApiError(text || "接口返回格式不正确。", {
      kind: "response",
      status: res.status,
    });
  }

  if (payload.code !== 0) {
    throw new AdminApiError(resolveApiErrorMessage(payload, `接口返回异常码 ${payload.code}。`), {
      kind: "response",
      status: res.status,
      responseCode: payload.code,
    });
  }

  return payload.data;
}

async function simulate<T>(value: T): Promise<ResourceResult<T>> {
  await delay(LOCAL_DATA_DELAY_MS);
  return { data: deepCopy(value), source: "local" };
}

export async function loginAdmin(phone: string, password: string, roleKey: RoleKey): Promise<ResourceResult<LoginResult>> {
  try {
    const data = await request<LoginResult>("/admin/login", {
      method: "POST",
      body: JSON.stringify({ phone, password, roleKey }),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) {
      throw error;
    }

    const localSession = getLocalSessionByCredentials(phone, password, roleKey);
    if (!localSession) {
      throw error;
    }
    return simulate(localSession);
  }
}

export async function fetchAdminConsole(token: string): Promise<ResourceResult<ConsoleBootstrap>> {
  try {
    const data = await request<ConsoleBootstrap>("/admin/bootstrap", {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) {
      throw error;
    }

    const localSession = getLocalSessionByToken(token);
    if (!localSession) {
      throw error;
    }
    return simulate(buildLocalConsoleBootstrap(localSession.user));
  }
}

async function listFromLocal<T>(token: string, picker: (data: ConsoleBootstrap) => T[]): Promise<ResourceResult<PagedItems<T>>> {
  const localSession = getLocalSessionByToken(token);
  if (!localSession) {
    throw new AdminApiError("未找到本地后台会话。", { kind: "network" });
  }
  const data = buildLocalConsoleBootstrap(localSession.user);
  return simulate({
    items: picker(data),
    total: picker(data).length,
    page: 1,
    pageSize: picker(data).length || 20,
  });
}

export async function fetchStoreRecords(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<StoreRecord>>> {
  try {
    const data = await request<PagedItems<StoreRecord>>(`/admin/stores${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    return listFromLocal(token, (data) => data.stores);
  }
}

export async function fetchUserAccounts(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<UserAccount>>> {
  try {
    const data = await request<PagedItems<UserAccount>>(`/admin/users${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    return listFromLocal(token, (data) => data.users);
  }
}

export async function fetchRoleTemplates(token: string): Promise<ResourceResult<RoleTemplate[]>> {
  try {
    const data = await request<RoleTemplate[]>("/admin/roles", {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    const localSession = getLocalSessionByToken(token);
    if (!localSession) throw error;
    return simulate(buildLocalConsoleBootstrap(localSession.user).roles);
  }
}

export async function fetchMemberProfiles(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<MemberProfile>>> {
  try {
    const data = await request<PagedItems<MemberProfile>>(`/admin/members${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    return listFromLocal(token, (data) => data.members);
  }
}

export async function fetchProductRecords(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<ProductRecord>>> {
  try {
    const data = await request<PagedItems<ProductRecord>>(`/admin/products${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    return listFromLocal(token, (data) => data.products);
  }
}

export async function fetchInventoryLedger(token: string): Promise<ResourceResult<InventoryLedgerSummary>> {
  const data = await request<InventoryLedgerSummary>("/admin/inventory/items", {
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function createInventoryLedgerItem(token: string, item: Partial<InventoryLedgerItem>): Promise<ResourceResult<InventoryLedgerItem>> {
  const data = await request<InventoryLedgerItem>("/admin/inventory/items", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(item),
  });
  return { data, source: "api" };
}

export async function saveInventoryLedgerItem(token: string, id: string, item: Partial<InventoryLedgerItem>): Promise<ResourceResult<InventoryLedgerItem>> {
  const data = await request<InventoryLedgerItem>(`/admin/inventory/items/${id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(item),
  });
  return { data, source: "api" };
}

export async function deleteInventoryLedgerItem(token: string, id: string): Promise<ResourceResult<InventoryLedgerItem>> {
  const data = await request<InventoryLedgerItem>(`/admin/inventory/items/${id}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function fetchMaterialLedger(token: string): Promise<ResourceResult<MaterialLedgerSummary>> {
  const data = await request<MaterialLedgerSummary>("/admin/materials", {
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function createMaterialLedgerItem(token: string, item: Partial<MaterialLedgerItem>): Promise<ResourceResult<MaterialLedgerItem>> {
  const data = await request<MaterialLedgerItem>("/admin/materials", {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify(item),
  });
  return { data, source: "api" };
}

export async function saveMaterialLedgerItem(token: string, id: string, item: Partial<MaterialLedgerItem>): Promise<ResourceResult<MaterialLedgerItem>> {
  const data = await request<MaterialLedgerItem>(`/admin/materials/${id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(item),
  });
  return { data, source: "api" };
}

export async function outboundMaterialLedgerItem(token: string, id: string, remark: string): Promise<ResourceResult<MaterialLedgerItem>> {
  const data = await request<MaterialLedgerItem>(`/admin/materials/${id}/outbound`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ remark }),
  });
  return { data, source: "api" };
}

export async function deleteMaterialLedgerItem(token: string, id: string): Promise<ResourceResult<MaterialLedgerItem>> {
  const data = await request<MaterialLedgerItem>(`/admin/materials/${id}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function fetchCashierOrders(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<CashierOrderView>>> {
  try {
    const data = await request<PagedItems<CashierOrderView>>(`/admin/cashier-orders${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    return listFromLocal(token, (data) => data.cashierOrders);
  }
}

export async function fetchCashierOrderDetail(token: string, id: string): Promise<ResourceResult<CashierOrderView>> {
  const data = await request<CashierOrderView>(`/admin/cashier-orders/${id}`, {
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function voidCashierOrder(token: string, id: string, reason: string): Promise<ResourceResult<CashierOrderView>> {
  const data = await request<CashierOrderView>(`/admin/cashier-orders/${id}/refund`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ reason }),
  });
  return { data, source: "api" };
}

export async function fetchRecycleOrders(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<RecycleOrderView>>> {
  try {
    const data = await request<PagedItems<RecycleOrderView>>(`/admin/recycle-orders${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    return listFromLocal(token, (data) => data.recycleOrders);
  }
}

export async function fetchRecycleOrderDetail(token: string, id: string): Promise<ResourceResult<RecycleOrderView>> {
  const data = await request<RecycleOrderView>(`/admin/recycle-orders/${id}`, {
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function cancelRecycleOrder(token: string, id: string, reason: string): Promise<ResourceResult<RecycleOrderView>> {
  const data = await request<RecycleOrderView>(`/admin/recycle-orders/${id}/cancel`, {
    method: "POST",
    headers: authHeaders(token),
    body: JSON.stringify({ reason }),
  });
  return { data, source: "api" };
}

export async function fetchOrderRows(token: string, query: ListQuery = {}): Promise<ResourceResult<PagedItems<AdminOrderRow>>> {
  try {
    const data = await request<PagedItems<AdminOrderRow>>(`/admin/orders${toQueryString(query)}`, {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) throw error;
    const localSession = getLocalSessionByToken(token);
    if (!localSession) throw error;
    const local = buildLocalConsoleBootstrap(localSession.user);
    const items: AdminOrderRow[] = [
      ...local.cashierOrders.map((item) => ({
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
      ...local.recycleOrders.map((item) => ({
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
    ];
    items.sort((left, right) => new Date(right.createdAt).getTime() - new Date(left.createdAt).getTime());
    return simulate({ items, total: items.length, page: 1, pageSize: items.length || 20 });
  }
}

export async function saveRoleTemplate(
  token: string,
  roleTemplate: RoleTemplate,
): Promise<ResourceResult<RoleTemplate>> {
  const data = await request<RoleTemplate>(`/admin/roles/${roleTemplate.id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(roleTemplate),
  });
  return { data, source: "api" };
}

export async function saveStoreRecord(token: string, storeRecord: StoreRecord): Promise<ResourceResult<StoreRecord>> {
  const data = await request<StoreRecord>(`/admin/stores/${storeRecord.id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(storeRecord),
  });
  return { data, source: "api" };
}

export async function saveUserAccount(token: string, userAccount: UserAccount): Promise<ResourceResult<UserAccount>> {
  const data = await request<UserAccount>(`/admin/users/${userAccount.id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(userAccount),
  });
  return { data, source: "api" };
}

export async function saveProductRecord(token: string, productRecord: ProductRecord): Promise<ResourceResult<ProductRecord>> {
  const data = await request<ProductRecord>(`/admin/products/${productRecord.id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(productRecord),
  });
  return { data, source: "api" };
}

export async function saveSystemProfile(token: string, systemProfile: SystemProfile): Promise<ResourceResult<SystemProfile>> {
  const data = await request<SystemProfile>("/admin/system-profile", {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(systemProfile),
  });
  return { data, source: "api" };
}

export async function savePrintTemplate(token: string, printTemplate: PrintTemplate): Promise<ResourceResult<PrintTemplate>> {
  const data = await request<PrintTemplate>("/admin/print-template", {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(printTemplate),
  });
  return { data, source: "api" };
}

export async function createStoreRecord(token: string): Promise<ResourceResult<StoreRecord>> {
  const data = await request<StoreRecord>("/admin/stores", {
    method: "POST",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function createUserAccount(token: string): Promise<ResourceResult<UserAccount>> {
  const data = await request<UserAccount>("/admin/users", {
    method: "POST",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function createProductRecord(token: string): Promise<ResourceResult<ProductRecord>> {
  const data = await request<ProductRecord>("/admin/products", {
    method: "POST",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function createMemberProfile(token: string): Promise<ResourceResult<MemberProfile>> {
  const data = await request<MemberProfile>("/admin/members", {
    method: "POST",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function saveMemberProfile(token: string, memberProfile: MemberProfile): Promise<ResourceResult<MemberProfile>> {
  const data = await request<MemberProfile>(`/admin/members/${memberProfile.id}`, {
    method: "PUT",
    headers: authHeaders(token),
    body: JSON.stringify(memberProfile),
  });
  return { data, source: "api" };
}

export async function disableStoreRecord(token: string, storeId: string): Promise<ResourceResult<StoreRecord>> {
  const data = await request<StoreRecord>(`/admin/stores/${storeId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function disableUserAccount(token: string, userId: string): Promise<ResourceResult<UserAccount>> {
  const data = await request<UserAccount>(`/admin/users/${userId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function disableProductRecord(token: string, productId: string): Promise<ResourceResult<ProductRecord>> {
  const data = await request<ProductRecord>(`/admin/products/${productId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}

export async function deleteProductRecords(token: string, productIds: string[]): Promise<ResourceResult<{ items: ProductRecord[]; count: number }>> {
  const data = await request<{ items: ProductRecord[]; count: number }>("/admin/products", {
    method: "DELETE",
    headers: authHeaders(token),
    body: JSON.stringify({ productIds }),
  });
  return { data, source: "api" };
}

export async function downloadProductImportTemplate(token: string, storeId?: string): Promise<Blob> {
  const query = storeId ? `?storeId=${encodeURIComponent(storeId)}` : "";
  const res = await fetch(`${API_BASE}/admin/products/import-template${query}`, {
    headers: authHeaders(token),
  });
  if (!res.ok) {
    throw new AdminApiError(`模板下载失败（HTTP ${res.status}）。`, { kind: "response", status: res.status });
  }
  return res.blob();
}

export async function importProducts(token: string, storeId: string, file: File): Promise<ResourceResult<ProductImportResult>> {
  const form = new FormData();
  form.set("storeId", storeId);
  form.set("file", file);
  const res = await fetch(`${API_BASE}/admin/products/import`, {
    method: "POST",
    headers: authHeaders(token),
    body: form,
  });
  const text = await res.text();
  let payload: ApiResponse<ProductImportResult> | null = null;
  if (text) {
    try {
      payload = JSON.parse(text) as ApiResponse<ProductImportResult>;
    } catch {
      payload = null;
    }
  }
  if (!res.ok || !payload || payload.code !== 0) {
    throw new AdminApiError(resolveApiErrorMessage(payload, text || `导入失败（HTTP ${res.status}）。`), {
      kind: "response",
      status: res.status,
      responseCode: payload?.code,
    });
  }
  return { data: payload.data, source: "api" };
}

export async function disableMemberProfile(token: string, memberId: string): Promise<ResourceResult<MemberProfile>> {
  const data = await request<MemberProfile>(`/admin/members/${memberId}`, {
    method: "DELETE",
    headers: authHeaders(token),
  });
  return { data, source: "api" };
}
