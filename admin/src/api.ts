import {
  buildMockConsoleBootstrap,
  getMockSessionByCredentials,
  getMockSessionByToken,
} from "./fallback";

export interface ApiResponse<T> {
  code: number;
  message: string;
  data: T;
}

export type DataSource = "api" | "mock";
export type DataScope = "all_stores" | "assigned_store" | "self";
export type RoleKey = "owner" | "manager" | "clerk";
export type StoreStatus = "active" | "pending" | "disabled";
export type UserStatus = "enabled" | "disabled" | "invited";
export type PaymentMethod = "wechat" | "cash" | "bank_transfer";
export type PaymentStatus = "paid" | "refunding" | "failed";
export type CallbackStatus = "delivered" | "pending" | "exception";
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
  | "payment-config"
  | "payment-records"
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
  | "payment.config.view"
  | "payment.config.manage"
  | "payment.record.view"
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

export interface PaymentConfig {
  wechat: {
    enabled: boolean;
    appId: string;
    merchantId: string;
    subMerchantId: string;
    certificateStatus: string;
    callbackUrl: string;
    sandboxMode: boolean;
    lastVerifiedAt: string;
  };
  cash: {
    enabled: boolean;
    receiptRequired: boolean;
    shiftReconciliationRequired: boolean;
  };
  bankTransfer: {
    enabled: boolean;
    accountName: string;
    accountSuffix: string;
  };
  reconciliation: {
    autoRetryEnabled: boolean;
    retryMinutes: number;
    abnormalNotify: string;
  };
}

export interface PaymentRecord {
  id: string;
  paymentNo: string;
  orderNo: string;
  bizType: "retail" | "recycle";
  storeId: string;
  storeName: string;
  amount: number;
  method: PaymentMethod;
  status: PaymentStatus;
  callbackStatus: CallbackStatus;
  paidAt: string;
  operatorName: string;
  customerLabel: string;
  remark: string;
  anomaly: boolean;
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
  paymentMethod: PaymentMethod;
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
  wechatPayEnabled: boolean;
  cashEnabled: boolean;
  bankTransferEnabled: boolean;
  domainName: string;
  domainStatus: string;
  ossStatus: string;
  appIdStatus: string;
  merchantStatus: string;
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
    showPaymentMethod: boolean;
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
  paymentConfig: PaymentConfig;
  printTemplate: PrintTemplate;
  paymentRecords: PaymentRecord[];
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
const MOCK_DELAY_MS = 180;

function deepCopy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function delay(ms: number) {
  return new Promise((resolve) => window.setTimeout(resolve, ms));
}

function authHeaders(token: string) {
  return { Authorization: `Bearer ${token}` };
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    ...options,
    headers: {
      "Content-Type": "application/json",
      ...(options.headers || {}),
    },
  });

  if (res.status === 204) {
    return undefined as T;
  }

  const text = await res.text();
  const payload = text
    ? (JSON.parse(text) as ApiResponse<T>)
    : ({ code: 0, message: "", data: undefined as T } satisfies ApiResponse<T>);

  if (!res.ok || payload.code !== 0) {
    throw new Error(payload.message || "Request failed");
  }

  return payload.data;
}

async function simulate<T>(value: T): Promise<ResourceResult<T>> {
  await delay(MOCK_DELAY_MS);
  return { data: deepCopy(value), source: "mock" };
}

export async function loginAdmin(username: string, password: string): Promise<ResourceResult<LoginResult>> {
  try {
    const data = await request<LoginResult>("/admin/login", {
      method: "POST",
      body: JSON.stringify({ username, password }),
    });
    return { data, source: "api" };
  } catch {
    const mockSession = getMockSessionByCredentials(username, password);
    if (!mockSession) {
      throw new Error("账号或密码不正确，请检查后重试。");
    }
    return simulate(mockSession);
  }
}

