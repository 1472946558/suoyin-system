import type {
  AbilityCode,
  AbilityGroup,
  AuditLogRecord,
  CashierOrderView,
  ConsoleBootstrap,
  DashboardMetric,
  DashboardShortcut,
  DashboardTodo,
  DataScope,
  LoginResult,
  PrintTemplate,
  ProductRecord,
  RecycleOrderView,
  RoleTemplate,
  SessionUser,
  StoreRecord,
  SystemProfile,
  TemplateInitPlan,
  UserAccount,
} from "./api";

function deepCopy<T>(value: T): T {
  return JSON.parse(JSON.stringify(value)) as T;
}

function formatCurrency(amount: number) {
  return `¥${amount.toLocaleString("zh-CN")}`;
}

function includesStore(accountStoreIds: string[], currentStoreIds: string[]) {
  return currentStoreIds.some((storeId) => accountStoreIds.includes(storeId));
}

const ALL_ABILITIES: AbilityCode[] = [
  "dashboard.view",
  "store.view",
  "store.manage",
  "user.view",
  "user.manage",
  "role.view",
  "role.manage",
  "product.view",
  "product.manage",
  "order.view",
  "recycle.view",
  "system.config.view",
  "system.config.manage",
  "template.init",
  "audit.view",
];

const ownerUser: SessionUser = {
  id: "user-owner-001",
  name: "廖总",
  account: "boss",
  roleKey: "owner",
  roleName: "老板",
  dataScope: "all_stores",
  storeIds: [],
  abilities: ["auth.login", ...ALL_ABILITIES],
  lastLoginAt: "2026-06-04 16:50",
};

const managerUser: SessionUser = {
  id: "user-manager-001",
  name: "李店长",
  account: "manager.sz",
  roleKey: "manager",
  roleName: "店长",
  dataScope: "assigned_store",
  storeIds: ["store-sz-luohu"],
  abilities: [
    "auth.login",
    "dashboard.view",
    "store.view",
    "store.manage",
    "user.view",
    "user.manage",
    "product.view",
    "product.manage",
    "order.view",
    "recycle.view",
  ],
  lastLoginAt: "2026-05-10 08:42",
};

const localSessions: LoginResult[] = [
  {
    token: "local-owner-token",
    user: ownerUser,
    landingPage: "dashboard",
  },
  {
    token: "local-manager-token",
    user: managerUser,
    landingPage: "dashboard",
  },
];

const stores: StoreRecord[] = [
  {
    id: "store-sz-luohu",
    code: "SZ-LH-01",
    name: "罗湖旗舰店",
    managerName: "李店长",
    city: "深圳",
    address: "罗湖区深南东路 1888 号 1F",
    contactPhone: "0755-8899 1201",
    businessHours: "10:00 - 22:00",
    status: "active",
    cashierDevices: 3,
    pendingTasks: 2,
    todayAmount: 18260,
    todayOrders: 16,
    lastSettlementAt: "2026-05-10 10:05",
    tags: ["黄金回收", "收银台", "微信转账登记"],
  },
  {
    id: "store-sz-nanshan",
    code: "SZ-NS-02",
    name: "南山体验店",
    managerName: "王店长",
    city: "深圳",
    address: "南山区科苑路 66 号 B1",
    contactPhone: "0755-8899 1202",
    businessHours: "10:00 - 21:30",
    status: "active",
    cashierDevices: 2,
    pendingTasks: 1,
    todayAmount: 13680,
    todayOrders: 11,
    lastSettlementAt: "2026-05-10 09:35",
    tags: ["黄金回收", "体验店", "现金记账"],
  },
  {
    id: "store-sz-baoan",
    code: "SZ-BA-03",
    name: "宝安社区店",
    managerName: "陈店长",
    city: "深圳",
    address: "宝安区前进一路 277 号",
    contactPhone: "0755-8899 1203",
    businessHours: "09:30 - 21:00",
    status: "pending",
    cashierDevices: 1,
    pendingTasks: 4,
    todayAmount: 0,
    todayOrders: 0,
    lastSettlementAt: "待开业",
    tags: ["模板初始化", "设备待配置"],
  },
  {
    id: "store-gz-tianhe",
    code: "GZ-TH-01",
    name: "广州天河店",
    managerName: "空缺",
    city: "广州",
    address: "天河区体育西路 88 号",
    contactPhone: "020-8899 1201",
    businessHours: "10:00 - 22:00",
    status: "disabled",
    cashierDevices: 0,
    pendingTasks: 3,
    todayAmount: 0,
    todayOrders: 0,
    lastSettlementAt: "2026-05-06 21:30",
    tags: ["暂停营业", "待重新授权"],
  },
];

