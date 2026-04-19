import axios from 'axios';
import { apiUrl, docsServer } from '../axios';
import { language, userId, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

// 获取公告
export const getAnnouncement = async () => {
    const resp = await fetch(`${docsServer}announcement.md`);
    return resp.text();
}

// 获取通知
export const getNotification = async () => {
    const resp = await fetch(`${docsServer}notice.md`);
    return resp.text();
}

// 更新用户简介
export const updateSynopsis = async (synopsis) => {
    const resp = await axios.patch(`${apiUrl}/profile`, {
        synopsis: synopsis
    }, {
        headers: authHeaders()
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            message: resp.data?.message || 'success'
        }
    }
}

// 修改头像
export const uploadAvatar = async (file) => {
    // Rebuild 后端当前阶段 UpdateProfile 仅支持 avatar 字符串字段，不支持文件上传。
    // 先传空字符串，避免页面调用时报错；后续可扩展独立上传能力。
    const resp = await axios.patch(`${apiUrl}/profile`, {
        avatar: ''
    }, {
        headers: authHeaders()
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            message: resp.data?.message || 'success'
        }
    }
}

// 获取排行榜
export const getRanking = async () => {
    return await axios.get(`${apiUrl}/ranking`, {
        headers: {
            ...authHeaders()
        }
    })
}