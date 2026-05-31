"use strict";

const { ok, fail, requireUser } = require("./_shared");

module.exports = async function profileRoutes(fastify) {
  fastify.get("/profile", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const profile = fastify.appStore.buildProfile(auth.user.id);
    if (!profile) {
      reply.code(404);
      return reply.send(fail("profile not found", 404));
    }

    return reply.send(ok(profile));
  });
};
