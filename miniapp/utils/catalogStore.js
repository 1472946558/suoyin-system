/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: catalogStore.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-10
 */

const MEMBER_KEY = "gr_member_records";
const MEMBER_META_KEY = "gr_member_records_meta";
const PRODUCT_KEY = "gr_product_records";
const PRODUCT_META_KEY = "gr_product_records_meta";
const { appConfig, getResolvedAppConfig } = require("./config");
const { request, formatRequestError } = require("./apiClient");
const { getScopedStorageSync, setScopedStorageSync } = require("./sessionStorage");

function deepCopy(value) {
  return JSON.parse(JSON.stringify(value));
}

function canUseStorage() {
  return typeof wx !== "undefined" && typeof wx.getStorageSync === "function";
}

function shouldUseLocalCatalogData() {
  return appConfig.mode === "offline";
}

function pad(value) {
  return value < 10 ? `0${value}` : String(value);
}

function formatTimestamp(date) {
  const current = date || new Date();
  return [
    current.getFullYear(),
    pad(current.getMonth() + 1),
    pad(current.getDate())
  ].join("-") + " " + [
    pad(current.getHours()),
    pad(current.getMinutes()),
    pad(current.getSeconds())
  ].join(":");
}

function getStorageArray(key) {
  if (!canUseStorage()) {
    return [];
  }
  const stored = getScopedStorageSync(key, []);
  return Array.isArray(stored) ? stored : [];
}

function saveStorageArray(key, list) {
  if (canUseStorage()) {
    setScopedStorageSync(key, Array.isArray(list) ? list : []);
  }
}

function getStorageMeta(key) {
  if (!canUseStorage()) {
    return {};
  }
  const stored = getScopedStorageSync(key, {});
  return stored && typeof stored === "object" && !Array.isArray(stored) ? stored : {};
}

function saveStorageMeta(key, meta) {
  if (canUseStorage()) {
    setScopedStorageSync(key, meta || {});
  }
}

function getMembers() {
  return getStorageArray(MEMBER_KEY);
}

function saveMembers(members) {
  saveStorageArray(MEMBER_KEY, members);
}

function getProducts() {
  return getStorageArray(PRODUCT_KEY);
}

function saveProducts(products) {
  saveStorageArray(PRODUCT_KEY, products);
}

function getMembersMeta() {
  return getStorageMeta(MEMBER_META_KEY);
}

function saveMembersMeta(meta) {
  saveStorageMeta(MEMBER_META_KEY, meta);
}

function getProductsMeta() {
  return getStorageMeta(PRODUCT_META_KEY);
}

function saveProductsMeta(meta) {
  saveStorageMeta(PRODUCT_META_KEY, meta);
}

function createCatalogId(prefix) {
  return `${prefix}-${Date.now()}-${Math.floor(Math.random() * 1000)}`;
}

function buildMemberSeed() {
  return [
    {
      id: "member-001",
      name: "张女士",
      phone: "13800002026",
      level: "VIP",
      status: "active",
      totalOrders: 6,
      totalRecycleAmount: 28640,
      lastVisitAt: "2026-05-10 16:20",
      preferredPurity: "足金999",
      sourceChannel: "熟客复购",
      managerName: "本地测试店长",
      idVerified: true,
      tags: ["高净值", "可回访", "到店复秤"],
      notes: "偏好工作日下午到店，成交前会要求复秤拍照。"
    },
    {
      id: "member-002",
      name: "陈先生",
      phone: "13900008866",
      level: "潜力客户",
      status: "follow_up",
      totalOrders: 2,
      totalRecycleAmount: 9240,
      lastVisitAt: "2026-05-08 11:15",
      preferredPurity: "18K",
      sourceChannel: "企业回访",
      managerName: "本地测试店长",
      idVerified: true,
      tags: ["银行卡结算", "待回访"],
      notes: "对到账时效敏感，偏好银行卡。"
    },
    {
      id: "member-003",
      name: "王阿姨",
      phone: "13700005588",
      level: "普通会员",
      status: "sleeping",
      totalOrders: 1,
      totalRecycleAmount: 3820,
      lastVisitAt: "2026-04-26 09:30",
      preferredPurity: "22K",
      sourceChannel: "渠道转介绍",
      managerName: "陈店员",
      idVerified: false,
      tags: ["老客家属", "待身份证补录"],
      notes: "上次成交后暂无二次回访记录。"
    }
  ];
}

