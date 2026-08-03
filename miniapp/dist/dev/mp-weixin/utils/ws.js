"use strict";
var __defProp = Object.defineProperty;
var __defNormalProp = (obj, key, value) => key in obj ? __defProp(obj, key, { enumerable: true, configurable: true, writable: true, value }) : obj[key] = value;
var __publicField = (obj, key, value) => {
  __defNormalProp(obj, typeof key !== "symbol" ? key + "" : key, value);
  return value;
};
const common_vendor = require("../common/vendor.js");
const utils_request = require("./request.js");
const HEARTBEAT_INTERVAL = 25e3;
const MAX_RETRY_DELAY = 3e4;
class WSClient {
  constructor() {
    __publicField(this, "task", null);
    __publicField(this, "handlers", /* @__PURE__ */ new Map());
    __publicField(this, "heartbeatTimer", null);
    __publicField(this, "reconnectTimer", null);
    __publicField(this, "retryCount", 0);
    __publicField(this, "manualClosed", false);
    __publicField(this, "connecting", false);
  }
  get connected() {
    return !!this.task && !this.connecting;
  }
  connect() {
    const token = common_vendor.index.getStorageSync("token");
    if (!token || this.task || this.connecting)
      return;
    this.manualClosed = false;
    this.connecting = true;
    const task = common_vendor.index.connectSocket({
      url: `${utils_request.BASE_URL.replace(/^http/, "ws")}/ws?token=${token}`,
      complete: () => {
      }
    });
    this.task = task;
    task.onOpen(() => {
      this.connecting = false;
      this.retryCount = 0;
      this.startHeartbeat();
      this.emit("__open", null);
    });
    task.onMessage((res) => {
      try {
        const env = JSON.parse(res.data);
        if (env.type === "pong")
          return;
        this.emit(env.type, env.data);
      } catch {
      }
    });
    task.onClose(() => {
      this.cleanup();
      this.scheduleReconnect();
    });
    task.onError(() => {
      this.connecting = false;
      this.cleanup();
      this.scheduleReconnect();
    });
  }
  close() {
    var _a;
    this.manualClosed = true;
    clearTimeout(this.reconnectTimer);
    try {
      (_a = this.task) == null ? void 0 : _a.close({});
    } catch {
    }
    this.cleanup();
  }
  send(type, data) {
    if (!this.task) {
      this.connect();
      return false;
    }
    this.task.send({ data: JSON.stringify({ type, data }) });
    return true;
  }
  on(type, handler) {
    if (!this.handlers.has(type))
      this.handlers.set(type, /* @__PURE__ */ new Set());
    this.handlers.get(type).add(handler);
  }
  off(type, handler) {
    var _a;
    (_a = this.handlers.get(type)) == null ? void 0 : _a.delete(handler);
  }
  emit(type, data) {
    var _a;
    (_a = this.handlers.get(type)) == null ? void 0 : _a.forEach((h) => h(data));
  }
  startHeartbeat() {
    clearInterval(this.heartbeatTimer);
    this.heartbeatTimer = setInterval(() => {
      var _a;
      (_a = this.task) == null ? void 0 : _a.send({ data: JSON.stringify({ type: "ping" }) });
    }, HEARTBEAT_INTERVAL);
  }
  cleanup() {
    clearInterval(this.heartbeatTimer);
    this.heartbeatTimer = null;
    this.task = null;
    this.connecting = false;
  }
  scheduleReconnect() {
    if (this.manualClosed || !common_vendor.index.getStorageSync("token"))
      return;
    const delay = Math.min(1e3 * 2 ** this.retryCount, MAX_RETRY_DELAY);
    this.retryCount++;
    clearTimeout(this.reconnectTimer);
    this.reconnectTimer = setTimeout(() => this.connect(), delay);
  }
}
const ws = new WSClient();
exports.ws = ws;
