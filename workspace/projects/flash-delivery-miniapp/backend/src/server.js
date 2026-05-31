"use strict";

const { buildApp } = require("./app");
const { config } = require("./config");

async function start() {
  const app = buildApp();

  try {
    await app.listen({
      host: config.host,
      port: config.port
    });
  } catch (error) {
    app.log.error(error);
    process.exit(1);
  }
}

start();
