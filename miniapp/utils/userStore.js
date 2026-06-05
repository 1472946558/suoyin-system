const USER_KEY = "gr_operator_profile";
const { appConfig, request } = require("./apiClient");
const { clearScopedBusinessData, clearLegacyBusinessData } = require("./sessionStorage");

const rolePresets = [
  {
    key: "owner",
    label: "老板",
    permissions: ["dashboard.view", "cashier.use", "recycle.use", "orders.view", "members.view", "products.view", "settings.view", "settings.manage"]
  },
  {
    key: "manager",
    label: "店长",
    permissions: ["dashboard.view", "cashier.use", "recycle.use", "orders.view", "members.view", "products.view", "settings.view"]
  },
  {
    key: "staff",
    label: "员工",
    permissions: ["cashier.use", "recycle.use", "orders.view", "members.view", "products.view"]
  }
];

const devtoolAccounts = {
  owner: { username: "boss", password: "Boss123!" },
  manager: { username: "manager.sz", password: "Manager123!" },
  staff: { username: "cashier.sz", password: "Cashier123!" }
};

const quickLoginProfiles = {
  owner: {
    name: "廖总",
    phone: "13800000001",
    roleKey: "owner",
    role: "老板",
    storeName: "示例门店1",
    storeCode: "DEMO-01",
    shiftName: "总部巡店"
  },
  manager: {
    name: "李店长",
    phone: "13800002001",
    roleKey: "manager",
    role: "店长",
    storeName: "示例门店1",
    storeCode: "DEMO-01",
    shiftName: "门店早班"
  },
  staff: {
    name: "张收银",
    phone: "13800003001",
    roleKey: "staff",
    role: "员工",
    storeName: "示例门店1",
    storeCode: "DEMO-01",
    shiftName: "早班"
  },
  customerOwner: {
    name: "示例老板",
    phone: "13800000000",
    roleKey: "owner",
    role: "老板",
    storeName: "示例门店1",
    storeCode: "DEMO-01",
    shiftName: "模板演示"
  }
};

function cloneArray(list) {
  return (list || []).slice();
}

function getRolePreset(roleKey) {
  roleKey = normalizeRoleKey(roleKey);
  return rolePresets.find(function(item) {
    return item.key === roleKey;
  }) || rolePresets[2];
}

function normalizeRoleKey(roleKey) {
  switch (String(roleKey || "").trim().toLowerCase()) {
    case "owner":
      return "owner";
    case "manager":
      return "manager";
    case "cashier":
    case "clerk":
    case "staff":
      return "staff";
    default:
      return "staff";
  }
}

function getDefaultProfile() {
  return {
    id: "",
    name: "",
    phone: "",
    username: "",
    roleKey: "staff",
    role: "员工",
    permissions: [],
    storeId: "",
    storeName: "",
    storeCode: "",
    shiftName: "",
    token: "",
    loggedIn: false,
    createdAt: ""
  };
}

function getProfile() {
  return Object.assign(getDefaultProfile(), wx.getStorageSync(USER_KEY) || {});
}

function saveProfile(profile) {
  const rolePreset = getRolePreset(profile.roleKey || "");
  const nextProfile = Object.assign(getDefaultProfile(), profile, {
    id: profile.id || `U${Date.now()}`,
    roleKey: rolePreset.key,
    role: profile.role || rolePreset.label,
    permissions: cloneArray(profile.permissions && profile.permissions.length ? profile.permissions : rolePreset.permissions),
    loggedIn: true,
    createdAt: profile.createdAt || new Date().toISOString()
  });
  wx.setStorageSync(USER_KEY, nextProfile);
  return nextProfile;
}

function clearProfile() {
  const profile = getProfile();
  clearScopedBusinessData(profile);
  clearScopedBusinessData({});
  clearLegacyBusinessData();
  wx.removeStorageSync(USER_KEY);
}

function isDevtoolsRuntime() {
  try {
    const deviceInfo = wx.getDeviceInfo ? wx.getDeviceInfo() : {};
    const systemInfo = wx.getSystemInfoSync && !wx.getDeviceInfo ? wx.getSystemInfoSync() : {};
    const platform = String(deviceInfo.platform || systemInfo.platform || "").toLowerCase();
    return platform === "devtools" || platform === "mac" || platform === "windows";
  } catch (error) {
    return false;
  }
}

