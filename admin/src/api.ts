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
export type RoleKey = "owner" | "manager" | "staff";
export type StoreStatus = "active" | "pending" | "disabled";
export type UserStatus = "enabled" | "disabled" | "invited";
export type MetricTone = "gold" | "emerald" | "slate" | "danger";
export type TodoLevel = "high" | "medium" | "low";
export type PageId =
  | "dashboard"
  | "stores"
  | "users"
  | "roles"
  | "products"
  | "orders"
  | "recycle"
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
  price: number;
  gramWeight: number;
  status: "active" | "draft" | "disabled";
  storeNames: string[];
  tags: string[];
}

export interface CashierOrderView {
  id: string;
  orderNo: string;
  storeId: string;
  storeName: string;
  status: "paid" | "pending" | "refunded";
  totalAmount: number;
  itemCount: number;
  createdBy: string;
  createdAt: string;
  remark: string;
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
  createdBy: string;
  createdAt: string;
  confirmedAt?: string;
  remark: string;
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
  cashierOrders: CashierOrderView[];
  recycleOrders: RecycleOrderView[];
  printTemplate: PrintTemplate;
  systemProfile: SystemProfile;
  auditLogs: AuditLogRecord[];
  templateInit: TemplateInitPlan;
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
  return JSON.parse(JSON.stringify(value)) as T;
}

function delay(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function authHeaders(token: string) {
  return { Authorization: `Bearer ${token}` };
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
  return ADMIN_LOCAL_DATA_ENABLED && error instanceof AdminApiError && error.kind === "network";
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

export async function loginAdmin(username: string, password: string): Promise<ResourceResult<LoginResult>> {
  try {
    const data = await request<LoginResult>("/admin/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    return { data, source: "api" };
  } catch (error) {
    if (!canUseLocalData(error)) {
      throw error;
    }

    const localSession = getLocalSessionByCredentials(username, password);
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