const accounts: UserAccount[] = [
  {
    id: "user-owner-001",
    name: "廖总",
    account: "boss",
    phone: "151****5083",
    roleKey: "owner",
    roleName: "老板",
    dataScope: "all_stores",
    storeIds: [],
    storeNames: ["全部门店"],
    status: "enabled",
    lastLoginAt: "2026-06-04 16:50",
    abilities: ["auth.login", ...ALL_ABILITIES],
  },
  {
    id: "user-owner-002",
    name: "示例老板",
    account: "boss.demo",
    phone: "134****4944",
    roleKey: "owner",
    roleName: "老板",
    dataScope: "all_stores",
    storeIds: [],
    storeNames: ["全部门店"],
    status: "enabled",
    lastLoginAt: "2026-06-04 16:50",
    abilities: ["auth.login", ...ALL_ABILITIES],
  },
  {
    id: "account-002",
    name: "李店长",
    account: "manager.sz",
    phone: "138****2108",
    roleKey: "manager",
    roleName: "店长",
    dataScope: "assigned_store",
    storeIds: ["store-sz-luohu"],
    storeNames: ["罗湖旗舰店"],
    status: "enabled",
    lastLoginAt: "2026-05-10 08:42",
    abilities: managerUser.abilities,
  },
  {
    id: "account-003",
    name: "王店长",
    account: "manager.sz02",
    phone: "138****2109",
    roleKey: "manager",
    roleName: "店长",
    dataScope: "assigned_store",
    storeIds: ["store-sz-nanshan"],
    storeNames: ["南山体验店"],
    status: "enabled",
    lastLoginAt: "2026-05-09 20:16",
    abilities: managerUser.abilities,
  },
  {
    id: "account-004",
    name: "陈财务",
    account: "finance.ops",
    phone: "138****3311",
    roleKey: "manager",
    roleName: "店长（受限）",
    dataScope: "assigned_store",
    storeIds: ["store-sz-luohu", "store-sz-nanshan"],
    storeNames: ["罗湖旗舰店", "南山体验店"],
    status: "invited",
    lastLoginAt: "未登录",
    abilities: [
      "auth.login",
      "dashboard.view",
      "store.view",
      "user.view",
    ],
  },
  {
    id: "account-005",
    name: "赵店员",
    account: "staff.sz01",
    phone: "138****4410",
    roleKey: "staff",
    roleName: "员工",
    dataScope: "self",
    storeIds: ["store-sz-luohu"],
    storeNames: ["罗湖旗舰店"],
    status: "disabled",
    lastLoginAt: "2026-05-05 18:20",
    abilities: ["auth.login"],
  },
];