function pickDevtoolAccount(roleKey) {
  return devtoolAccounts[normalizeRoleKey(roleKey)] || devtoolAccounts.staff;
}

function loginDevtoolsAccount(profile) {
  if (!isDevtoolsRuntime()) {
    return Promise.reject(new Error("开发者工具联调登录只能在微信开发者工具内使用。"));
  }

  const devtoolsProfile = profile && profile.name ? profile : getDevtoolsOwnerProfile();
  const account = pickDevtoolAccount(devtoolsProfile.roleKey || "owner");
  return loginWithPassword(account.username, account.password);
}

function resolveStoreForProfile(baseProfile, user) {
  return request({
    endpoint: appConfig.endpoints.listStores
  }).then(function(payload) {
    const data = payload && payload.data ? payload.data : payload;
    const stores = Array.isArray(data) ? data : (data.items || []);
    const firstStore = stores[0] || {};
    const storeIds = user.storeIds || user.storeIDs || [];
    return saveProfile(Object.assign({}, baseProfile, {
      storeId: firstStore.id || storeIds[0] || baseProfile.storeId || appConfig.defaultStoreId,
      storeName: user.dataScope === "org_all" ? "全部门店" : (firstStore.name || baseProfile.storeName || appConfig.storeName),
      storeCode: firstStore.code || baseProfile.storeCode || appConfig.defaultStoreCode,
      visibleStores: stores
    }));
  }).catch(function() {
    return baseProfile;
  });
}

function buildProfileFromLoginResult(data) {
  const user = data.user || {};
  const rolePreset = getRolePreset(user.roleCode || user.roleKey || data.roleKey || "");
  const storeIds = user.storeIds || user.storeIDs || [];
  return {
    id: user.id || data.userId || data.id || "",
    username: user.username || data.username || "",
    name: user.displayName || user.name || data.name || rolePreset.label,
    phone: user.phone || data.phone || "",
    token: data.token || "",
    expiresAt: data.expiresAt || "",
    roleKey: rolePreset.key,
    role: user.roleName || data.roleName || rolePreset.label,
    permissions: user.permissions || data.permissions || rolePreset.permissions,
    dataScope: user.dataScope || data.dataScope || "",
    storeId: storeIds[0] || data.storeId || appConfig.defaultStoreId,
    storeName: user.dataScope === "org_all" ? "全部门店" : (data.storeName || appConfig.storeName),
    storeCode: data.storeCode || appConfig.defaultStoreCode
  };
}

function loginWithPassword(username, password) {
  const trimmedUsername = String(username || "").trim();
  if (!trimmedUsername || !password) {
    return Promise.reject(new Error("请输入账号和密码。"));
  }

  return request({
    endpoint: appConfig.endpoints.passwordLogin,
    method: "POST",
    data: {
      username: trimmedUsername,
      password: String(password || "")
    }
  }).then(function(payload) {
    const data = payload && payload.data ? payload.data : payload;
    const baseProfile = saveProfile(buildProfileFromLoginResult(data || {}));
    return resolveStoreForProfile(baseProfile, data.user || {});
  });
}

function loginOnline(profile, options) {
  const loginProfile = Object.assign(getDefaultProfile(), profile || {});
  const loginOptions = options || {};

  if (appConfig.mode === "offline") {
    return Promise.resolve(saveProfile(loginProfile));
  }

  return new Promise((resolve, reject) => {
    wx.login({
      success(loginResult) {
        if (!loginResult.code) {
          reject({ type: "WX_LOGIN_ERROR", error: loginResult });
          return;
        }

        const phoneAuthPayload = {};
        ["phoneCode", "phoneEncryptedData", "phoneIv"].forEach(function(key) {
          if (loginOptions[key]) {
            phoneAuthPayload[key] = loginOptions[key];
          }
        });

        request({
          endpoint: appConfig.endpoints.login,
          method: "POST",
          data: Object.assign({
            code: loginResult.code,
            profile: loginProfile
          }, phoneAuthPayload)
        }).then((payload) => {
          const data = payload && payload.data ? payload.data : payload;
          const rolePreset = getRolePreset(data.roleKey || loginProfile.roleKey || "");
          resolve(saveProfile(Object.assign({}, loginProfile, {
            id: data.userId || data.id || loginProfile.id,
            name: data.name || data.displayName || loginProfile.name,
            phone: data.phone || loginProfile.phone,
            token: data.token || loginProfile.token || "",
            roleKey: rolePreset.key,
            role: data.roleName || loginProfile.role || rolePreset.label,
            permissions: data.permissions || rolePreset.permissions,
            storeId: data.storeId || loginProfile.storeId || appConfig.defaultStoreId,
            storeName: data.storeName || loginProfile.storeName,
            storeCode: data.storeCode || loginProfile.storeCode
          })));
        }).catch(reject);
      },
      fail(error) {
        reject({ type: "WX_LOGIN_ERROR", error });
      }
    });
  });
}

