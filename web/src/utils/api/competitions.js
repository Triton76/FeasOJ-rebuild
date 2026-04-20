import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

// 获取竞赛列表
export const getAllCompetitions = async () => {
    return await axios.get(`${apiUrl}/contests`, {
        headers: {
            ...authHeaders(),
        }
    });
}

// 获取指定竞赛ID信息
export const getCompetitionById = async (competitionId) => {
    return await axios.get(`${apiUrl}/contests/${competitionId}`, {
        headers: {
            ...authHeaders(),
        }
    });
}

// 加入竞赛
export const joinCompetition = async (competitionId) => {
    return await axios.post(`${apiUrl}/contests/join`, {
        contest_id: Number(competitionId)
    }, {
        headers: {
            ...authHeaders()
        }
    });
}

// 加入有密码的竞赛
export const joinCompWithPwd = async (competitionId, competitionPwd) => {
    return await axios.post(`${apiUrl}/contests/join`, {
        contest_id: Number(competitionId),
        password: competitionPwd
    }, {
        headers: {
            ...authHeaders()
        }
    });
}

// 退出竞赛
export const quitCompetition = async (competitionId) => {
    return await axios.delete(`${apiUrl}/contests/${competitionId}/participant/self`, {
        headers: {
            ...authHeaders()
        }
    });
}

// 查询用户是否在指定竞赛中
export const isInCompetition = async (competitionId) => {
    return await axios.get(`${apiUrl}/contests/${competitionId}/participant/self`, {
        headers: {
            ...authHeaders()
        }
    });
}

// 获取指定竞赛的所有用户
export const getCompetitionUsers = async (competitionId) => {
    return await axios.get(`${apiUrl}/contests/${competitionId}/participants`, {
        headers: {
            ...authHeaders()
        }
    });
}

// 获取指定竞赛的所有题目
export const getCompetitionProblems = async (competitionId) => {
    return await axios.get(`${apiUrl}/contests/${competitionId}/problems`, {
        headers: {
            ...authHeaders()
        }
    });
}

// 管理员替换竞赛题目绑定
export const replaceCompetitionProblems = async (competitionId, items) => {
    return await axios.put(`${apiUrl}/contests/${competitionId}/problems`, {
        items
    }, {
        headers: {
            ...authHeaders()
        }
    });
}