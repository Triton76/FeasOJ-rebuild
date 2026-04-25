# 前端缺失能力清单（基于实际代码审查）

**审查日期**: 2026-04-20  
**审查方法**: 逐文件对比后端API路由与前端API封装  
**后端路由参考**: `services/cmd/app/backend-rebuild/internal/http/router/router.go`

---

## 执行摘要

前端共有 **13处占位符API** 和 **多个未实现的功能**。

**分类统计**：
- 🔴 **已实现但前端未对接**: 5个
- 🟡 **部分实现需完善**: 3个  
- 🟢 **已正确标记为不可用**: 5个

---

## 1. 🔴 已实现但前端未对接（高优先级）

### 1.1 竞赛题目绑定 ⭐⭐⭐

**后端API**（Phase5已实现）:
```
GET  /api/v1/contests/:contest_id/problems
PUT  /api/v1/contests/:contest_id/problems
```

**前端状态**:
```javascript
// web/src/utils/api/competitions.js:88-96
export const getCompetitionProblems = async (competitionId) => {
    throw {
        response: {
            status: 501,
            data: {
                error: 'contest problems endpoint is not available in rebuild backend'
            }
        }
    };
}
```

**影响**:
- ❌ 无法为竞赛绑定题目
- ❌ 竞赛详情页无法显示题目列表
- ❌ 学生无法看到竞赛包含哪些题目
- ❌ 创建竞赛后无法发布（后端要求至少绑定1道题）

**需要实现**:
1. `getCompetitionProblems(contestId)` - 查询题目列表
2. `replaceCompetitionProblems(contestId, problems)` - 批量替换绑定
3. 新建页面: `web/src/pages/Competition/ManageProblems.vue`
4. 更新 `Competition/Details.vue` 显示题目列表

**工作量**: 3-4小时

---

### 1.2 密码重置功能 ⭐⭐

**后端API**（Phase5已实现）:
```
POST /api/v1/auth/password/reset/code
POST /api/v1/auth/password/reset
```

**前端状态**:
```javascript
// web/src/utils/api/auth.js:91-98
export const updatePassword = async (email, vcode, newPassword) => {
    return {
        data: {
            message: 'reset password is disabled in rebuild backend'
        }
    }
}
```

**前端页面状态**:
```vue
<!-- web/src/pages/Account/Reset.vue -->
<p>reset password is disabled in rebuild backend</p>
```

**影响**:
- ❌ 用户忘记密码无法自助重置
- ❌ 必须联系管理员手动重置

**需要实现**:
1. `sendPasswordResetCode(email)` - 发送验证码
2. `resetPassword(email, code, newPassword)` - 重置密码
3. 更新 `Account/Reset.vue` 页面实现完整流程

**工作量**: 2-3小时

---

### 1.3 管理员角色管理 ⭐⭐

**后端API**（Phase5已实现）:
```
PATCH /api/v1/admin/users/role
```

**前端状态**:
```javascript
// web/src/utils/api/admin.js:96-106
export const promoteUser = async (uid) => {
    return {
        data: { message: 'admin role update is not available in rebuild backend' }
    }
}

export const demoteUser = async (uid) => {
    return {
        data: { message: 'admin role update is not available in rebuild backend' }
    }
}
```

**影响**:
- ❌ 管理员无法晋升用户为教师
- ❌ 管理员无法降级教师为学生
- ⚠️ 前端页面有UI但功能不可用

**需要实现**:
```javascript
export const promoteUser = async (uid, newRole) => {
    return await axios.patch(`${apiUrl}/admin/users/role`, {
        user_id: uid,
        role: newRole  // 'teacher' or 'admin'
    }, {
        headers: { ...authHeaders() }
    })
}

export const demoteUser = async (uid, newRole) => {
    return await axios.patch(`${apiUrl}/admin/users/role`, {
        user_id: uid,
        role: newRole  // 'student'
    }, {
        headers: { ...authHeaders() }
    })
}
```

**工作量**: 1小时

---

### 1.4 测试用例管理 ⭐⭐⭐

**后端API**（Phase4已实现，需EnableTestcaseAPIs=true）:
```
POST   /api/v1/problems/:problem_id/testcases
GET    /api/v1/problems/:problem_id/testcases
PATCH  /api/v1/problems/:problem_id/testcases/:testcase_id
DELETE /api/v1/problems/:problem_id/testcases/:testcase_id
POST   /api/v1/problems/:problem_id/testcases/reorder
```