function loginOnlineWithPhoneCode(profile, phoneCodeOrPayload) {
  const options = typeof phoneCodeOrPayload === "object"
    ? phoneCodeOrPayload
    : { phoneCode: phoneCodeOrPayload };
  return loginOnline(profile, options);
}

function isMiniAppBindingRequired(error) {
  const response = error && error.response ? error.response : {};
  const message = String(response.message || response.msg || error && error.message || "");
  return error && error.statusCode === 403 && (response.code === 40304 || /not bound|未绑定/i.test(message));
}

function isLoggedIn() {
  return Boolean(getProfile().loggedIn);
}

function hasPermission(code) {
  return getProfile().permissions.indexOf(code) > -1;
}

function canAccessFeature(featureKey) {
  const profile = getProfile();
  const rules = {
    cashier: ["cashier.use", "cashier.order.create", "cashier.order.read"],
    recycle: ["recycle.use", "recycle.order.draft", "recycle.order.confirm", "recycle.order.read"],
    orders: ["orders.view", "cashier.order.read", "recycle.order.read"],
    members: ["members.view", "member.read"],
    products: ["products.view", "catalog.product.read"],
    settings: ["settings.view", "settings.read"],
    "settings.manage": ["settings.manage", "settings.read"]
  };
  if (rules[featureKey] && !profile.loggedIn) {
    return false;
  }
  if (featureKey === "settings") {
    return rules.settings.some(function(code) {
      return hasPermission(code);
    }) || hasPermission("settings.manage");
  }
  const codes = rules[featureKey];
  if (!codes) return true;
  return codes.some(function(code) {
    return hasPermission(code);
  });
}

function getRolePresets() {
  return rolePresets.map(function(item) {
    return {
      key: item.key,
      label: item.label
    };
  });
}

function getQuickLoginProfiles() {
  return ["owner", "manager", "staff"].map(function(roleKey) {
    return Object.assign({}, getDefaultProfile(), quickLoginProfiles[roleKey], {
      storeId: appConfig.defaultStoreId,
      permissions: cloneArray(getRolePreset(roleKey).permissions)
    });
  });
}

function getDevtoolsOwnerProfile(ownerKey) {
  const profileKey = ownerKey === "customerOwner" ? "customerOwner" : "owner";
  return Object.assign({}, getDefaultProfile(), quickLoginProfiles[profileKey], {
    storeId: appConfig.defaultStoreId,
    permissions: cloneArray(getRolePreset("owner").permissions)
  });
}

function logoutOnline() {
  const profile = getProfile();
  if (!profile.token || appConfig.mode === "offline") {
    clearProfile();
    return Promise.resolve({ loggedOut: true });
  }
  return request({
    endpoint: appConfig.endpoints.logout,
    method: "POST"
  }).catch(function() {
    return { loggedOut: false };
  }).then(function(result) {
    clearProfile();
    return result;
  });
}

module.exports = {
  getDefaultProfile,
  getProfile,
  saveProfile,
  loginWithPassword,
  loginOnline,
  loginOnlineWithPhoneCode,
  loginDevtoolsAccount,
  isMiniAppBindingRequired,
  isDevtoolsRuntime,
  clearProfile,
  logoutOnline,
  isLoggedIn,
  hasPermission,
  canAccessFeature,
  getRolePresets,
  getQuickLoginProfiles,
  getDevtoolsOwnerProfile
};
