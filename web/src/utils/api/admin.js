import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

const reqConfig = () => ({
    headers: {
        ...authHeaders()
    }
});

// 管理员获取竞赛列表
export const getAllCompetitionsInfo = async () => {
    return await axios.get(`${apiUrl}/contests`, reqConfig())
}

// 管理员获取指定列表信息
export const getCompetitionInfoByIDAdmin = async (cid) => {
    return await axios.get(`${apiUrl}/contests/${cid}`, reqConfig())
}

// 管理员获取指定题目所有信息
export const getProblemAllInfoByAdmin = async (pid) => {
    return await axios.get(`${apiUrl}/problems/${pid}`, reqConfig())
}

// 管理员获取所有用户信息
export const getAllUsersInfo = async () => {
    return await axios.get(`${apiUrl}/admin/users`, {
        headers: {
            ...authHeaders()
        }
    })
}

// 管理员添加/更新题目信息
export const updateProblemInfo = async (problemInfo) => {
    const payload = {
        title: problemInfo.title,
        content: problemInfo.content,
        input: problemInfo.input,
        output: problemInfo.output,
        difficulty: Number(problemInfo.difficulty ?? 0),
        time_limit_ms: Number(problemInfo.time_limit_ms ?? 1000),
        memory_limit_mb: Number(problemInfo.memory_limit_mb ?? 128),
        class_id: problemInfo.class_id ? String(problemInfo.class_id) : "",
        visibility: problemInfo.visibility ?? 'public',
        status: problemInfo.status ?? 'published'
    }
    try {
        if (problemInfo.id) {
            return await axios.patch(`${apiUrl}/problems/${problemInfo.id}`, payload, reqConfig())
        }
    } catch (error) {
        if (error?.response?.status !== 404) {
            throw error
        }
    }
    return await axios.post(`${apiUrl}/problems`, payload, reqConfig())
}

// 管理员删除题目信息
export const deleteProblemAllInfo = async (pid) => {
    return await axios.delete(`${apiUrl}/problems/${pid}`, reqConfig())
}

// 封禁用户
export const banUser = async (uid) => {
    return await axios.patch(`${apiUrl}/admin/users/status`, {
        user_id: uid,
        status: 'banned'
    }, {
        headers: {
            ...authHeaders()
        }
    })
}

// 解封用户
export const unbanUser = async (uid) => {
    return await axios.patch(`${apiUrl}/admin/users/status`, {
        user_id: uid,
        status: 'active'
    }, {
        headers: {
            ...authHeaders()
        }
    })
}

// 晋升用户
export const promoteUser = async (uid) => {
    return {
        data: { message: 'admin role update is not available in rebuild backend' }
    }
}

// 降级用户
export const demoteUser = async (uid) => {
    return {
        data: { message: 'admin role update is not available in rebuild backend' }
    }
}

// 管理员获取题目列表
export const getAllProblemsAdmin = async () => {
    return await axios.get(`${apiUrl}/problems`, reqConfig())
}

// 管理员删除竞赛
export const deleteCompetition = async (cid) => {
    return await axios.delete(`${apiUrl}/contests/${cid}`, reqConfig())
}

// 管理员添加/更新竞赛信息
export const updateComInfo = async (comInfo) => {
    const payload = {
        title: comInfo.title,
        subtitle: comInfo.subtitle,
        description: comInfo.description ?? comInfo.subtitle ?? "",
        announcement: comInfo.announcement,
        class_id: comInfo.class_id ? String(comInfo.class_id) : "",
        visibility: comInfo.visibility ?? 'public',
        rule_type: comInfo.rule_type ?? 'acm',
        status: comInfo.status ?? 'scheduled',
        is_encrypted: Boolean(comInfo.is_encrypted),
        password: comInfo.is_encrypted ? (comInfo.password ?? "") : "",
        start_at: comInfo.start_at,
        end_at: comInfo.end_at
    }
    try {
        if (comInfo.id) {
            return await axios.patch(`${apiUrl}/contests/${comInfo.id}`, payload, reqConfig())
        }
    } catch (error) {
        if (error?.response?.status !== 404) {
            throw error
        }
    }
    return await axios.post(`${apiUrl}/contests`, payload, reqConfig())
}

// 管理员启用竞赛计分
export const caculateComScore = async (cid) => {
    return await getScores(cid, 1, 10)
}

// 管理员查看竞赛得分情况
export const getScores = async (cid, page, itemsPerPage) => {
    const resp = await axios.get(`${apiUrl}/contests/${cid}/scoreboard`, reqConfig())
    const rows = resp.data?.data?.items ?? []
    const start = Math.max((Number(page) - 1) * Number(itemsPerPage), 0)
    const end = start + Number(itemsPerPage)
    const users = rows.slice(start, end).map((item) => ({
        username: item.username,
        score: Number(item.total_score ?? item.solved ?? 0)
    }))
    return {
        ...resp,
        data: {
            users,
            total: rows.length
        }
    }
}

// 管理员获取IP统计
export const getIpStat = async () => {
    return {
        data: { data: [] }
    }
}
