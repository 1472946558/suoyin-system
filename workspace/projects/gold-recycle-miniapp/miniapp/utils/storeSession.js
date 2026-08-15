/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: storeSession.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-10
 */

const { appConfig, request } = require("./apiClient");
const { getProfile, saveProfile } = require("./userStore");
const { clearAllSessionData } = require("./sessionStorage");

function normalizeStoreChoice(user, stores) {
  const list = Array.isArray(stores) ? stores : [];
  const allowedStoreIds = Array.isArray(user && user.storeIds) ? user.storeIds : [];
  const matched = list.find(function(store) {
    return allowedStoreIds.indexOf(store.id) > -1;
  });
  return matched || list[0] || null;
}

function buildUpdatedProfile(currentProfile, user, stores) {
  const list = Array.isArray(stores) ? stores : [];
  const selectedStore = normalizeStoreChoice(user, stores);
  const isOwner = user && user.dataScope === "org_all";
  return Object.assign({}, currentProfile, {
    id: user && user.id ? user.id : currentProfile.id,
    username: user && user.username ? user.username : currentProfile.username,
    name: user && user.displayName ? user.displayName : currentProfile.name,
    phone: user && user.phone ? user.phone : currentProfile.phone,
    roleKey: user && user.roleCode ? user.roleCode : currentProfile.roleKey,
    role: user && user.roleName ? user.roleName : currentProfile.role,
    permissions: user && Array.isArray(user.permissions) ? user.permissions : currentProfile.permissions,
    dataScope: user && user.dataScope ? user.dataScope : currentProfile.dataScope,
    storeId: selectedStore && selectedStore.id ? selectedStore.id : currentProfile.storeId,
    storeName: isOwner
      ? ((selectedStore && selectedStore.name) || currentProfile.storeName || "")
      : ((selectedStore && selectedStore.name) || currentProfile.storeName || ""),
    storeCode: selectedStore && selectedStore.code ? selectedStore.code : currentProfile.storeCode,
    visibleStores: list
  });
}

function refreshProfileStoreBinding() {
  const profile = getProfile();
  if (!profile.loggedIn || !profile.token || appConfig.mode === "offline") {
    return Promise.resolve(profile);
  }

  return Promise.all([
    request({ endpoint: appConfig.endpoints.me }),
    request({ endpoint: appConfig.endpoints.listStores })
  ]).then(function(results) {
    const userPayload = results[0] && results[0].data ? results[0].data : results[0];
    const storePayload = results[1] && results[1].data ? results[1].data : results[1];
    const stores = Array.isArray(storePayload) ? storePayload : (storePayload.items || []);
    const nextProfile = buildUpdatedProfile(profile, userPayload || {}, stores);
    return saveProfile(nextProfile);
  }).catch(function(error) {
    if (error && error.statusCode === 401) {
      clearAllSessionData();
      return getProfile();
    }
    return profile;
  });
}

module.exports = {
  refreshProfileStoreBinding
};
