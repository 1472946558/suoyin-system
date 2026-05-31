"use strict";

const path = require("path");
const dotenv = require("dotenv");

dotenv.config({
  path: process.env.DOTENV_CONFIG_PATH || path.resolve(__dirname, "../.env")
});

function numberFromEnv(name, fallback) {
  const value = process.env[name];
  if (!value) return fallback;
  const parsed = Number(value);
  return Number.isFinite(parsed) ? parsed : fallback;
}

const config = {
  appName: process.env.APP_NAME || "疾蜂急送后端",
  host: process.env.HOST || "127.0.0.1",
  port: numberFromEnv("PORT", 18092),
  appUrl: process.env.APP_URL || "http://127.0.0.1:18092",
  apiPrefix: process.env.API_PREFIX || "/api/v1",
  jwtSecret: process.env.JWT_SECRET || "flash-delivery-dev-secret",
  tenantId: process.env.TENANT_ID || "demo-tenant",
  appMode: process.env.APP_MODE || "memory",
  mysqlHost: process.env.MYSQL_HOST || "127.0.0.1",
  mysqlPort: numberFromEnv("MYSQL_PORT", 3306),
  mysqlUser: process.env.MYSQL_USER || "",
  mysqlPassword: process.env.MYSQL_PASSWORD || "",
  mysqlDatabase: process.env.MYSQL_DATABASE || "",
  corsOrigin: process.env.CORS_ORIGIN || "*"
};

function mysqlDsn() {
  const auth = `${encodeURIComponent(config.mysqlUser)}:${encodeURIComponent(config.mysqlPassword)}`;
  return `mysql://${auth}@${config.mysqlHost}:${config.mysqlPort}/${config.mysqlDatabase}`;
}

module.exports = {
  config,
  mysqlDsn
};