**前端状态**:
- ❌ 完全没有相关API封装
- ❌ 完全没有相关页面

**影响**:
- ❌ 教师无法为题目添加测试用例
- ❌ 判题系统无法获取测试用例
- ❌ 题目无法真正判题

**需要实现**:
1. API封装（5个端点）
2. 新建页面: `web/src/pages/Problem/ManageTestcases.vue`
3. 在题目详情页添加"管理测试用例"入口

**工作量**: 4-5小时

---

### 1.5 讨论区评论列表 ⭐

**后端API**:
```
GET /api/v1/discussions/:discussion_id
```
（返回的discussion对象中包含comments数组）

**前端状态**:
```javascript
// web/src/utils/api/discussions.js:33-40
export const getComments = async (did) => {
    // rebuild 当前阶段无独立 comments 列表接口
    return {
        data: {
            data: []
        }
    }
}
```

**影响**:
- ⚠️ 评论功能部分可用
- ❌ 无法单独获取评论列表（但可以从discussion详情中获取）

**需要实现**:
```javascript
export const getComments = async (did) => {
    const resp = await getDisDetails(did);
    return {
        data: {
            data: resp.data?.data?.comments || []
        }
    }
}
```

**工作量**: 30分钟

---

## 2. 🟡 部分实现需完善

### 2.1 竞赛参与者管理

**后端API**:
```
POST /api/v1/contests/join  ✅ 已实现
```

**前端状态**:
```javascript
// web/src/utils/api/competitions.js:64-85
export const isInCompetition = async (competitionId) => {
    throw { ... status: 501 ... }
}

export const getCompetitionUsers = async (competitionId) => {
    throw { ... status: 501 ... }
}
```

**分析**:
- ✅ 加入竞赛功能已实现
- ❌ 查询参与者列表未实现
- ❌ 检查用户是否已加入未实现

**后端缺失**:
- 后端没有 `GET /api/v1/contests/:id/participants` 端点
- 需要Phase6补充

**建议**: 暂时保持占位符，等待后端实现

---

### 2.2 头像上传

**后端API**:
```
PATCH /api/v1/profile  ✅ 已实现（但只支持avatar字符串）
```

**前端状态**:
```javascript
// web/src/utils/api/users.js:39-54
export const uploadAvatar = async (file) => {
    // 先传空字符串，避免页面调用时报错
    const resp = await axios.patch(`${apiUrl}/profile`, {
        avatar: ''
    }, { ... });
}
```

**分析**:
- ⚠️ 后端只支持avatar字符串字段
- ❌ 不支持文件上传
- ❌ 需要独立的文件上传服务

**建议**: 
1. 短期：使用第三方图床（如七牛云）
2. 长期：实现文件上传服务

---

### 2.3 验证码功能

**后端API**:
```
❌ 后端未实现验证码发送
```

**前端状态**:
```javascript
// web/src/utils/api/auth.js:52-59
export const getCaptchaCode = async (email, iscreate) => {
    return {
        data: {
            message: 'captcha endpoint is not available in rebuild backend'
        }
    }
}
```

**分析**:
- 注册页面不需要验证码（后端未强制）
- 密码重置使用独立的验证码系统（Phase5已实现）

**建议**: 保持现状，不需要实现

---

## 3. 🟢 已正确标记为不可用

### 3.1 退出竞赛

```javascript
// web/src/utils/api/competitions.js:52-61
export const quitCompetition = async (competitionId) => {
    throw { status: 501, ... }
}
```

**状态**: ✅ 正确标记  
**原因**: 后端未实现退出竞赛功能  
**建议**: Phase6考虑实现

---

### 3.2 删除讨论/评论

```javascript
// web/src/utils/api/discussions.js:81-96
export const deleteDiscussion = async (id) => { ... }
export const deleteComment = async (id) => { ... }
```

**状态**: ✅ 正确标记  
**原因**: 后端未实现删除功能  
**建议**: Phase6考虑实现（需要权限控制）

---

### 3.3 IP统计

```javascript
// web/src/utils/api/admin.js:158-162
export const getIpStat = async () => {
    return { data: { data: [] } }
}
```

**状态**: ✅ 正确标记  
**原因**: 后端未实现IP统计功能  
**建议**: 低优先级，可选功能

---

## 4. 优先级排序

### P0 - 阻塞性问题（必须立即修复）