const roleTemplates: RoleTemplate[] = [
  {
    id: 1,
    key: "owner",
    name: "老板模板",
    description: "全局主控权限，覆盖组织、商品、订单、系统与模板初始化。",
    dataScope: "all_stores",
    memberCount: 1,
    locked: true,
    abilities: ALL_ABILITIES,
  },
  {
    id: 2,
    key: "manager",
    name: "店长模板",
    description: "本门店运营台，可管理门店资料、账号并查看业务订单。",
    dataScope: "assigned_store",
    memberCount: 3,
    locked: true,
    abilities: [
      "dashboard.view",
      "store.view",
      "store.manage",
      "user.view",
      "user.manage",
      "product.view",
      "product.manage",
      "order.view",
      "recycle.view",
    ],
  },
  {
    id: 3,
    key: "staff",
    name: "员工模板",
    description: "第一版默认不开放后台入口，仅保留模板位用于后续扩展。",
    dataScope: "self",
    memberCount: 1,
    locked: true,
    abilities: [],
  },
];

const abilityGroups: AbilityGroup[] = [
  {
    key: "dashboard",
    label: "首页",
    description: "后台首页概览与快捷入口能力。",
    items: [{ code: "dashboard.view", label: "查看首页", description: "查看经营概览、待办和快捷入口。" }],
  },
  {
    key: "organization",
    label: "组织与权限",
    description: "门店、账号和角色模板相关能力。",
    items: [
      { code: "store.view", label: "查看门店", description: "查看门店资料、状态和营业信息。" },
      { code: "store.manage", label: "管理门店", description: "编辑门店资料、负责人和启停用入口。" },
      { code: "user.view", label: "查看账号", description: "查看后台账号、角色绑定和登录状态。" },
      { code: "user.manage", label: "管理账号", description: "启停用账号、重置密码和绑定门店。" },
      { code: "role.view", label: "查看角色权限", description: "查看固定角色模板和数据范围。" },
      { code: "role.manage", label: "管理角色权限", description: "调整能力项和数据范围配置。" },
    ],
  },
  {
    key: "business",
    label: "商品与业务",
    description: "第二阶段业务主线，当前先占位能力码。",
    items: [
      { code: "product.view", label: "查看商品", description: "查看商品分类、SKU 和状态。" },
      { code: "product.manage", label: "管理商品", description: "维护分类、价格和上架状态。" },
      { code: "order.view", label: "查看订单", description: "查看收银订单、金额和操作人信息。" },
      { code: "recycle.view", label: "查看回收单", description: "查看回收单、图片留痕和确认状态。" },
    ],
  },
  {
    key: "system",
    label: "系统与模板",
    description: "全局配置、模板化初始化和审计留痕。",
    items: [
      {
        code: "system.config.view",
        label: "查看系统配置",
        description: "查看拍照规则、编号规则和品牌信息。",
      },
      {
        code: "system.config.manage",
        label: "管理系统配置",
        description: "修改全局业务规则和基础配置。",
      },
      { code: "template.init", label: "模板初始化", description: "执行新客户初始化向导。" },
      { code: "audit.view", label: "查看操作日志", description: "查看权限变更、系统配置等审计记录。" },
    ],
  },
];

const products: ProductRecord[] = [
  {
    id: "prd-001",
    name: "足金项链",
    sku: "GJ-XL-001",
    category: "项链",
    price: 3298,
    gramWeight: 8.6,
    status: "active",
    storeNames: ["罗湖旗舰店", "南山体验店"],
    tags: ["热卖", "足金"],
  },
  {
    id: "prd-002",
    name: "古法手镯",
    sku: "GJ-SZ-018",
    category: "手镯",
    price: 5680,
    gramWeight: 15.2,
    status: "active",
    storeNames: ["罗湖旗舰店"],
    tags: ["回购高", "礼盒装"],
  },
  {
    id: "prd-003",
    name: "回收服务单",
    sku: "REC-SRV-001",
    category: "回收服务",
    price: 0,
    gramWeight: 0,
    status: "draft",
    storeNames: ["罗湖旗舰店", "南山体验店", "广州天河店"],
    tags: ["系统单据", "非销售商品"],
  },
];

