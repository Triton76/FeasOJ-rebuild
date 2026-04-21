import axios from 'axios';
import { apiUrl } from '../axios';
import { language, token } from '../account';

const authHeaders = () => ({
    Authorization: token.value ? `Bearer ${token.value}` : '',
    "Accept-Language": language.value
});

const reqConfig = () => ({
    headers: {
        ...authHeaders()
    }
});

export const getProblemTestcases = async (problemId) => {
    return await axios.get(`${apiUrl}/problems/${problemId}/testcases`, reqConfig());
};

export const createProblemTestcase = async (problemId, payload) => {
    return await axios.post(`${apiUrl}/problems/${problemId}/testcases`, payload, reqConfig());
};

export const updateProblemTestcase = async (problemId, testcaseId, payload) => {
    return await axios.patch(`${apiUrl}/problems/${problemId}/testcases/${testcaseId}`, payload, reqConfig());
};

export const deleteProblemTestcase = async (problemId, testcaseId) => {
    return await axios.delete(`${apiUrl}/problems/${problemId}/testcases/${testcaseId}`, reqConfig());
};

export const reorderProblemTestcases = async (problemId, testcaseIds) => {
    return await axios.post(`${apiUrl}/problems/${problemId}/testcases/reorder`, {
        testcase_ids: testcaseIds
    }, reqConfig());
};
