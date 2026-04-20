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
    throw {
        response: {
            status: 501,
            data: {
                error: 'contest quit endpoint is not available in rebuild backend'
            }
        }
    };
}

// 查询用户是否在指定竞赛中
export const isInCompetition = async (competitionId) => {
    throw {
        response: {
            status: 501,
            data: {
                error: 'membership lookup endpoint is not available in rebuild backend'
            }
        }
    };
}

// 获取指定竞赛的所有用户
export const getCompetitionUsers = async (competitionId) => {
    throw {
        response: {
            status: 501,
            data: {
                error: 'contest participants endpoint is not available in rebuild backend'
            }
        }
    };
}

// 获取指定竞赛的所有题目
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