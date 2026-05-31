"use strict";

const Fastify = require("fastify");
const cors = require("@fastify/cors");
const { config } = require("./config");
const { createStore } = require("./lib/store");

function buildApp() {
  const app = Fastify({
    logger: true
  });

  app.decorate("appConfig", config);
  app.decorate("appStore", createStore());

  app.register(cors, {
    origin: config.corsOrigin === "*" ? true : config.corsOrigin.split(",")
  });

  app.register(require("./routes/health"));
  app.register(require("./routes/pricing"), { prefix: config.apiPrefix });
  app.register(require("./routes/auth"), { prefix: config.apiPrefix });
  app.register(require("./routes/orders"), { prefix: config.apiPrefix });
  app.register(require("./routes/profile"), { prefix: config.apiPrefix });
  app.register(require("./routes/addresses"), { prefix: config.apiPrefix });
  app.register(require("./routes/payments"), { prefix: config.apiPrefix });
  app.register(require("./routes/support"), { prefix: config.apiPrefix });

  return app;
}

module.exports = {
  buildApp
};