function buildProductSeed() {
  return [
    {
      id: "product-001",
      name: "足金手镯标准款",
      sku: "GJG-SZ-001",
      category: "金饰",
      categoryTab: "黄金饰品",
      imageUrl: "/assets/ui/price.png",
      purity: "足金999",
      benchPrice: 742,
      retailPrice: 786,
      gramWeight: 12.68,
      status: "active",
      inventory: 9,
      stockStatus: "normal",
      stores: ["本地测试门店1", "本地测试门店2"],
      tags: ["主推", "高转化"],
      recommendedScene: "适合门店当面议价、快速复秤录单。",
      quoteLeadTime: "30 秒内可完成报价"
    },
    {
      id: "product-002",
      name: "18K 项链",
      sku: "GJG-KG-018",
      category: "K金",
      categoryTab: "K金",
      imageUrl: "/assets/ui/weight.png",
      purity: "18K",
      benchPrice: 558,
      retailPrice: 612,
      gramWeight: 8.92,
      status: "active",
      inventory: 4,
      stockStatus: "low",
      stores: ["本地测试门店1"],
      tags: ["K金", "常见回收"],
      recommendedScene: "适合门店快速带入常见 K 金参数。",
      quoteLeadTime: "60 秒内完成复检"
    },
    {
      id: "product-003",
      name: "足金金条 50g",
      sku: "GJG-BAR-050",
      category: "金条",
      categoryTab: "投资金",
      imageUrl: "/assets/ui/document.png",
      purity: "足金9999",
      benchPrice: 748,
      retailPrice: 802,
      gramWeight: 50,
      status: "draft",
      inventory: 2,
      stockStatus: "review",
      stores: ["本地测试门店2"],
      tags: ["大额", "待复核"],
      recommendedScene: "大额回收建议店长复核并走双人确认。",
      quoteLeadTime: "需二次审批"
    },
    {
      id: "product-004",
      name: "旧金料混合回收包",
      sku: "GJG-OLD-888",
      category: "旧金料",
      categoryTab: "其他",
      imageUrl: "/assets/ui/list.png",
      purity: "22K",
      benchPrice: 680,
      retailPrice: 0,
      gramWeight: 0,
      status: "disabled",
      inventory: 0,
      stockStatus: "disabled",
      stores: ["本地测试门店1"],
      tags: ["临时下架"],
      recommendedScene: "等待总部调整损耗规则后重新启用。",
      quoteLeadTime: "暂不建议使用"
    }
  ];
}

function seedCatalog() {
  if (!getMembers().length) {
    const members = buildMemberSeed();
    saveMembers(members);
    saveMembersMeta({
      sourceType: "local",
      updatedAt: formatTimestamp(),
      itemCount: members.length,
      lastError: "",
      lastErrorAt: ""
    });
  }

  if (!getProducts().length) {
    const products = buildProductSeed();
    saveProducts(products);
    saveProductsMeta({
      sourceType: "local",
      updatedAt: formatTimestamp(),
      itemCount: products.length,
      lastError: "",
      lastErrorAt: ""
    });
  }
}

function getMemberStats(members) {
  const list = members || getMembers();
  return {
    total: list.length,
    vipCount: list.filter(function(item) {
      return item.level === "VIP";
    }).length,
    activeCount: list.filter(function(item) {
      return item.status === "active";
    }).length,
    followUpCount: list.filter(function(item) {
      return item.status === "follow_up";
    }).length
  };
}

function getProductStats(products) {
  const list = products || getProducts();
  return {
    total: list.length,
    activeCount: list.filter(function(item) {
      return item.status === "active";
    }).length,
    lowStockCount: list.filter(function(item) {
      return item.inventory > 0 && item.inventory <= 4;
    }).length,
    draftCount: list.filter(function(item) {
      return item.status === "draft";
    }).length
  };
}

function normalizeCatalogList(payload) {
  const data = payload && payload.data ? payload.data : payload;
  if (Array.isArray(data)) {
    return data;
  }
  if (data && Array.isArray(data.items)) {
    return data.items;
  }
  return [];
}

function normalizeDetailPayload(payload) {
  return payload && payload.data ? payload.data : payload;
}

function buildSourceMeta(sourceType, extra) {
  return Object.assign({
    sourceType,
    mode: appConfig.mode,
    apiBaseUrl: getResolvedAppConfig().apiBaseUrl,
    updatedAt: "",
    lastError: "",
    lastErrorAt: "",
    itemCount: 0,
    legacyDataBlocked: false
  }, extra || {});
}

function isTrustedApiCache(meta, list) {
  return !!(meta && meta.sourceType === "api" && Array.isArray(list));
}

function saveApiCatalogSnapshot(list, saveList, saveMeta) {
  saveList(list);
  saveMeta({
    sourceType: "api",
    updatedAt: formatTimestamp(),
    itemCount: list.length,
    lastError: "",
    lastErrorAt: ""
  });
}

