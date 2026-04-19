import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

// 管理员获取竞赛列表
export const getAllCompetitionsInfo = async () => {
    return {
        data: { data: [] }
    }
}

// 管理员获取指定列表信息
export const getCompetitionInfoByIDAdmin = async (cid) => {
    return {
        data: { data: null }
    }
}

// 管理员获取指定题目所有信息
export const getProblemAllInfoByAdmin = async (pid) => {
    return {
        data: { data: null }
    }
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
    return {
        data: { message: 'admin problem write is not available in rebuild backend' }
    }
}

// 管理员删除题目信息
export const deleteProblemAllInfo = async (pid) => {
    return {
        data: { message: 'admin problem delete is not available in rebuild backend' }
    }
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
    return {
        data: { data: [] }
    }
}

// 管理员删除竞赛
export const deleteCompetition = async (cid) => {
    return {
        data: { message: 'admin contest delete is not available in rebuild backend' }
    }
}

// 管理员添加/更新竞赛信息
export const updateComInfo = async (comInfo) => {
    return {
        data: { message: 'admin contest write is not available in rebuild backend' }
    }
}

// 管理员启用竞赛计分
export const caculateComScore = async (cid) => {
    return {
        data: { message: 'admin score calculation is not available in rebuild backend' }
    }
}

// 管理员查看竞赛得分情况
export const getScores = async (cid, page, itemsPerPage) => {
    return {
        data: { data: [] }
    }
}

// 管理员获取IP统计
export const getIpStat = async () => {
    return {
        data: { data: [] }
    }
}
