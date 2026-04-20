import axios from 'axios';
import { apiUrl } from '../axios';
import { token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : ''
});

// 获取指定用户提交记录
export const getUserSubmitRecords = async (targetUserId) => {
    const params = {}
    if (targetUserId) {
        params.user_id = targetUserId
    }
    return await axios.get(`${apiUrl}/submit-records`,{
        params,
        headers: {
            ...authHeaders()
        }
    })
}

// 获取30天内的提交记录
export const getSubmitRecords = async () => {
    return await axios.get(`${apiUrl}/submit-records`, {
        headers: {
            ...authHeaders()
        }
    })
}