export async function fetchAdminConsole(token: string): Promise<ResourceResult<ConsoleBootstrap>> {
  try {
    const data = await request<ConsoleBootstrap>("/admin/bootstrap", {
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch {
    const mockSession = getMockSessionByToken(token);
    if (!mockSession) {
      throw new Error("登录态已失效，请重新登录。");
    }
    return simulate(buildMockConsoleBootstrap(mockSession.user));
  }
}

export async function saveRoleTemplate(
  token: string,
  roleTemplate: RoleTemplate,
): Promise<ResourceResult<RoleTemplate>> {
  try {
    const data = await request<RoleTemplate>(`/admin/roles/${roleTemplate.id}`, {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(roleTemplate),
    });
    return { data, source: "api" };
  } catch {
    return simulate(roleTemplate);
  }
}

export async function savePaymentCenterConfig(
  token: string,
  paymentConfig: PaymentConfig,
): Promise<ResourceResult<PaymentConfig>> {
  try {
    const data = await request<PaymentConfig>("/admin/payment-config", {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(paymentConfig),
    });
    return { data, source: "api" };
  } catch {
    return simulate(paymentConfig);
  }
}

export async function saveStoreRecord(token: string, storeRecord: StoreRecord): Promise<ResourceResult<StoreRecord>> {
  try {
    const data = await request<StoreRecord>(`/admin/stores/${storeRecord.id}`, {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(storeRecord),
    });
    return { data, source: "api" };
  } catch {
    return simulate(storeRecord);
  }
}

export async function saveUserAccount(token: string, userAccount: UserAccount): Promise<ResourceResult<UserAccount>> {
  try {
    const data = await request<UserAccount>(`/admin/users/${userAccount.id}`, {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(userAccount),
    });
    return { data, source: "api" };
  } catch {
    return simulate(userAccount);
  }
}

export async function saveProductRecord(token: string, productRecord: ProductRecord): Promise<ResourceResult<ProductRecord>> {
  try {
    const data = await request<ProductRecord>(`/admin/products/${productRecord.id}`, {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(productRecord),
    });
    return { data, source: "api" };
  } catch {
    return simulate(productRecord);
  }
}

export async function saveSystemProfile(token: string, systemProfile: SystemProfile): Promise<ResourceResult<SystemProfile>> {
  try {
    const data = await request<SystemProfile>("/admin/system-profile", {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(systemProfile),
    });
    return { data, source: "api" };
  } catch {
    return simulate(systemProfile);
  }
}

export async function savePrintTemplate(token: string, printTemplate: PrintTemplate): Promise<ResourceResult<PrintTemplate>> {
  try {
    const data = await request<PrintTemplate>("/admin/print-template", {
      method: "PUT",
      headers: authHeaders(token),
      body: JSON.stringify(printTemplate),
    });
    return { data, source: "api" };
  } catch {
    return simulate(printTemplate);
  }
}

export async function createStoreRecord(token: string): Promise<ResourceResult<StoreRecord>> {
  try {
    const data = await request<StoreRecord>("/admin/stores", {
      method: "POST",
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch {
    return simulate({
      id: `store-local-${Date.now()}`,
      code: "LOCAL",
      name: "新门店",
      managerName: "待分配",
      city: "待填写",
      address: "待填写地址",
      contactPhone: "待填写电话",
      businessHours: "10:00 - 22:00",
      status: "pending",
      cashierDevices: 0,
      pendingTasks: 0,
      todayAmount: 0,
      todayOrders: 0,
      lastSettlementAt: "尚未营业",
      tags: ["新建门店", "待完善"],
    });
  }
}

export async function createUserAccount(token: string): Promise<ResourceResult<UserAccount>> {
  try {
    const data = await request<UserAccount>("/admin/users", {
      method: "POST",
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch {
    return simulate({
      id: `user-local-${Date.now()}`,
      name: "新账号",
      account: "new.user",
      phone: "待填写",
      roleKey: "clerk",
      roleName: "员工",
      dataScope: "self",
      storeIds: [],
      storeNames: ["待绑定门店"],
      status: "invited",
      lastLoginAt: "未登录",
      abilities: [],
    });
  }
}

export async function createProductRecord(token: string): Promise<ResourceResult<ProductRecord>> {
  try {
    const data = await request<ProductRecord>("/admin/products", {
      method: "POST",
      headers: authHeaders(token),
    });
    return { data, source: "api" };
  } catch {
    return simulate({
      id: `prd-local-${Date.now()}`,
      name: "新商品",
      sku: "LOCAL-SKU",
      category: "待分类",
      price: 0,
      gramWeight: 0,
      status: "draft",
      storeNames: ["待分配门店"],
      tags: ["新建", "待完善"],
    });
  }
}