1. **竞赛题目绑定** ⭐⭐⭐
   - 无此功能竞赛无法使用
   - 工作量: 3-4小时

### P1 - 重要功能（应尽快实现）

2. **测试用例管理** ⭐⭐⭐
   - 无此功能题目无法判题
   - 工作量: 4-5小时

3. **密码重置** ⭐⭐
   - 影响用户体验
   - 工作量: 2-3小时

4. **管理员角色管理** ⭐⭐
   - 管理功能不完整
   - 工作量: 1小时

### P2 - 次要功能（可以延后）

5. **讨论区评论列表优化** ⭐
   - 工作量: 30分钟

6. **竞赛参与者列表**
   - 需要后端先实现

7. **头像上传**
   - 需要文件服务支持

---

## 5. 实施建议

### 短期（1周内）

**目标**: 解决P0和P1问题

```
Day 1-2: 竞赛题目绑定
  - API封装
  - ManageProblems.vue页面
  - Details.vue集成

Day 3-4: 测试用例管理
  - API封装
  - ManageTestcases.vue页面
  - 题目详情页集成

Day 5: 密码重置 + 角色管理
  - 密码重置流程
  - 角色管理API对接
```

### 中期（2周内）

**目标**: 完善次要功能

- 讨论区评论优化
- 竞赛参与者列表（需后端配合）
- 头像上传方案设计

---

## 6. 技术债务

### 6.1 API封装不一致

**问题**:
- 有些API返回占位符对象
- 有些API直接throw错误
- 错误处理不统一

**建议**:
```javascript
// 统一的占位符模式
const notImplemented = (feature) => {
    throw {
        response: {
            status: 501,
            data: {
                error: `${feature} is not available in rebuild backend`
            }
        }
    };
}
```

### 6.2 前端页面与API脱节

**问题**:
- `Account/Reset.vue` 显示"功能已禁用"，但后端已实现
- `Admin/Account.vue` 有角色管理UI，但API是占位符

**建议**:
- 定期审查前后端对齐情况
- 建立前后端API契约测试

---

## 7. 总结

### 当前状态

```
┌─────────────────────────────────────────────┐
│         前端功能完整度分析                    │
├─────────────────────────────────────────────┤
│                                             │
│  ✅ 已完整实现: 60%                          │
│  ├─ 认证登录                                 │
│  ├─ 题目浏览                                 │
│  ├─ 提交代码                                 │
│  ├─ 班级管理                                 │
│  ├─ 讨论区                                   │
│  └─ 用户状态管理                             │
│                                             │
│  🔴 缺失关键功能: 25%                        │
│  ├─ 竞赛题目绑定 ⭐⭐⭐                       │
│  ├─ 测试用例管理 ⭐⭐⭐                       │
│  ├─ 密码重置 ⭐⭐                            │
│  └─ 角色管理 ⭐⭐                            │
│                                             │
│  🟡 部分实现: 10%                            │
│  ├─ 头像上传                                 │
│  └─ 评论列表                                 │
│                                             │
│  ⚪ 未规划: 5%                               │
│  ├─ 退出竞赛                                 │
│  ├─ 删除讨论                                 │
│  └─ IP统计                                   │
│                                             │
└─────────────────────────────────────────────┘
```

### 关键发现

1. **Phase5后端已实现，但前端未对接**
   - 密码重置 ✅ 后端完成 ❌ 前端占位符
   - 角色管理 ✅ 后端完成 ❌ 前端占位符
   - 竞赛题目绑定 ✅ 后端完成 ❌ 前端占位符

2. **Phase4后端已实现，但前端未对接**
   - 测试用例管理 ✅ 后端完成 ❌ 前端完全缺失

3. **前端页面与实际能力不符**
   - Reset.vue显示"已禁用"，实际后端已支持
   - Admin/Account.vue有UI，实际API不可用

### 建议行动

**立即行动**（本周）:
1. 实现竞赛题目绑定（阻塞竞赛功能）
2. 实现测试用例管理（阻塞判题功能）

**短期行动**（2周内）:
3. 对接密码重置功能
4. 对接角色管理功能

**中期规划**（1个月内）:
5. 完善次要功能
6. 清理技术债务
7. 建立前后端契约测试

---

**报告版本**: v1.0  
**创建日期**: 2026-04-20  
**审查方法**: 逐文件对比 + 实际代码验证  
**下次审查**: 实现P0/P1功能后
