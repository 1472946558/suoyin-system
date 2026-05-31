"use strict";

const itemTypes = ["文件证件", "数码配件", "鲜花蛋糕", "生鲜食品", "其它物品"];
const weightTypes = ["1kg以内", "1-3kg", "3-5kg", "5kg以上"];

function estimatePrice(form = {}) {
  const weightIndex = weightTypes.indexOf(form.weight);
  const itemIndex = itemTypes.indexOf(form.itemType);
  const base = 18;
  const weightFee = Math.max(weightIndex, 0) * 8;
  const itemFee = itemIndex === 2 || itemIndex === 3 ? 6 : 0;
  const fromLen = String(form.fromAddress || "").trim().length;
  const toLen = String(form.toAddress || "").trim().length;
  const distanceValue = 4.8 + ((fromLen + toLen) % 8);

  return {
    price: base + weightFee + itemFee + Math.round(distanceValue * 1.6),
    distance: `${distanceValue.toFixed(1)}km`
  };
}

module.exports = {
  estimatePrice
};
