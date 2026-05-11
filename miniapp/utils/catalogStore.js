const MEMBER_KEY = "gr_member_records";
const PRODUCT_KEY = "gr_product_records";
const { appConfig, request } = require("./apiClient");

function deepCopy(value) {
  return JSON.parse(JSON.stringify(value));
}

function getMembers() {
  return wx.getStorageSync(MEMBER_KEY) || [];
}

function saveMembers(members) {
  wx.setStorageSync(MEMBER_KEY, members);
}

function getProducts() {
  return wx.getStorageSync(PRODUCT_KEY) || [];
}

function saveProducts(products) {
  wx.setStorageSync(PRODUCT_KEY, products);
}

function seedCatalog() {
  if (!getMembers().length) {
    saveMembers([
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
        managerName: "廖店长",
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
        managerName: "廖店长",
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
    ]);
  }

  if (!getProducts().length) {
    saveProducts([
      {
        id: "product-001",
        name: "足金手镯标准款",
        sku: "GJG-SZ-001",
        category: "金饰",
        purity: "足金999",
        benchPrice: 742,
        retailPrice: 786,
        gramWeight: 12.68,
        status: "active",
        inventory: 9,
        stores: ["华强北体验店", "福田旗舰店"],
        tags: ["主推", "高转化"],
        recommendedScene: "适合门店当面议价、快速复秤录单。",
        quoteLeadTime: "30 秒内可完成报价"
      },
      {
        id: "product-002",
        name: "18K 项链回收模板",
        sku: "GJG-KG-018",
        category: "K金",
        purity: "18K",
        benchPrice: 558,
        retailPrice: 612,
        gramWeight: 8.92,
        status: "active",
        inventory: 4,
        stores: ["华强北体验店"],
        tags: ["K金", "常见回收"],
        recommendedScene: "适合员工快速带入常见 K 金参数。",
        quoteLeadTime: "60 秒内完成复检"
      },
      {
        id: "product-003",
        name: "足金金条 50g",
        sku: "GJG-BAR-050",
        category: "金条",
        purity: "足金9999",
        benchPrice: 748,
        retailPrice: 802,
        gramWeight: 50,
        status: "draft",
        inventory: 2,
        stores: ["福田旗舰店"],
        tags: ["大额", "待复核"],
        recommendedScene: "大额回收建议店长复核并走双人确认。",
        quoteLeadTime: "需二次审批"
      },
      {
        id: "product-004",
        name: "旧金料混合回收包",
        sku: "GJG-OLD-888",
        category: "旧金料",
        purity: "22K",
        benchPrice: 680,
        retailPrice: 0,
        gramWeight: 0,
        status: "disabled",
        inventory: 0,
        stores: ["华强北体验店"],
        tags: ["临时下架"],
        recommendedScene: "等待总部调整损耗规则后重新启用。",
        quoteLeadTime: "暂不建议使用"
      }
    ]);
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

function listMembersOnline() {
  if (appConfig.mode === "mock") {
    return Promise.resolve(deepCopy(getMembers()));
  }

  return request({
    endpoint: appConfig.endpoints.listMembers
  }).then(function(payload) {
    const data = payload && payload.data ? payload.data : payload;
    const list = Array.isArray(data) ? data : (data.items || []);
    saveMembers(list);
    return deepCopy(list);
  }).catch(function() {
    return deepCopy(getMembers());
  });
}

function listProductsOnline() {
  if (appConfig.mode === "mock") {
    return Promise.resolve(deepCopy(getProducts()));
  }

  return request({
    endpoint: appConfig.endpoints.listProducts
  }).then(function(payload) {
    const data = payload && payload.data ? payload.data : payload;
    const list = Array.isArray(data) ? data : (data.items || []);
    saveProducts(list);
    return deepCopy(list);
  }).catch(function() {
    return deepCopy(getProducts());
  });
}

module.exports = {
  seedCatalog,
  getMembers,
  getProducts,
  getMemberStats,
  getProductStats,
  listMembersOnline,
  listProductsOnline
};
