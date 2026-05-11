const USER_KEY = "gr_operator_profile";
const { appConfig, request } = require("./apiClient");

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
    key: "clerk",
    label: "员工",
    permissions: ["cashier.use", "recycle.use", "orders.view", "members.view", "products.view"]
  }
];

function cloneArray(list) {
  return (list || []).slice();
}

function getRolePreset(roleKey) {
  return rolePresets.find(function(item) {
    return item.key === roleKey;
  }) || rolePresets[2];
}

function getDefaultProfile() {
  return {
    id: "",
    name: "",
    phone: "",
    roleKey: "clerk",
    role: "店员",
    permissions: cloneArray(getRolePreset("clerk").permissions),
    storeId: appConfig.defaultStoreId,
    storeName: appConfig.storeName,
    storeCode: appConfig.defaultStoreCode,
    shiftName: "早班",
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
  wx.removeStorageSync(USER_KEY);
}

function loginOnline(profile) {
  if (appConfig.mode === "mock") {
    return Promise.resolve(saveProfile(profile));
  }

  return new Promise((resolve, reject) => {
    wx.login({
      success(loginResult) {
        if (!loginResult.code) {
          reject({ type: "WX_LOGIN_ERROR", error: loginResult });
          return;
        }

        request({
          endpoint: appConfig.endpoints.login,
          method: "POST",
          data: {
            code: loginResult.code,
            profile
          }
        }).then((payload) => {
          const data = payload && payload.data ? payload.data : payload;
          const rolePreset = getRolePreset(data.roleKey || profile.roleKey || "");
          resolve(saveProfile(Object.assign({}, profile, {
            id: data.userId || data.id || profile.id,
            token: data.token || profile.token || "",
            roleKey: data.roleKey || rolePreset.key,
            role: data.roleName || profile.role || rolePreset.label,
            permissions: data.permissions || rolePreset.permissions,
            storeId: data.storeId || profile.storeId || appConfig.defaultStoreId,
            storeName: data.storeName || profile.storeName,
            storeCode: data.storeCode || profile.storeCode
          })));
        }).catch(reject);
      },
      fail(error) {
        reject({ type: "WX_LOGIN_ERROR", error });
      }
    });
  });
}

function isLoggedIn() {
  return Boolean(getProfile().loggedIn);
}

function hasPermission(code) {
  return getProfile().permissions.indexOf(code) > -1;
}

function canAccessFeature(featureKey) {
  const rules = {
    cashier: "cashier.use",
    recycle: "recycle.use",
    orders: "orders.view",
    members: "members.view",
    products: "products.view",
    settings: "settings.view",
    "settings.manage": "settings.manage"
  };
  if (featureKey === "settings") {
    return hasPermission("settings.view") || hasPermission("settings.manage");
  }
  const code = rules[featureKey];
  if (!code) return true;
  return hasPermission(code);
}

function getRolePresets() {
  return rolePresets.map(function(item) {
    return {
      key: item.key,
      label: item.label
    };
  });
}

module.exports = {
  getDefaultProfile,
  getProfile,
  saveProfile,
  loginOnline,
  clearProfile,
  isLoggedIn,
  hasPermission,
  canAccessFeature,
  getRolePresets
};
