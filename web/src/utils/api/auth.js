import axios from 'axios';
import { apiUrl } from '../axios';
import { language, userId, userName } from '../account';

const buildAuthHeader = (rawToken) => {
    const t = (rawToken || '').trim();
    if (!t) return '';
    return t.startsWith('Bearer ') ? t : `Bearer ${t}`;
}

// 注册
export const registerRequest = async (username, password, email, vcode) => {
    const resp = await axios.post(`${apiUrl}/auth/register`, {
        email: email,
        username: username,
        password: password,
    },{
        headers: {
            "Accept-Language": language.value
        }
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            message: resp.data?.message || 'success'
        }
    }
}

// 登录
export const loginRequest = async (username, password) => {
    const resp = await axios.post(`${apiUrl}/auth/login`, {
            username: username,
            password: password
        }, {
        headers: {
            "Accept-Language": language.value,
        }
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            token: resp.data?.data?.token || '',
            message: resp.data?.message || 'success'
        }
    }
}

// 获取验证码
export const getCaptchaCode = async (email,iscreate) => {
    // Rebuild 后端当前阶段不提供验证码接口，保留函数签名避免页面崩溃
    return {
        data: {
            message: 'captcha endpoint is not available in rebuild backend'
        }
    }
}

// 验证个人用户信息
export const verifyUserInfo = async (username, token) => {
    const resp = await axios.get(`${apiUrl}/auth/verify`, {
        headers: {
            Authorization: buildAuthHeader(token),
            "Accept-Language": language.value
        }
    });
    return {
        ...resp,
        data: {
            ...resp.data,
            data: resp.data?.data?.user || resp.data?.data || {},
            message: resp.data?.message || 'success'
        }
    }
}

// 获取用户信息
export const getUserInfo = async (username) => {
    const target = username === userName.value ? userId.value : username;
    return await axios.get(`${apiUrl}/users/${target}`, {
        headers: {
            Authorization: buildAuthHeader(localStorage.getItem('token') || ''),
            "Accept-Language": language.value
        }
    });
}

// 修改密码
export const updatePassword = async (email, vcode, newPassword) => {
    // 决议已下线重置密码能力，这里返回可读提示
    return {
        data: {
            message: 'reset password is disabled in rebuild backend'
        }
    }
}
