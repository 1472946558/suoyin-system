"use strict";

const { ok, requireUser } = require("./_shared");

module.exports = async function addressRoutes(fastify) {
  fastify.get("/addresses", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const items = fastify.appStore.getDefaultAddresses(auth.user.id);
    reply.send(ok({
      items
    }));
  });
};
