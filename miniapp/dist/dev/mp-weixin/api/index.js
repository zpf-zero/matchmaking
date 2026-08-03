"use strict";
Object.defineProperty(exports, Symbol.toStringTag, { value: "Module" });
const utils_request = require("../utils/request.js");
const authApi = {
  /** mock 模式下 code 直接作为 openid，换不同 code 可模拟多个账号 */
  wxLogin: (code) => utils_request.http.post(
    "/auth/wx-login",
    { code },
    { noAuthRedirect: true }
  ),
  bindPhone: (data) => utils_request.http.post("/auth/bind-phone", data),
  logout: () => utils_request.http.post("/auth/logout"),
  deleteAccount: () => utils_request.http.del("/auth/account")
};
const userApi = {
  me: () => utils_request.http.get("/user/me"),
  saveProfile: (data) => utils_request.http.put("/user/profile", data),
  saveRequirement: (data) => utils_request.http.put("/user/requirement", data),
  submitAudit: () => utils_request.http.post("/user/submit-audit"),
  detail: (id) => utils_request.http.get(`/user/detail/${id}`),
  addPhoto: (url) => utils_request.http.post("/user/photo", { url }),
  deletePhoto: (id) => utils_request.http.del(`/user/photo/${id}`),
  sortPhotos: (ids) => utils_request.http.put("/user/photo/sort", { ids })
};
const discoverApi = {
  recommend: (page = 1, pageSize = 20) => utils_request.http.get("/discover/recommend", { page, page_size: pageSize }),
  search: (params) => utils_request.http.get("/discover/search", params)
};
const interactionApi = {
  like: (toUserId) => utils_request.http.post("/interaction/like", {
    to_user_id: toUserId
  }),
  pass: (toUserId) => utils_request.http.post("/interaction/pass", { to_user_id: toUserId }),
  unlike: (toUserId) => utils_request.http.del(`/interaction/like/${toUserId}`),
  likedMe: (page = 1) => utils_request.http.get("/interaction/liked-me", { page }),
  myLikes: (page = 1) => utils_request.http.get("/interaction/my-likes", { page }),
  viewedMe: (page = 1) => utils_request.http.get("/interaction/viewed-me", { page }),
  unreadCount: () => utils_request.http.get(
    "/interaction/unread-count",
    void 0,
    { silent: true }
  ),
  markRead: (type) => utils_request.http.put(`/interaction/read?type=${type}`)
};
const chatApi = {
  conversations: () => utils_request.http.get("/conversation/list"),
  closeConversation: (id) => utils_request.http.del(`/conversation/${id}`),
  messages: (conversationId, beforeId = 0, limit = 20) => utils_request.http.get("/message/list", {
    conversation_id: conversationId,
    before_id: beforeId,
    limit
  }),
  send: (conversationId, content, msgType = 1) => utils_request.http.post("/message/send", {
    conversation_id: conversationId,
    content,
    msg_type: msgType
  }),
  read: (conversationId) => utils_request.http.put("/message/read", { conversation_id: conversationId })
};
const commonApi = {
  dict: () => utils_request.http.get("/common/dict"),
  config: () => utils_request.http.get(
    "/common/config",
    void 0,
    { noAuthRedirect: true }
  ),
  report: (data) => utils_request.http.post("/common/report", data),
  addBlacklist: (toUserId) => utils_request.http.post("/common/blacklist", { to_user_id: toUserId }),
  removeBlacklist: (id) => utils_request.http.del(`/common/blacklist/${id}`),
  blacklist: () => utils_request.http.get(
    "/common/blacklist"
  )
};
const adminApi = {
  overview: () => utils_request.http.get("/admin/overview"),
  auditList: (page = 1) => utils_request.http.get("/admin/audit/list", { page }),
  audit: (userId, pass, remark = "") => utils_request.http.post(`/admin/audit/${userId}`, { pass, remark }),
  userList: (params) => utils_request.http.get("/admin/user/list", params),
  userDetail: (id) => utils_request.http.get(`/admin/user/${id}`),
  ban: (id, reason) => utils_request.http.post(`/admin/user/${id}/ban`, { reason }),
  unban: (id) => utils_request.http.post(`/admin/user/${id}/unban`),
  reportList: (status = -1, page = 1) => utils_request.http.get("/admin/report/list", { status, page }),
  handleReport: (id, action, remark = "") => utils_request.http.post(`/admin/report/${id}/handle`, { action, remark })
};
exports.adminApi = adminApi;
exports.authApi = authApi;
exports.chatApi = chatApi;
exports.commonApi = commonApi;
exports.discoverApi = discoverApi;
exports.interactionApi = interactionApi;
exports.userApi = userApi;
