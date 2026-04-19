import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

// 获取讨论列表
export const getAllDis = async (page, itemsPerPage) => {
    return await axios.get(`${apiUrl}/discussions`, {
        params: {
            page: page,
            limit: itemsPerPage
        },
        headers: {
            ...authHeaders()
        }
    })
}

// 获取讨论详细信息
export const getDisDetails = async (did) => {
    return await axios.get(`${apiUrl}/discussions/${did}`, {
        headers: {
            ...authHeaders()
        }
    })
}

// 获取指定讨论的所有回复
export const getComments = async (did) => {
    // rebuild 当前阶段无独立 comments 列表接口
    return {
        data: {
            data: []
        }
    }
}

// 添加讨论
export const addDiscussion = async (Title, Content) => {
    const resp = await axios.post(`${apiUrl}/discussions`, {
        title: Title,
        content: Content
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

// 添加评论
export const addComment = async (did, content) => {
    const resp = await axios.post(`${apiUrl}/comments`, {
        discussion_id: did,
        content: content
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

// 删除讨论
export const deleteDiscussion = async (id) => {
    return {
        data: {
            message: 'delete discussion is not available in rebuild backend'
        }
    }
}

// 删除讨论评论
export const deleteComment = async (id) => {
    return {
        data: {
            message: 'delete comment is not available in rebuild backend'
        }
    }
}
