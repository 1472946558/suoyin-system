/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: smoke-test.mjs
 * 功能描述: 测试代码
 * 作者: 廖心慈
 * 创建日期: 2026-05-11
 */

const apiBase = process.env.API_BASE || "http://127.0.0.1:18082";

async function request(path, options = {}) {
  const response = await fetch(`${apiBase}${path}`, {
    headers: {
      "content-type": "application/json",
      ...(options.headers || {}),
    },
    ...options,
  });

  const text = await response.text();
  const payload = text ? JSON.parse(text) : {};
  if (!response.ok || payload.code !== 0) {
    throw new Error(`${path} failed: ${response.status} ${payload.message || "unknown error"}`);
  }
  return payload.data;
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

async function main() {
  const health = await request("/health", { headers: {} });
  assert(health.service === "gold-recycle-miniapp-backend", "unexpected health payload");

  const adminLogin = await request("/api/admin/login", {
    method: "POST",
    body: JSON.stringify({ username: "boss", password: "Boss123!" }),
  });
  assert(adminLogin.token, "admin token missing");

  const bootstrap = await request("/api/admin/bootstrap", {
    headers: {
      authorization: `Bearer ${adminLogin.token}`,
    },
  });
  assert(Array.isArray(bootstrap.stores) && bootstrap.stores.length > 0, "bootstrap stores missing");

  const miniLogin = await request("/api/v1/auth/wechat-login", {
    method: "POST",
    body: JSON.stringify({
      code: "mock-wechat-code",
      profile: {
        name: "李店长",
        roleKey: "manager",
        storeName: "南山旗舰店",
        storeCode: "SZ-NS",
      },
    }),
  });
  assert(miniLogin.token, "miniapp token missing");

  const authHeader = {
    authorization: `Bearer ${miniLogin.token}`,
  };

  const members = await request("/api/v1/members", { headers: authHeader });
  assert(Array.isArray(members.items) && members.items.length > 0, "members list missing");

  const products = await request("/api/v1/products", { headers: authHeader });
  assert(Array.isArray(products.items) && products.items.length > 0, "products list missing");

  const cashier = await request("/api/v1/cashier/orders", {
    method: "POST",
    headers: authHeader,
    body: JSON.stringify({
      storeId: miniLogin.storeId,
      customerName: "收银客户",
      customerPhone: "13800138001",
      paymentMethod: "cash",
      remark: "smoke test cashier order",
      items: [
        {
          name: "足金手镯标准款",
          quantity: 1,
          unitPrice: 1288,
        },
      ],
    }),
  });
  assert(cashier.id, "cashier order create failed");

  const recycleDraft = await request("/api/v1/recycle/orders", {
    method: "POST",
    headers: authHeader,
    body: JSON.stringify({
      storeId: miniLogin.storeId,
      customerName: "烟测客户",
      customerPhone: "13800138000",
      estimatedAmount: 3200,
      items: [
        {
          category: "金饰",
          purity: "足金999",
          weightGram: 5.2,
        },
      ],
      attachmentUrls: [
        "https://mock.example.com/test-1.jpg",
        "https://mock.example.com/test-2.jpg",
        "https://mock.example.com/test-3.jpg",
      ],
      remark: "smoke draft",
    }),
  });
  assert(recycleDraft.id, "recycle draft create failed");

  const recycleConfirmed = await request(`/api/v1/recycle/orders/${recycleDraft.id}/confirm`, {
    method: "POST",
    headers: authHeader,
    body: JSON.stringify({
      confirmedAmount: 3180,
      attachmentUrls: recycleDraft.attachmentUrls,
      remark: "smoke confirm",
    }),
  });
  assert(recycleConfirmed.status === "confirmed", "recycle confirm failed");

  const recycleDetail = await request(`/api/v1/recycle/orders/${recycleDraft.id}`, {
    headers: authHeader,
  });
  assert(recycleDetail.id === recycleDraft.id, "recycle detail lookup failed");

  console.log("SMOKE TEST PASSED");
  console.log(`health ok: ${health.serverTime}`);
  console.log(`admin stores: ${bootstrap.stores.length}`);
  console.log(`members: ${members.items.length}, products: ${products.items.length}`);
  console.log(`cashier order: ${cashier.orderNo}, recycle order: ${recycleConfirmed.orderNo}`);
}

main().catch((error) => {
  console.error("SMOKE TEST FAILED");
  console.error(error.message);
  process.exit(1);
});