const cashierOrders: CashierOrderView[] = [
  {
    id: "ord-001",
    orderNo: "ORD202605100001",
    storeId: "store-sz-luohu",
    storeName: "罗湖旗舰店",
    status: "paid",
    totalAmount: 3298,
    itemCount: 2,
    createdBy: "李店长",
    createdAt: "2026-05-10 10:08",
    remark: "足金项链成交",
  },
  {
    id: "ord-002",
    orderNo: "ORD202605090145",
    storeId: "store-sz-nanshan",
    storeName: "南山体验店",
    status: "pending",
    totalAmount: 2680,
    itemCount: 1,
    createdBy: "王店长",
    createdAt: "2026-05-09 19:42",
    remark: "待主管复核",
  },
  {
    id: "ord-003",
    orderNo: "ORD202605070201",
    storeId: "store-gz-tianhe",
    storeName: "广州天河店",
    status: "refunded",
    totalAmount: 980,
    itemCount: 1,
    createdBy: "系统迁移",
    createdAt: "2026-05-07 20:20",
    remark: "停店前历史订单",
  },
];

const recycleOrders: RecycleOrderView[] = [
  {
    id: "rec-001",
    orderNo: "REC202605100014",
    storeId: "store-sz-luohu",
    storeName: "罗湖旗舰店",
    status: "confirmed",
    customerName: "周女士",
    customerPhone: "134****2188",
    estimatedAmount: 5800,
    confirmedAmount: 5680,
    photoCount: 3,
    createdBy: "李店长",
    createdAt: "2026-05-10 09:18",
    confirmedAt: "2026-05-10 09:32",
    remark: "三张现场图已留存",
  },
  {
    id: "rec-002",
    orderNo: "REC202605090021",
    storeId: "store-sz-nanshan",
    storeName: "南山体验店",
    status: "draft",
    customerName: "王先生",
    customerPhone: "135****1018",
    estimatedAmount: 4320,
    confirmedAmount: 0,
    photoCount: 2,
    createdBy: "王店长",
    createdAt: "2026-05-09 16:18",
    remark: "待复核回收价",
  },
];

const systemProfile: SystemProfile = {
  brandName: "金匠馆回收",
  servicePhone: "400-888-2026",
  receiptTitle: "黄金回收收银系统",
  minPhotoCount: 2,
  maxPhotoCount: 3,
  requireExactThree: false,
  requireIdCheck: true,
  domainName: "example.com",
  domainStatus: "已购买，实名认证审核中",
  ossStatus: "对象存储待配置",
  appIdStatus: "使用占位 AppID，正式 AppID 请通过私有环境配置",
  printerStatus: "待确定小票机 / 标签机型号，可按门店需求配置打印设备",
};

const printTemplate: PrintTemplate = {
  receipt: {
    enabled: true,
    paperWidth: "80mm",
    headerTitle: "黄金回收收银系统",
    footerNote: "请当面核对金额与留痕信息",
    showStoreName: true,
    showOperatorName: true,
    showPhotoSummary: true,
    showPhotoThumbnails: false,
    fields: ["订单号", "客户信息", "商品信息", "金额", "照片已留存"],
  },
  label: {
    enabled: true,
    size: "50x30mm",
    copies: 1,
    fields: ["标签标题", "门店", "SKU/单号", "金额/克重"],
    barcodeType: "CODE128",
  },
  recycle: {
    enabled: true,
    printPhotoSummary: true,
    printPhotoThumbnail: false,
    summaryText: "已留存现场照片 {photoCount} 张",
  },
};

const auditLogs: AuditLogRecord[] = [
  {
    id: "log-001",
    module: "系统配置",
    action: "确认业务范围",
    operatorName: "周老板",
    result: "warning",
    riskLevel: "high",
    summary: "小程序仅保留业务订单和回收留档。",
    createdAt: "2026-05-10 11:05",
  },
  {
    id: "log-002",
    module: "角色权限",
    action: "冻结店长模板",
    operatorName: "周老板",
    result: "success",
    riskLevel: "medium",
    summary: "已确认店长默认只能看所属门店数据。",
    createdAt: "2026-05-10 10:26",
  },
  {
    id: "log-003",
    module: "系统配置",
    action: "调整拍照规则",
    operatorName: "系统",
    result: "info",
    riskLevel: "low",
    summary: "当前规则为至少 2 张最多 3 张，保留切换为必须 3 张的能力。",
    createdAt: "2026-05-10 09:40",
  },
];

