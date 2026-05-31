"use strict";

const { z } = require("zod");
const { estimatePrice } = require("../lib/estimate");
const { ok, fail } = require("../lib/response");

const schema = z.object({
  fromAddress: z.string().default(""),
  toAddress: z.string().default(""),
  itemType: z.string().default("文件证件"),
  weight: z.string().default("1kg以内")
});

module.exports = async function pricingRoutes(fastify) {
  fastify.post("/pricing/estimate", async (request, reply) => {
    const parsed = schema.safeParse(request.body || {});
    if (!parsed.success) {
      reply.code(400);
      return reply.send(fail("invalid pricing payload", 400, {
        issues: parsed.error.issues
      }));
    }

    return reply.send(ok(estimatePrice(parsed.data)));
  });
};
