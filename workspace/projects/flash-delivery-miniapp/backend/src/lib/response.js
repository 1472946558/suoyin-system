"use strict";

function ok(data, message = "ok") {
  return {
    code: 0,
    message,
    data
  };
}

function fail(message, code = 1, details = {}) {
  return {
    code,
    message,
    data: details
  };
}

module.exports = {
  ok,
  fail
};