const templatePlan: TemplateInitPlan = {
  title: "模板初始化入口",
  description: "用于新客户开通时一次性落默认门店、角色模板、拍照规则和品牌基础配置。",
  steps: [
    {
      id: "step-1",
      title: "品牌基础信息",
      description: "品牌名、客服电话、小票抬头和 Logo 资料。",
      status: "done",
    },
    {
      id: "step-2",
      title: "默认门店模板",
      description: "创建默认门店、营业时间、负责人和设备位。",
      status: "current",
    },
    {
      id: "step-3",
      title: "固定角色模板",
      description: "下发老板、店长、员工模板与初始能力组。",
      status: "planned",
    },
    {
      id: "step-4",
      title: "系统规则",
      description: "拍照规则、编号规则、对象存储和打印模板。",
      status: "planned",
    },
  ],
  outputs: [
    "默认门店模板 1 套",
    "固定角色模板 3 套",
    "拍照留档规则预设 1 组",
    "品牌基础配置包 1 份",
  ],
};

function getScopeStoreIds(user: SessionUser) {
  return user.dataScope === "all_stores" ? stores.map((item) => item.id) : user.storeIds;
}

function buildDashboardMetrics(user: SessionUser, scopedStores: StoreRecord[]): DashboardMetric[] {
  const pendingCount = scopedStores.reduce((sum, store) => sum + store.pendingTasks, 0);
  const storeSummary = user.dataScope === "all_stores" ? `${scopedStores.length} 店视角` : "本店视角";

  return [
    {
      key: "gross",
      label: "今日收银额",
      value: formatCurrency(scopedStores.reduce((sum, store) => sum + store.todayAmount, 0)),
      delta: `${storeSummary} / 结构占位可接真实报表`,
      tone: "gold",
    },
    {
      key: "orders",
      label: "今日订单数",
      value: scopedStores.reduce((sum, store) => sum + store.todayOrders, 0).toString(),
      delta: user.dataScope === "all_stores" ? "老板主控台聚合" : "默认按所属门店过滤",
      tone: "emerald",
    },
    {
      key: "profit",
      label: "回收利润占位",
      value: formatCurrency(Math.round(scopedStores.reduce((sum, store) => sum + store.todayAmount, 0) * 0.082)),
      delta: "等待后端利润口径冻结后替换",
      tone: "slate",
    },
    {
      key: "pending",
      label: "待处理事项",
      value: pendingCount.toString(),
      delta: "权限确认 / 开店准备 / 打印适配",
      tone: pendingCount > 1 ? "danger" : "slate",
    },
  ];
}

function buildDashboardTodos(user: SessionUser, scopedStores: StoreRecord[]): DashboardTodo[] {
  const pendingStores = scopedStores.filter((store) => store.pendingTasks > 0);

  const todos: DashboardTodo[] = [];

  if (pendingStores.length > 0) {
    todos.push({
      id: "task-store",
      title: `${pendingStores[0].name} 仍有 ${pendingStores[0].pendingTasks} 项开店待办`,
      description: "门店营业时间、负责人或设备位仍未补齐。",
      level: "high",
      page: "stores",
    });
  }

  if (user.roleKey === "owner") {
    todos.push({
      id: "task-role",
      title: "固定角色模板待老板确认",
      description: "建议冻结店长模板的数据范围和业务查看能力。",
      level: "medium",
      page: "roles",
    });
  }

  return todos;
}

