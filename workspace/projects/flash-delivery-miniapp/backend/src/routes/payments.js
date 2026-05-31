"use strict";

const { fail, ok, requireUser } = require("./_shared");

module.exports = async function paymentRoutes(fastify) {
  fastify.post("/payments/wechat", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const orderId = request.body?.orderId;
    if (!orderId) {
      reply.code(400);
      return reply.send(fail("orderId is required", 400));
    }

    const order = orderId ? fastify.appStore.getOrder(orderId) : null;

    if (!order) {
      reply.code(404).send(fail("order not found", 404));
      return;
    }

    return reply.send(ok(fastify.appStore.buildPayment(order), "mock payment created"));
  });
};
