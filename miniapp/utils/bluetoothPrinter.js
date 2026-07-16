/*
 * Copyright (c) 2026 北京纵横时空科技有限责任公司
 *
 * 本软件（包含源代码、可执行文件及所有相关文档）受中华人民共和国著作权法
 * 及其他知识产权相关法律保护。未经北京纵横时空科技有限责任公司事先书面授权，
 * 任何单位或个人不得以任何形式复制、修改、分发、出租、反编译本软件或其任何部分。
 *
 * 文件名: bluetoothPrinter.js
 * 功能描述: 工具函数
 * 作者: 廖心慈
 * 创建日期: 2026-06-10
 */

const esc = require("./printer-sdk/esc.js");
const tsc = require("./printer-sdk/tsc.js");
const { getSystemProfile } = require("./systemStore");

const app = getApp();

const PRINTER_KEY = "gr_bluetooth_printer_v2";
const DISCOVERY_TIMEOUT = 4000;
const WRITE_CHUNK_SIZE = 60;
const WRITE_INTERVAL = 40;

function normalizeLine(value) {
  return String(value || "").replace(/\r?\n/g, " ").trim();
}

function sanitizePrintText(value) {
  return normalizeLine(value).replace(/"/g, "'");
}

function moneyText(value) {
  const parsed = Number(value || 0);
  return Number.isFinite(parsed) ? parsed.toFixed(2) : "0.00";
}

function truncateText(value, maxLength) {
  const text = sanitizePrintText(value);
  if (text.length <= maxLength) {
    return text;
  }
  return `${text.slice(0, Math.max(0, maxLength - 3))}...`;
}

function getOrderNo(record) {
  return normalizeLine(record && (record.orderNo || record.id)) || "未生成";
}

function getPrintTitle(record) {
  const profile = getSystemProfile ? getSystemProfile() : {};
  return normalizeLine(
    profile.receiptTitle ||
    profile.appName ||
    (record && record.storeName) ||
    "金匠倌收银"
  ) || "金匠倌收银";
}

function getShortBrandTitle(record) {
  const profile = getSystemProfile ? getSystemProfile() : {};
  return truncateText(
    profile.appName ||
    (record && record.storeName) ||
    "黄金门店",
    10
  );
}

function safeCall(fn) {
  try {
    return fn();
  } catch (error) {
    return null;
  }
}

function getAppBleInformation() {
  if (!app || !app.BLEInformation) {
    return null;
  }
  return app.BLEInformation;
}

function applyBleInformation(printer) {
  const ble = getAppBleInformation();
  if (!ble || !printer) {
    return;
  }
  ble.platform = app.getPlatform ? app.getPlatform() : ble.platform;
  ble.deviceId = printer.deviceId || "";
  ble.writeCharaterId = printer.characteristicId || "";
  ble.writeServiceId = printer.serviceId || "";
  ble.notifyCharaterId = printer.notifyCharacteristicId || "";
  ble.notifyServiceId = printer.notifyServiceId || "";
  ble.readCharaterId = printer.readCharacteristicId || "";
  ble.readServiceId = printer.readServiceId || "";
}

function clearBleInformation() {
  const ble = getAppBleInformation();
  if (!ble) {
    return;
  }
  ble.deviceId = "";
  ble.writeCharaterId = "";
  ble.writeServiceId = "";
  ble.notifyCharaterId = "";
  ble.notifyServiceId = "";
  ble.readCharaterId = "";
  ble.readServiceId = "";
}

function wxCall(method, options) {
  return new Promise(function(resolve, reject) {
    wx[method](Object.assign({}, options || {}, {
      success(result) {
        resolve(result || {});
      },
      fail(error) {
        reject(error || new Error(`${method} failed`));
      }
    }));
  });
}

function sleep(ms) {
  return new Promise(function(resolve) {
    setTimeout(resolve, ms);
  });
}

function getDeviceName(device) {
  return normalizeLine(device && (device.name || device.localName)) || "";
}

function shouldKeepDevice(device) {
  return !!(device && device.deviceId && getDeviceName(device));
}

function assignDisplayName(device) {
  return Object.assign({}, device, {
    displayName: getDeviceName(device)
  });
}

function savePrinter(printer) {
  wx.setStorageSync(PRINTER_KEY, printer);
  applyBleInformation(printer);
  return printer;
}

function ensureBluetoothReady() {
  return wxCall("openBluetoothAdapter").catch(function(error) {
    const message = (error && error.errCode === 10001)
      ? "蓝牙未开启，请先在手机系统设置中打开蓝牙。"
      : "蓝牙初始化失败，请确认微信已获得蓝牙权限。";
    return Promise.reject(new Error(message));
  });
}

function collectDiscoveredDevices(foundMap) {
  return wxCall("getBluetoothDevices")
    .catch(function() {
      return { devices: [] };
    })
    .then(function(result) {
      (result.devices || []).forEach(function(device) {
        if (!shouldKeepDevice(device)) {
          return;
        }
        foundMap[device.deviceId] = assignDisplayName(device);
      });
      return Object.keys(foundMap).map(function(key) {
        return foundMap[key];
      });
    });
}

function discoverDevices() {
  const foundMap = {};
  return ensureBluetoothReady()
    .then(function() {
      if (wx.offBluetoothDeviceFound) {
        safeCall(function() {
          wx.offBluetoothDeviceFound();
        });
      }
      wx.onBluetoothDeviceFound(function(result) {
        (result.devices || []).forEach(function(device) {
          if (!shouldKeepDevice(device)) {
            return;
          }
          foundMap[device.deviceId] = assignDisplayName(device);
        });
      });
      return wxCall("startBluetoothDevicesDiscovery", {
        allowDuplicatesKey: false
      });
    })
    .then(function() {
      return sleep(DISCOVERY_TIMEOUT);
    })
    .then(function() {
      return collectDiscoveredDevices(foundMap);
    })
    .finally(function() {
      return wxCall("stopBluetoothDevicesDiscovery").catch(function() {});
    });
}

function chooseDevice(devices) {
  if (!devices.length) {
    return Promise.reject(new Error("没有发现蓝牙打印机，请确认打印机已开机并进入配对状态。"));
  }
  const candidates = devices.slice(0, 6);
  return new Promise(function(resolve, reject) {
    wx.showActionSheet({
      itemList: candidates.map(function(device) {
        return device.displayName;
      }),
      success(result) {
        resolve(candidates[result.tapIndex]);
      },
      fail(error) {
        reject(error || new Error("未选择打印机"));
      }
    });
  });
}

function inspectCharacteristics(deviceId, services, index, printer) {
  if (index >= services.length) {
    if (!printer.characteristicId || !printer.serviceId) {
      return Promise.reject(new Error("没有找到可写入的蓝牙特征值，请换一台设备或联系卖家确认协议。"));
    }
    return Promise.resolve(printer);
  }

  const service = services[index];
  if (!service || !service.uuid || service.isPrimary === false) {
    return inspectCharacteristics(deviceId, services, index + 1, printer);
  }

  return wxCall("getBLEDeviceCharacteristics", {
    deviceId,
    serviceId: service.uuid
  })
    .then(function(result) {
      (result.characteristics || []).forEach(function(item) {
        if (!item || !item.uuid || !item.properties) {
          return;
        }
        if (!printer.characteristicId && (item.properties.write || item.properties.writeNoResponse)) {
          printer.characteristicId = item.uuid;
          printer.serviceId = service.uuid;
        }
        if (!printer.notifyCharacteristicId && item.properties.notify) {
          printer.notifyCharacteristicId = item.uuid;
          printer.notifyServiceId = service.uuid;
        }
        if (!printer.readCharacteristicId && item.properties.read) {
          printer.readCharacteristicId = item.uuid;
          printer.readServiceId = service.uuid;
        }
      });
      if (printer.characteristicId && printer.serviceId) {
        return printer;
      }
      return inspectCharacteristics(deviceId, services, index + 1, printer);
    })
    .catch(function() {
      return inspectCharacteristics(deviceId, services, index + 1, printer);
    });
}

function resolvePrinterProfile(device) {
  const printer = {
    deviceId: device.deviceId,
    displayName: device.displayName || getDeviceName(device),
    serviceId: "",
    characteristicId: "",
    notifyCharacteristicId: "",
    notifyServiceId: "",
    readCharacteristicId: "",
    readServiceId: ""
  };

  return wxCall("getBLEDeviceServices", {
    deviceId: device.deviceId
  }).then(function(result) {
    const services = result.services || [];
    return inspectCharacteristics(device.deviceId, services, 0, printer);
  });
}

function connectToDevice(device) {
  return wxCall("createBLEConnection", {
    deviceId: device.deviceId
  })
    .then(function() {
      return sleep(300);
    })
    .then(function() {
      return resolvePrinterProfile(device);
    })
    .then(savePrinter);
}

function connectSavedPrinter(savedPrinter) {
  return ensureBluetoothReady()
    .then(function() {
      return wxCall("createBLEConnection", {
        deviceId: savedPrinter.deviceId
      });
    })
    .then(function() {
      applyBleInformation(savedPrinter);
      return savedPrinter;
    });
}

function connectPrinter() {
  const savedPrinter = wx.getStorageSync(PRINTER_KEY);
  if (savedPrinter && savedPrinter.deviceId && savedPrinter.serviceId && savedPrinter.characteristicId) {
    return connectSavedPrinter(savedPrinter).catch(function() {
      resetSavedPrinter();
      return connectPrinter();
    });
  }

  return discoverDevices()
    .then(chooseDevice)
    .then(connectToDevice);
}

function bytesToArrayBuffer(bytes) {
  const buffer = new ArrayBuffer(bytes.length);
  const view = new Uint8Array(buffer);
  for (let index = 0; index < bytes.length; index += 1) {
    view[index] = bytes[index];
  }
  return buffer;
}

function writeChunks(printer, bytes) {
  let offset = 0;

  function next() {
    if (offset >= bytes.length) {
      return Promise.resolve(printer);
    }

    const chunk = bytes.slice(offset, offset + WRITE_CHUNK_SIZE);
    offset += WRITE_CHUNK_SIZE;

    return wxCall("writeBLECharacteristicValue", {
      deviceId: printer.deviceId,
      serviceId: printer.serviceId,
      characteristicId: printer.characteristicId,
      value: bytesToArrayBuffer(chunk)
    })
      .then(function() {
        return sleep(WRITE_INTERVAL);
      })
      .then(next);
  }

  return next();
}

function buildReceiptPrintText(record, detailType, quote, photoValidation) {
  const source = record || {};
  const lines = [
    getPrintTitle(source),
    "------------------------",
    `单号：${getOrderNo(source)}`,
    `类型：${detailType === "cashier" ? "门店收银" : "黄金回收"}`,
    `门店：${normalizeLine(source.storeName) || "未填写"}`,
    `操作员：${normalizeLine(source.operatorName || source.createdBy) || "未填写"}`,
    `时间：${normalizeLine(source.createdAt || source.updatedAt) || "刚刚"}`,
    "------------------------"
  ];

  if (detailType === "cashier") {
    const items = Array.isArray(source.items) ? source.items : [];
    lines.push(`客户：${normalizeLine(source.customerName) || "未填写"}`);
    lines.push(`电话：${normalizeLine(source.customerPhone) || "-"}`);
    lines.push("商品明细：");
    items.forEach(function(item, index) {
      lines.push(`${index + 1}. ${normalizeLine(item.name) || "未命名商品"}`);
      if (item.sku) {
        lines.push(`   编码：${normalizeLine(item.sku)}`);
      }
      lines.push(`   ${Number(item.quantity || 0)} x ${moneyText(item.unitPrice)} = ${moneyText(item.amount)}`);
    });
    lines.push("------------------------");
    lines.push(`收款金额：￥${moneyText(source.totalAmount || source.amount)}`);
    lines.push(`订单状态：${source.statusText || (source.status === "refunded" ? "已退单" : "已完成")}`);
    if (source.status === "refunded") {
      lines.push(`退单原因：${normalizeLine(source.refundReason || source.voidReason) || "-"}`);
      lines.push(`退单人：${normalizeLine(source.refundedBy || source.voidedBy) || "-"}`);
      lines.push(`退单时间：${normalizeLine(source.refundedAt || source.voidedAt) || "-"}`);
    }
  } else {
    const nextQuote = quote || {};
    const nextPhotos = photoValidation || {};
    lines.push(`客户：${normalizeLine(source.customerName) || "-"}`);
    lines.push(`电话：${normalizeLine(source.customerPhone) || "-"}`);
    lines.push(`品类：${normalizeLine(source.itemCategory)} / ${normalizeLine(source.itemName)}`);
    lines.push(`成色：${normalizeLine(source.purity)}`);
    lines.push(`毛重：${nextQuote.grossWeightText || "0.00"}g`);
    lines.push(`扣减：${nextQuote.deductionWeightText || "0.00"}g`);
    lines.push(`净重：${nextQuote.netWeightText || "0.00"}g`);
    lines.push(`单价：￥${nextQuote.recyclePriceText || "0.00"}/g`);
    lines.push("------------------------");
    lines.push(`结算金额：￥${nextQuote.amountText || "0.00"}`);
    lines.push(`照片留档：${Number(nextPhotos.count || 0)} 张`);
  }

  lines.push("------------------------");
  lines.push("请核对金额与实物信息");
  lines.push("");
  lines.push("");
  return lines.join("\n");
}

function buildLabelPrintText(record, detailType, quote) {
  const source = record || {};
  const nextQuote = quote || {};
  const lines = [
    getShortBrandTitle(source),
    `单号：${getOrderNo(source)}`,
    `门店：${normalizeLine(source.storeName) || "未填写"}`
  ];

  if (detailType === "cashier") {
    lines.push(`商品：${normalizeLine(source.itemSummary || source.itemName) || "门店商品"}`);
    lines.push(`金额：￥${moneyText(source.totalAmount || source.amount)}`);
  } else {
    lines.push(`品名：${normalizeLine(source.itemName) || "黄金回收"}`);
    lines.push(`成色：${normalizeLine(source.purity) || "-"}`);
    lines.push(`净重：${nextQuote.netWeightText || "0.00"}g`);
    lines.push(`金额：￥${nextQuote.amountText || "0.00"}`);
  }

  lines.push("");
  return lines.join("\n");
}

function buildReceiptCommand(text) {
  const lines = String(text || "").split(/\r?\n/);
  const command = esc.jpPrinter.createNew();

  command.init();
  command.setSelectJustification(1);
  command.setBoldMode(1);
  command.setCharacterSize(17);
  command.setText(lines[0] || "金匠倌收银");
  command.setPrint();

  command.setBoldMode(0);
  command.setCharacterSize(0);
  command.setSelectJustification(0);

  lines.slice(1).forEach(function(line) {
    command.setText(line);
    command.setPrint();
  });

  command.setPrintAndFeedRow(4);
  return command.getData();
}

function buildLabelCommand(record, detailType, quote) {
  const source = record || {};
  const nextQuote = quote || {};
  const orderNo = getOrderNo(source);
  const brandTitle = getShortBrandTitle(source);
  const storeLine = truncateText(`门店 ${source.storeName || "未填写"}`, 14);
  const summary = detailType === "cashier"
    ? truncateText(source.itemSummary || source.itemName || "门店商品", 12)
    : truncateText(source.itemName || "黄金回收", 12);
  const amount = detailType === "cashier"
    ? moneyText(source.totalAmount || source.amount)
    : (nextQuote.amountText || "0.00");
  const extraLine = detailType === "cashier"
    ? `数量 ${Number(source.itemCount || (source.items || []).length || 1)}`
    : `净重 ${nextQuote.netWeightText || "0.00"}g`;
  const specLine = detailType === "cashier"
    ? truncateText(normalizeLine(source.customerName) || "散客", 10)
    : `${truncateText(source.purity || "-", 8)} / ${truncateText(source.itemCategory || "-", 8)}`;

  const command = tsc.jpPrinter.createNew();
  command.setSize(50, 30);
  command.setGap(2);
  command.setCls();
  command.setSpeed(3);
  command.setDensity(8);
  command.setDirection(0);
  command.setReference(0, 0);

  command.setText(16, 16, "TSS24.BF2", 0, 1, 1, brandTitle);
  command.setText(16, 50, "TSS24.BF2", 0, 1, 1, truncateText(`单号 ${orderNo}`, 16));
  command.setBar(16, 82, 248, 2);
  command.setText(16, 96, "TSS24.BF2", 0, 1, 1, storeLine);
  command.setText(16, 128, "TSS24.BF2", 0, 1, 1, truncateText(`品名 ${summary}`, 14));
  command.setText(16, 160, "TSS24.BF2", 0, 1, 1, truncateText(specLine, 16));
  command.setText(16, 192, "TSS24.BF2", 0, 1, 1, truncateText(extraLine, 14));
  command.setText(16, 224, "TSS24.BF2", 0, 1, 1, truncateText(`金额 ￥${amount}`, 14));
  command.setQrcode(296, 56, "M", 4, "A", orderNo);
  command.setPagePrint();

  return command.getData();
}

function printReceiptViaBluetooth(text) {
  const bytes = buildReceiptCommand(text);
  return connectPrinter().then(function(printer) {
    return writeChunks(printer, bytes);
  });
}

function printLabelViaBluetooth(record, detailType, quote) {
  const bytes = buildLabelCommand(record, detailType, quote);
  return connectPrinter().then(function(printer) {
    return writeChunks(printer, bytes);
  });
}

function resetSavedPrinter() {
  const printer = wx.getStorageSync(PRINTER_KEY);
  wx.removeStorageSync(PRINTER_KEY);
  clearBleInformation();
  if (printer && printer.deviceId) {
    safeCall(function() {
      wx.closeBLEConnection({
        deviceId: printer.deviceId
      });
    });
  }
}

module.exports = {
  buildReceiptPrintText,
  buildLabelPrintText,
  printReceiptViaBluetooth,
  printLabelViaBluetooth,
  resetSavedPrinter
};