function createCatalogLoader(options) {
  const {
    endpoint,
    getList,
    saveList,
    getMeta,
    saveMeta
  } = options;

  return function loadCatalog() {
    const cachedItems = getList();
    const cachedMeta = getMeta();

    if (shouldUseLocalCatalogData()) {
      seedCatalog();
      const localItems = getList();
      const localMeta = getMeta();
      return Promise.resolve({
        items: deepCopy(localItems),
        meta: buildSourceMeta("local", {
          updatedAt: localMeta.updatedAt || formatTimestamp(),
          lastError: "",
          lastErrorAt: "",
          itemCount: localItems.length
        })
      });
    }

    return request({ endpoint }).then(function(payload) {
      const list = normalizeCatalogList(payload);
      saveApiCatalogSnapshot(list, saveList, saveMeta);
      return {
        items: deepCopy(list),
        meta: buildSourceMeta("api", {
          updatedAt: formatTimestamp(),
          itemCount: list.length
        })
      };
    }).catch(function(error) {
      const errorMessage = formatRequestError(error);
      const errorTime = formatTimestamp();
      saveMeta(Object.assign({}, cachedMeta, {
        lastError: errorMessage,
        lastErrorAt: errorTime
      }));

      if (isTrustedApiCache(cachedMeta, cachedItems)) {
        return {
          items: deepCopy(cachedItems),
          meta: buildSourceMeta("cache", {
            updatedAt: cachedMeta.updatedAt || "",
            lastError: errorMessage,
            lastErrorAt: errorTime,
            itemCount: cachedItems.length
          })
        };
      }

      return {
        items: [],
        meta: buildSourceMeta("error", {
          updatedAt: "",
          lastError: errorMessage,
          lastErrorAt: errorTime,
          itemCount: 0,
          legacyDataBlocked: Array.isArray(cachedItems) && cachedItems.length > 0
        })
      };
    });
  };
}

function getMemberByIdOnline(id) {
  if (shouldUseLocalCatalogData()) {
    seedCatalog();
    return Promise.resolve(deepCopy(getMembers().find(function(item) {
      return item.id === id;
    }) || null));
  }

  return request({
    endpoint: appConfig.endpoints.memberDetail,
    params: { id }
  }).then(function(payload) {
    const member = normalizeDetailPayload(payload);
    if (!member || !member.id) {
      return null;
    }
    const list = [member].concat(getMembers().filter(function(item) {
      return item.id !== member.id;
    }));
    saveApiCatalogSnapshot(list, saveMembers, saveMembersMeta);
    return deepCopy(member);
  });
}

function getProductByIdOnline(id) {
  if (shouldUseLocalCatalogData()) {
    seedCatalog();
    return Promise.resolve(deepCopy(getProducts().find(function(item) {
      return item.id === id;
    }) || null));
  }

  return request({
    endpoint: appConfig.endpoints.productDetail,
    params: { id }
  }).then(function(payload) {
    const product = normalizeDetailPayload(payload);
    if (!product || !product.id) {
      return null;
    }
    const list = [product].concat(getProducts().filter(function(item) {
      return item.id !== product.id;
    }));
    saveApiCatalogSnapshot(list, saveProducts, saveProductsMeta);
    return deepCopy(product);
  });
}

function normalizeTags(value) {
  if (Array.isArray(value)) {
    return value.filter(Boolean);
  }
  return String(value || "")
    .split(/[,\s，、]+/)
    .map(function(item) { return item.trim(); })
    .filter(Boolean);
}

function saveMemberRecord(member) {
  if (shouldUseLocalCatalogData()) {
    const current = getMembers();
    const id = member.id || createCatalogId("member");
    const nextMember = Object.assign({
      id,
      name: "",
      phone: "",
      level: "普通会员",
      status: "active",
      totalOrders: 0,
      totalRecycleAmount: 0,
      lastVisitAt: formatTimestamp(),
      preferredPurity: "足金999",
      sourceChannel: "门店登记",
      managerName: "",
      idVerified: false,
      tags: [],
      notes: ""
    }, member, {
      id,
      tags: normalizeTags(member.tags || member.tagsText)
    });
    const list = [nextMember].concat(current.filter(function(item) {
      return item.id !== id;
    }));
    saveMembers(list);
    saveMembersMeta({
      sourceType: "local",
      updatedAt: formatTimestamp(),
      itemCount: list.length,
      lastError: "",
      lastErrorAt: ""
    });
    return Promise.resolve(deepCopy(nextMember));
  }

  return request({
    endpoint: member.id ? appConfig.endpoints.memberDetail : appConfig.endpoints.createMember,
    method: member.id ? "PUT" : "POST",
    params: member.id ? { id: member.id } : {},
    data: {
      name: String(member.name || "").trim(),
      phone: String(member.phone || "").trim(),
      level: String(member.level || "普通会员").trim(),
      status: String(member.status || "active").trim(),
      preferredPurity: String(member.preferredPurity || "").trim(),
      sourceChannel: String(member.sourceChannel || "").trim(),
      managerName: String(member.managerName || "").trim(),
      idVerified: !!member.idVerified,
      tags: normalizeTags(member.tags || member.tagsText),
      notes: String(member.notes || "").trim()
    }
  }).then(function(payload) {
    const savedMember = normalizeDetailPayload(payload);
    const list = [savedMember].concat(getMembers().filter(function(item) {
      return item.id !== savedMember.id;
    }));
    saveApiCatalogSnapshot(list, saveMembers, saveMembersMeta);
    return deepCopy(savedMember);
  });
}

