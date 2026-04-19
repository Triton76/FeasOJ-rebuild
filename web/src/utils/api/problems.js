import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

// 获取题目列表
export const getAllProblems = async () => {
    return await axios.get(`${apiUrl}/problems`, {
        headers: {
            ...authHeaders()
        }
    });
}

// 获取每日一题
export const getDailyProblem = async () => {
    const resp = await axios.get(`${apiUrl}/problems`, {
        headers: {
            ...authHeaders()
        }
    });
    const list = resp.data?.data || [];
    return {
        ...resp,
        data: {
            ...resp.data,
            data: list.length > 0 ? list[0] : null
        }
    }
}

// 获取题目详细信息
export const getPbDetails = async (pid) => {
    return await axios.get(`${apiUrl}/problems/${pid}`, {
        headers: {
            ...authHeaders()
        }
    });
}

// 提交代码文件
export const uploadCode = async (file, pid) => {
    const sourceCode = await file.text();
    const resp = await axios.post(`${apiUrl}/submit-records`, {
        problem_id: Number(pid),
        contest_id: 0,
        language: 'cpp',
        source_code: sourceCode
    }, {
        headers: {
            ...authHeaders()
        },
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            message: resp.data?.message || 'success'
        }
    }
}