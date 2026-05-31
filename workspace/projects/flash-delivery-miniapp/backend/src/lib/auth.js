"use strict";

const jwt = require("jsonwebtoken");

function signUserToken(config, user) {
  return jwt.sign(
    {
      sub: user.id,
      phone: user.phone || "",
      tenantId: config.tenantId
    },
    config.jwtSecret,
    {
      expiresIn: "7d"
    }
  );
}

function readBearerToken(headerValue = "") {
  if (!headerValue || !headerValue.startsWith("Bearer ")) return "";
  return headerValue.slice("Bearer ".length).trim();
}

function verifyUserToken(config, token) {
  if (!token) return null;
  try {
    return jwt.verify(token, config.jwtSecret);
  } catch {
    return null;
  }
}

module.exports = {
  signUserToken,
  readBearerToken,
  verifyUserToken
};
