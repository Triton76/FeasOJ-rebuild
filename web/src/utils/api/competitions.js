import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

const joinedKey = 'joined_contests';

const getJoinedContests = () => {
    const raw = localStorage.getItem(joinedKey);
    if (!raw) return [];
    try {
        const arr = JSON.parse(raw);
        return Array.isArray(arr) ? arr : [];
    } catch {
        return [];
    }
}

const setJoinedContests = (list) => {
    localStorage.setItem(joinedKey, JSON.stringify(Array.from(new Set(list))));
}

const markJoined = (contestId) => {
    const list = getJoinedContests();
    list.push(Number(contestId));
    setJoinedContests(list);
}

const markQuit = (contestId) => {
    const target = Number(contestId);
    const list = getJoinedContests().filter((id) => id !== target);
    setJoinedContests(list);
}

const mapContestStatus = (status) => {
    // 兼容旧前端的 0/1/2 状态显示
    switch (status) {
        case 'scheduled':
            return 0;
        case 'running':
            return 1;
        case 'ended':
            return 2;
        default:
            return 0;
    }
}

const normalizeContest = (c) => ({
    ...c,
    encrypted: !!c?.is_encrypted,
    status: mapContestStatus(c?.status),
    difficulty: c?.difficulty ?? 0,
});

// 获取竞赛列表
export const getAllCompetitions = async () => {
    const resp = await axios.get(`${apiUrl}/contests`, {
        headers: {
            ...authHeaders(),
        }
    });
    const list = (resp.data?.data || []).map(normalizeContest);
    return {
        ...resp,
        data: {
            ...resp.data,
            data: list
        }
    }
}

// 获取指定竞赛ID信息
export const getCompetitionById = async (competitionId) => {
    const resp = await axios.get(`${apiUrl}/contests/${competitionId}`, {
        headers: {
            ...authHeaders(),
        }
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            data: normalizeContest(resp.data?.data || {})
        }
    }
}

// 加入竞赛
export const joinCompetition = async (competitionId) => {
    const resp = await axios.post(`${apiUrl}/contests/join`, {
        contest_id: Number(competitionId)
    }, {
        headers: {
            ...authHeaders()
        }
    });
    markJoined(competitionId);
    return {
        ...resp,
        data: {
            ...resp.data,
            message: resp.data?.message || 'success'
        }
    }
}

// 加入有密码的竞赛
export const joinCompWithPwd = async (competitionId, competitionPwd) => {
    const resp = await axios.post(`${apiUrl}/contests/join`, {
        contest_id: Number(competitionId),
        password: competitionPwd
    }, {
        headers: {
            ...authHeaders()
        }
    });
    markJoined(competitionId);
    return {
        ...resp,
        data: {
            ...resp.data,
            message: resp.data?.message || 'success'
        }
    }
}

// 退出竞赛
export const quitCompetition = async (competitionId) => {
    markQuit(competitionId);
    return {
        data: {
            message: 'success'
        }
    }
}

// 查询用户是否在指定竞赛中
export const isInCompetition = async (competitionId) => {
    const inLocal = getJoinedContests().includes(Number(competitionId));
    return {
        data: {
            isIn: inLocal
        }
    }
}

// 获取指定竞赛的所有用户
export const getCompetitionUsers = async (competitionId) => {
    return {
        data: {
            data: []
        }
    }
}

// 获取指定竞赛的所有题目
export const getCompetitionProblems = async (competitionId) => {
    return {
        data: {
            data: []
        }
    }
}