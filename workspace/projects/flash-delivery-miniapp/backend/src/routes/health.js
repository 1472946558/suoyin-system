"use strict";

const { ok } = require("../lib/response");

module.exports = async function healthRoutes(fastify) {
  async function handleHealth() {
    return ok({
      status: "ok",
      service: "flash-delivery-miniapp-backend",
      appMode: fastify.appConfig.appMode,
      timestamp: new Date().toISOString()
    });
  }

  fastify.get("/health", handleHealth);
  fastify.get("/healthz", handleHealth);
};
