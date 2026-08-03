"use strict";
const common_vendor = require("../common/vendor.js");
const BASE_URL = "http://localhost:8080";
const API_URL = BASE_URL + "/api/v1";
const Code = {
  Success: 0,
  ParamError: 400,
  Unauthorized: 401,
  Forbidden: 403,
  ProfileIncomplete: 4001,
  ProfileAuditing: 4002,
  ProfileRejected: 4003,
  Banned: 4004
};
let redirecting = false;
function redirectOnce(fn) {
  if (redirecting)
    return;
  redirecting = true;
  fn();
  setTimeout(() => redirecting = false, 1200);
}
function request(options) {
  const token = common_vendor.index.getStorageSync("token");
  return new Promise((resolve, reject) => {
    common_vendor.index.request({
      url: API_URL + options.url,
      method: options.method || "GET",
      data: options.data,
      header: {
        "Content-Type": "application/json",
        ...token ? { Authorization: `Bearer ${token}` } : {}
      },
      timeout: 15e3,
      success(res) {
        const body = res.data;
        if (!body || typeof body.code !== "number") {
          if (!options.silent)
            common_vendor.index.showToast({ title: "服务异常", icon: "none" });
          reject(new Error("响应格式异常"));
          return;
        }
        if (body.code === Code.Success) {
          resolve(body.data);
          return;
        }
        handleBizError(body, options);
        reject(body);
      },
      fail(err) {
        if (!options.silent) {
          common_vendor.index.showToast({ title: "网络连接失败", icon: "none" });
        }
        reject(err);
      }
    });
  });
}
function handleBizError(body, options) {
  if (options.noAuthRedirect) {
    if (!options.silent)
      common_vendor.index.showToast({ title: body.message, icon: "none" });
    return;
  }
  switch (body.code) {
    case Code.Unauthorized:
      common_vendor.index.removeStorageSync("token");
      redirectOnce(() => common_vendor.index.reLaunch({ url: "/pages/login/index" }));
      break;
    case Code.ProfileIncomplete:
      redirectOnce(() => {
        common_vendor.index.showToast({ title: "请先完善资料", icon: "none" });
        setTimeout(() => common_vendor.index.reLaunch({ url: "/pages/profile/setup" }), 800);
      });
      break;
    case Code.ProfileAuditing:
    case Code.ProfileRejected:
    case Code.Banned:
      redirectOnce(
        () => common_vendor.index.navigateTo({
          url: `/pages/common/status?code=${body.code}&msg=${encodeURIComponent(body.message)}`
        })
      );
      break;
    default:
      if (!options.silent) {
        common_vendor.index.showToast({ title: body.message || "操作失败", icon: "none" });
      }
  }
}
const http = {
  get: (url, data, opts) => request({ url, method: "GET", data, ...opts }),
  post: (url, data, opts) => request({ url, method: "POST", data, ...opts }),
  put: (url, data, opts) => request({ url, method: "PUT", data, ...opts }),
  del: (url, data, opts) => request({ url, method: "DELETE", data, ...opts })
};
function uploadImage(filePath, scene) {
  const token = common_vendor.index.getStorageSync("token");
  return new Promise((resolve, reject) => {
    common_vendor.index.uploadFile({
      url: `${API_URL}/common/upload/image?scene=${scene}`,
      filePath,
      name: "file",
      header: token ? { Authorization: `Bearer ${token}` } : {},
      success(res) {
        try {
          const body = JSON.parse(res.data);
          if (body.code === Code.Success) {
            resolve(body.data.url);
          } else {
            common_vendor.index.showToast({ title: body.message, icon: "none" });
            reject(body);
          }
        } catch (e) {
          reject(e);
        }
      },
      fail: reject
    });
  });
}
exports.BASE_URL = BASE_URL;
exports.Code = Code;
exports.http = http;
exports.uploadImage = uploadImage;
