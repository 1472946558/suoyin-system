const PRINTER_KEY = "gr_bluetooth_printer";
const DISCOVERY_TIMEOUT = 3200;
const WRITE_CHUNK_SIZE = 120;

function normalizeLine(value) {
  return String(value || "").replace(/\r?\n/g, " ").trim();
}

function moneyText(value) {
  const parsed = Number(value || 0);
  return Number.isFinite(parsed) ? parsed.toFixed(2) : "0.00";
}

function getOrderNo(record) {
  return normalizeLine(record && (record.orderNo || record.id)) || "未生成";
}

function buildReceiptPrintText(record, detailType, quote, photoValidation) {
  const source = record || {};
  const lines = [
    "黄金收银",
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
      lines.push(`   ${Number(item.quantity || 0)} x ${moneyText(item.unitPrice)} = ${moneyText(item.amount)}`);
    });
    lines.push("------------------------");
    lines.push(`收款金额：￥${moneyText(source.totalAmount || source.amount)}`);
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
    "黄金门店",
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

function textToArrayBuffer(text) {
  const commandText = `\x1B@\n${text}\n\x1DVA\x00`;
  const encoded = encodeURIComponent(commandText);
  const bytes = [];
  for (let index = 0; index < encoded.length; index += 1) {
    const char = encoded.charAt(index);
    if (char === "%") {
      bytes.push(parseInt(encoded.substr(index + 1, 2), 16));
      index += 2;
    } else {
      bytes.push(char.charCodeAt(0));
    }
  }
  return new Uint8Array(bytes).buffer;
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
  return normalizeLine(device && (device.name || device.localName)) || "未命名打印机";
}

function discoverDevices() {
  const foundMap = {};
  return wxCall("openBluetoothAdapter")
    .then(function() {
      wx.onBluetoothDeviceFound(function(result) {
        (result.devices || []).forEach(function(device) {
          const name = getDeviceName(device);
          if (!device.deviceId || name === "未命名打印机") return;
          foundMap[device.deviceId] = Object.assign({}, device, { displayName: name });
        });
      });
      return wxCall("startBluetoothDevicesDiscovery", {
        allowDuplicatesKey: false,
        powerLevel: "high"
      });
    })
    .then(function() {
      return sleep(DISCOVERY_TIMEOUT);
    })
    .then(function() {
      return wxCall("stopBluetoothDevicesDiscovery").catch(function() {});
    })
    .then(function() {
      return Object.keys(foundMap).map(function(key) {
        return foundMap[key];
      });
    });
}

function chooseDevice(devices) {
  if (!devices.length) {
    return Promise.reject(new Error("没有发现蓝牙打印机，请确认打印机已开机并打开蓝牙。"));
  }
  const list = devices.slice(0, 6);
  return new Promise(function(resolve, reject) {
    wx.showActionSheet({
      itemList: list.map(function(device) {
        return device.displayName;
      }),
      success(result) {
        resolve(list[result.tapIndex]);
      },
      fail(error) {
        reject(error || new Error("未选择打印机"));
      }
    });
  });
}

function findWritableCharacteristic(deviceId) {
  return wxCall("getBLEDeviceServices", { deviceId }).then(function(serviceResult) {
    const services = (serviceResult.services || []).filter(function(service) {
      return service && service.isPrimary;
    });
    return services.reduce(function(chain, service) {
      return chain.catch(function() {
        return wxCall("getBLEDeviceCharacteristics", {
          deviceId,
          serviceId: service.uuid
        }).then(function(charResult) {
          const writable = (charResult.characteristics || []).find(function(item) {
            return item.properties && (item.properties.write || item.properties.writeNoResponse);
          });
          if (!writable) {
            return Promise.reject(new Error("当前服务没有可写特征值"));
          }
          return {
            deviceId,
            serviceId: service.uuid,
            characteristicId: writable.uuid
          };
        });
      });
    }, Promise.reject(new Error("未找到可写蓝牙服务")));
  });
}

function connectPrinter() {
  const saved = wx.getStorageSync(PRINTER_KEY);
  if (saved && saved.deviceId && saved.serviceId && saved.characteristicId) {
    return wxCall("createBLEConnection", { deviceId: saved.deviceId })
      .then(function() {
        return saved;
      })
      .catch(function() {
        wx.removeStorageSync(PRINTER_KEY);
        return connectPrinter();
      });
  }

  return discoverDevices()
    .then(chooseDevice)
    .then(function(device) {
      return wxCall("createBLEConnection", { deviceId: device.deviceId }).then(function() {
        return findWritableCharacteristic(device.deviceId);
      });
    })
    .then(function(printer) {
      wx.setStorageSync(PRINTER_KEY, printer);
      return printer;
    });
}

function writeChunks(printer, buffer) {
  const bytes = new Uint8Array(buffer);
  let offset = 0;
  function next() {
    if (offset >= bytes.length) {
      return Promise.resolve();
    }
    const chunk = bytes.slice(offset, offset + WRITE_CHUNK_SIZE).buffer;
    offset += WRITE_CHUNK_SIZE;
    return wxCall("writeBLECharacteristicValue", {
      deviceId: printer.deviceId,
      serviceId: printer.serviceId,
      characteristicId: printer.characteristicId,
      value: chunk
    }).then(function() {
      return sleep(80);
    }).then(next);
  }
  return next();
}

function printTextViaBluetooth(text) {
  return connectPrinter().then(function(printer) {
    return writeChunks(printer, textToArrayBuffer(text));
  });
}

function resetSavedPrinter() {
  wx.removeStorageSync(PRINTER_KEY);
}

module.exports = {
  buildReceiptPrintText,
  buildLabelPrintText,
  printTextViaBluetooth,
  resetSavedPrinter
};