function saveProductRecord(product) {
  const inventory = Number(product.inventory || 0);
  const status = String(product.status || "active").trim();
  const stockStatus = status === "disabled" ? "disabled" : (inventory > 0 && inventory <= 4 ? "low" : "normal");

  if (shouldUseLocalCatalogData()) {
    const current = getProducts();
    const id = product.id || createCatalogId("product");
    const nextProduct = Object.assign({
      id,
      name: "",
      sku: "",
      category: "金饰",
      categoryTab: "黄金饰品",
      imageUrl: "/assets/ui/price.png",
      purity: "足金999",
      benchPrice: 0,
      retailPrice: 0,
      gramWeight: 0,
      status,
      inventory,
      stockStatus,
      stores: [appConfig.storeName],
      tags: [],
      recommendedScene: "适合门店快速带入录单。",
      quoteLeadTime: "当场可报价"
    }, product, {
      id,
      benchPrice: Number(product.benchPrice || 0),
      retailPrice: Number(product.retailPrice || 0),
      gramWeight: Number(product.gramWeight || 0),
      inventory,
      stockStatus,
      tags: normalizeTags(product.tags || product.tagsText)
    });
    const list = [nextProduct].concat(current.filter(function(item) {
      return item.id !== id;
    }));
    saveProducts(list);
    saveProductsMeta({
      sourceType: "local",
      updatedAt: formatTimestamp(),
      itemCount: list.length,
      lastError: "",
      lastErrorAt: ""
    });
    return Promise.resolve(deepCopy(nextProduct));
  }

  return request({
    endpoint: product.id ? appConfig.endpoints.productDetail : appConfig.endpoints.createProduct,
    method: product.id ? "PUT" : "POST",
    params: product.id ? { id: product.id } : {},
    data: {
      name: String(product.name || "").trim(),
      sku: String(product.sku || "").trim(),
      category: String(product.category || "").trim(),
      categoryTab: String(product.categoryTab || product.category || "").trim(),
      imageUrl: String(product.imageUrl || "").trim(),
      purity: String(product.purity || "").trim(),
      benchPrice: Number(product.benchPrice || 0),
      retailPrice: Number(product.retailPrice || 0),
      gramWeight: Number(product.gramWeight || 0),
      status,
      inventory,
      stockStatus,
      tags: normalizeTags(product.tags || product.tagsText),
      recommendedScene: String(product.recommendedScene || "").trim(),
      quoteLeadTime: String(product.quoteLeadTime || "").trim()
    }
  }).then(function(payload) {
    const savedProduct = normalizeDetailPayload(payload);
    const list = [savedProduct].concat(getProducts().filter(function(item) {
      return item.id !== savedProduct.id;
    }));
    saveApiCatalogSnapshot(list, saveProducts, saveProductsMeta);
    return deepCopy(savedProduct);
  });
}

const listMembersOnline = createCatalogLoader({
  endpoint: appConfig.endpoints.listMembers,
  getList: getMembers,
  saveList: saveMembers,
  getMeta: getMembersMeta,
  saveMeta: saveMembersMeta
});

const listProductsOnline = createCatalogLoader({
  endpoint: appConfig.endpoints.listProducts,
  getList: getProducts,
  saveList: saveProducts,
  getMeta: getProductsMeta,
  saveMeta: saveProductsMeta
});

module.exports = {
  seedCatalog,
  getMembers,
  getProducts,
  getMemberByIdOnline,
  getProductByIdOnline,
  saveMemberRecord,
  saveProductRecord,
  getMemberStats,
  getProductStats,
  listMembersOnline,
  listProductsOnline
};
