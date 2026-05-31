"use strict";

const { z } = require("zod");
const { ok, fail } = require("../lib/response");
const { signUserToken } = require("../lib/auth");

const profileSchema = z.object({
  id: z.string().optional(),
  name: z.string().default("演示用户"),
  phone: z.string().default(""),
  wechat: z.string().default(""),
  qq: z.string().default(""),
  company: z.string().default(""),
  city: z.string().default("深圳"),
  defaultFromAddress: z.string().default(""),
  defaultToAddress: z.string().default("")
});

const loginSchema = z.object({
  code: z.string().min(1).default("mock-code"),
  profile: profileSchema
});

module.exports = async function authRoutes(fastify) {
  fastify.post("/auth/wechat-login", async (request, reply) => {
    const parsed = loginSchema.safeParse(request.body || {});
    if (!parsed.success) {
      reply.code(400);
      return reply.send(fail("invalid auth payload", 400, {
        issues: parsed.error.issues
      }));
    }

    const payload = parsed.data;
    const user = fastify.appStore.saveUser(payload.profile);
    const token = signUserToken(fastify.appConfig, user);

    return reply.send(ok({
      userId: user.id,
      token
    }));
  });
};
