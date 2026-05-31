"use strict";

const { z } = require("zod");
const { fail, ok, requireUser } = require("./_shared");

const createOrderSchema = z.object({
  fromAddress: z.string().min(1),
  toAddress: z.string().min(1),
  itemType: z.string().min(1),
  weight: z.string().min(1),
  note: z.string().optional().default(""),
  clientEstimate: z.object({
    price: z.number().optional(),
    distance: z.string().optional()
  }).optional()
});

module.exports = async function orderRoutes(fastify) {
  fastify.post("/orders", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const parsed = createOrderSchema.safeParse(request.body || {});
    if (!parsed.success) {
      reply.code(400);
      return reply.send(fail("invalid order payload", 400, {
        issues: parsed.error.issues
      }));
    }

    const payload = parsed.data;
    const order = fastify.appStore.createOrder(payload);
    return reply.send(ok(order));
  });

  fastify.get("/orders", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    return reply.send(ok({
      items: fastify.appStore.listOrders()
    }));
  });

  fastify.get("/orders/:id", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const order = fastify.appStore.getOrder(request.params.id);
    if (!order) {
      reply.code(404).send(fail("order not found", 404));
      return;
    }
    return reply.send(ok(order));
  });

  fastify.get("/orders/:id/track", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const order = fastify.appStore.getOrder(request.params.id);
    if (!order) {
      reply.code(404).send(fail("order not found", 404));
      return;
    }
    return reply.send(ok(fastify.appStore.buildTrack(order)));
  });
};
