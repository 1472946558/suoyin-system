const { seedOrders } = require("./utils/orderStore");
const { seedCatalog } = require("./utils/catalogStore");
const { appConfig } = require("./utils/config");

App({
  globalData: {
    brandName: appConfig.appName,
    brandSlogan: appConfig.brandSlogan,
    storeName: appConfig.storeName,
    servicePhone: appConfig.servicePhone,
    city: appConfig.storeCity,
    apiBaseUrl: appConfig.apiBaseUrl,
    mode: appConfig.mode
  },

  onLaunch() {
    seedOrders();
    seedCatalog();
  }
});
