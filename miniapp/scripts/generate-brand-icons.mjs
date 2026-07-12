#!/usr/bin/env node
/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: generate-brand-icons.mjs
 * 功能描述: 业务模块实现
 * 作者: 廖心慈
 * 创建日期: 2026-05-09
 */

import { mkdirSync } from "node:fs";
import path from "node:path";
import { createRequire } from "node:module";
import { fileURLToPath } from "node:url";
import sharp from "sharp";

const require = createRequire(import.meta.url);
const lucideNodes = require("lucide-static/icon-nodes.json");

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const root = path.resolve(__dirname, "..");
const uiDir = path.join(root, "assets", "ui");
const tabbarDir = path.join(root, "assets", "tabbar");

const ink = "#101114";
const muted = "#8a8f99";
const brand = "#ffce35";
const brandStrong = "#ff9f1c";
const green = "#18b77b";
const blue = "#36a3ff";
const rose = "#ff6f91";

mkdirSync(uiDir, { recursive: true });
mkdirSync(tabbarDir, { recursive: true });

function escapeAttr(value) {
  return String(value).replaceAll("&", "&amp;").replaceAll('"', "&quot;");
}

function renderNode([tag, attrs]) {
  const attrString = Object.entries(attrs)
    .map(([key, value]) => `${key}="${escapeAttr(value)}"`)
    .join(" ");
  return `<${tag} ${attrString} />`;
}

function iconMarkup(name) {
  const nodes = lucideNodes[name];
  if (!nodes) throw new Error(`Missing lucide icon: ${name}`);
  return nodes.map(renderNode).join("");
}

function uiSvg(name, accent = brand) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg width="96" height="96" viewBox="0 0 96 96" xmlns="http://www.w3.org/2000/svg">
  <defs>
    <linearGradient id="card" x1="18" y1="12" x2="82" y2="88" gradientUnits="userSpaceOnUse">
      <stop offset="0" stop-color="#ffffff"/>
      <stop offset="1" stop-color="#f3f5f8"/>
    </linearGradient>
    <filter id="shadow" x="-30%" y="-30%" width="160%" height="160%">
      <feDropShadow dx="0" dy="8" stdDeviation="8" flood-color="#101114" flood-opacity="0.12"/>
    </filter>
  </defs>
  <rect x="10" y="10" width="76" height="76" rx="27" fill="url(#card)" filter="url(#shadow)"/>
  <rect x="10.5" y="10.5" width="75" height="75" rx="26.5" fill="none" stroke="#ffffff" stroke-opacity="0.9"/>
  <circle cx="70" cy="26" r="12" fill="${accent}" fill-opacity="0.22"/>
  <circle cx="72" cy="25" r="4" fill="${accent}" fill-opacity="0.95"/>
  <g transform="translate(25 25) scale(1.92)" fill="none" stroke="${ink}" stroke-width="1.65" stroke-linecap="round" stroke-linejoin="round">
    ${iconMarkup(name)}
  </g>
</svg>`;
}

function tabbarSvg(name, color) {
  return `<?xml version="1.0" encoding="UTF-8"?>
<svg width="81" height="81" viewBox="0 0 81 81" xmlns="http://www.w3.org/2000/svg">
  <g transform="translate(18 18) scale(1.88)" fill="none" stroke="${color}" stroke-width="1.85" stroke-linecap="round" stroke-linejoin="round">
    ${iconMarkup(name)}
  </g>
</svg>`;
}

const uiIcons = {
  api: ["server-cog", blue],
  city: ["map", green],
  clock: ["clock-3", brand],
  cold: ["snowflake", blue],
  coupon: ["ticket-percent", brandStrong],
  device: ["smartphone", blue],
  document: ["file-text", brand],
  dropoff: ["map-pin-house", brandStrong],
  enterprise: ["building-2", brand],
  flower: ["flower-2", rose],
  list: ["list-checks", brand],
  note: ["notebook-pen", brand],
  phone: ["phone", green],
  pickup: ["map-pinned", green],
  price: ["circle-dollar-sign", brand],
  qq: ["user-round", blue],
  rider: ["bike", brand],
  route: ["route", green],
  shield: ["shield-check", brand],
  support: ["headphones", brand],
  user: ["user-round", brand],
  warning: ["triangle-alert", brandStrong],
  wechat: ["message-circle", green],
  weight: ["weight", brand],
};

const tabbarIcons = {
  home: "house",
  order: "square-pen",
  track: "route",
  orders: "package-check",
  profile: "user-round",
};

for (const [assetName, [lucideName, accent]] of Object.entries(uiIcons)) {
  await sharp(Buffer.from(uiSvg(lucideName, accent)))
    .png()
    .toFile(path.join(uiDir, `${assetName}.png`));
}

for (const [assetName, lucideName] of Object.entries(tabbarIcons)) {
  for (const [state, color] of [["normal", muted], ["active", brand]]) {
    await sharp(Buffer.from(tabbarSvg(lucideName, color)))
      .png()
      .toFile(path.join(tabbarDir, `${assetName}-${state}.png`));
  }
}

console.log(`Generated ${Object.keys(uiIcons).length} UI icons and ${Object.keys(tabbarIcons).length * 2} tabbar icons from Lucide.`);
