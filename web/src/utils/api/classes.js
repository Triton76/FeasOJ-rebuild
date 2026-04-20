import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

export const createClass = async (payload) => {
    return await axios.post(`${apiUrl}/classes`, payload, {
        headers: {
            ...authHeaders()
        }
    });
}

export const updateClass = async (classId, payload) => {
    return await axios.patch(`${apiUrl}/classes/${classId}`, payload, {
        headers: {
            ...authHeaders()
        }
    });
}

export const archiveClass = async (classId) => {
    return await axios.post(`${apiUrl}/classes/${classId}/archive`, {}, {
        headers: {
            ...authHeaders()
        }
    });
}

export const applyJoinClass = async (classCode) => {
    return await axios.post(`${apiUrl}/classes/join`, {
        class_code: classCode
    }, {
        headers: {
            ...authHeaders()
        }
    });
}

export const reviewMembership = async (membershipId, approve) => {
    return await axios.post(`${apiUrl}/classes/memberships/review`, {
        membership_id: membershipId,
        approve: !!approve
    }, {
        headers: {
            ...authHeaders()
        }
    });
}

export const listClassMemberships = async (classId) => {
    return await axios.get(`${apiUrl}/classes/${classId}/memberships`, {
        headers: {
            ...authHeaders()
        }
    });
}

export const listMyMemberships = async () => {
    return await axios.get(`${apiUrl}/classes/memberships/self`, {
        headers: {
            ...authHeaders()
        }
    });
}