function buildDashboardShortcuts(user: SessionUser): DashboardShortcut[] {
  const shortcuts: DashboardShortcut[] = [
    {
      id: "shortcut-store",
      label: "门店状态巡检",
      hint: "检查营业状态、负责人和设备位",
      page: "stores",
      ability: "store.view",
    },
    {
      id: "shortcut-user",
      label: "账号启停用",
      hint: "处理离岗账号和邀请中的账号",
      page: "users",
      ability: "user.view",
    },
  ];

  if (user.roleKey === "owner") {
    shortcuts.unshift({
      id: "shortcut-role",
      label: "角色模板冻结",
      hint: "调整能力项和数据范围骨架",
      page: "roles",
      ability: "role.view",
    });
  }

  return shortcuts;
}

function buildDashboardNotices(user: SessionUser): string[] {
  const notices = [
    "当前离线数据仅用于内部浏览，保存和新建仍以真实后端为准。",
    "当前版本聚焦门店收银、回收留档和权限管理。",
  ];

  if (user.roleKey === "manager") {
    notices.unshift("店长默认仅查看所属门店数据，菜单和默认筛选已按数据范围收敛。");
  }

  if (user.roleKey === "owner") {
    notices.unshift("老板模板拥有全门店主控权限，后续联调应由后端继续做真实鉴权和审计留痕。");
  }

  return notices;
}

export function getLocalSessionByCredentials(username: string, password: string) {
  const allowed = localSessions.find((item) => item.user.account === username);
  if (!allowed) return null;
  const passwordMap: Record<string, string> = {
    boss: "Boss123!",
    "manager.sz": "Manager123!",
  };
  if (password !== passwordMap[username]) return null;
  const session = localSessions.find((item) => item.user.account === username);
  return session ? deepCopy(session) : null;
}

export function getLocalSessionByToken(token: string) {
  const session = localSessions.find((item) => item.token === token);
  return session ? deepCopy(session) : null;
}

export function buildLocalConsoleBootstrap(user: SessionUser): ConsoleBootstrap {
  const scopedStoreIds = getScopeStoreIds(user);
  const scopedStores =
    user.dataScope === "all_stores"
      ? stores
      : stores.filter((store) => scopedStoreIds.includes(store.id));
  const scopedUsers =
    user.dataScope === "all_stores"
      ? accounts
      : accounts.filter((account) => includesStore(account.storeIds, scopedStoreIds) || account.id === user.id);
  return {
    currentUser: deepCopy(user),
    dashboard: {
      title: user.roleKey === "owner" ? "老板主控台" : "门店运营台",
      subtitle:
        user.roleKey === "owner"
          ? "覆盖门店、账号、角色权限、商品订单和模板初始化的首版后台骨架。"
          : "已按本门店视角收敛数据与可见菜单，可直接作为店长受限后台演示。",
      metrics: buildDashboardMetrics(user, scopedStores),
      todos: buildDashboardTodos(user, scopedStores),
      shortcuts: buildDashboardShortcuts(user),
      notices: buildDashboardNotices(user),
    },
    stores: deepCopy(scopedStores),
    users: deepCopy(scopedUsers),
    roles: deepCopy(roleTemplates),
    abilityGroups: deepCopy(abilityGroups),
    products: deepCopy(products),
    cashierOrders: deepCopy(
      user.dataScope === "all_stores"
        ? cashierOrders
        : cashierOrders.filter((order) => scopedStoreIds.includes(order.storeId)),
    ),
    recycleOrders: deepCopy(
      user.dataScope === "all_stores"
        ? recycleOrders
        : recycleOrders.filter((order) => scopedStoreIds.includes(order.storeId)),
    ),
    printTemplate: deepCopy(printTemplate),
    systemProfile: deepCopy(systemProfile),
    auditLogs: deepCopy(auditLogs),
    templateInit: deepCopy(templatePlan),
    updatedAt: "2026-05-10 11:30",
  };
}
