"use strict";

const { z } = require("zod");
const { ok, fail, requireUser } = require("./_shared");

const ticketSchema = z.object({
  orderId: z.string().optional().default(""),
  name: z.string().optional().default(""),
  phone: z.string().optional().default(""),
  content: z.string().min(1),
  subject: z.string().optional().default("配送咨询")
});

module.exports = async function supportRoutes(fastify) {
  fastify.post("/support/tickets", async (request, reply) => {
    const auth = requireUser(fastify, request, reply);
    if (!auth.ok) {
      return auth.response;
    }

    const parsed = ticketSchema.safeParse(request.body || {});
    if (!parsed.success) {
      reply.code(400);
      return reply.send(fail("invalid support payload", 400, {
        issues: parsed.error.issues
      }));
    }

    const payload = parsed.data;
    const ticket = fastify.appStore.createTicket(payload);
    return reply.send(ok(ticket));
  });
};
